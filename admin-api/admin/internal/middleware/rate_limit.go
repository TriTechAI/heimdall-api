package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heimdall-api/common/cache"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 通用限流配置
	GeneralRPS   int           // 普通请求每秒限制
	GeneralBurst int           // 普通请求突发限制
	Window       time.Duration // 时间窗口

	// 登录接口特殊限流
	LoginRPS    int           // 登录请求每秒限制
	LoginBurst  int           // 登录请求突发限制
	LoginWindow time.Duration // 登录时间窗口

	// API创建操作限流
	CreateRPS    int           // 创建操作每秒限制
	CreateBurst  int           // 创建操作突发限制
	CreateWindow time.Duration // 创建操作时间窗口
}

// IPRateLimitMiddleware IP限流中间件
type IPRateLimitMiddleware struct {
	redis  *redis.Client
	config RateLimitConfig
	cache  *cache.RateLimiterCache
}

// NewIPRateLimitMiddleware 创建IP限流中间件
func NewIPRateLimitMiddleware(redis *redis.Client, config RateLimitConfig) *IPRateLimitMiddleware {
	if config.GeneralRPS == 0 {
		config.GeneralRPS = 100 // 默认每秒100请求
	}
	if config.GeneralBurst == 0 {
		config.GeneralBurst = 200 // 默认突发200请求
	}
	if config.Window == 0 {
		config.Window = time.Minute // 默认1分钟窗口
	}

	// 登录接口默认配置
	if config.LoginRPS == 0 {
		config.LoginRPS = 5 // 每秒5次登录尝试
	}
	if config.LoginBurst == 0 {
		config.LoginBurst = 10 // 突发10次
	}
	if config.LoginWindow == 0 {
		config.LoginWindow = 5 * time.Minute // 5分钟窗口
	}

	// 创建操作默认配置
	if config.CreateRPS == 0 {
		config.CreateRPS = 10 // 每秒10次创建操作
	}
	if config.CreateBurst == 0 {
		config.CreateBurst = 20 // 突发20次
	}
	if config.CreateWindow == 0 {
		config.CreateWindow = time.Minute // 1分钟窗口
	}

	rateLimiterCache := cache.NewRateLimiterCache(redis, "rate_limit:")

	return &IPRateLimitMiddleware{
		redis:  redis,
		config: config,
		cache:  rateLimiterCache,
	}
}

// Handle IP限流处理函数
func (m *IPRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取客户端IP
		clientIP := m.getClientIP(r)
		if clientIP == "" {
			logx.Error("无法获取客户端IP地址")
			m.writeErrorResponse(w, "请求异常", http.StatusBadRequest)
			return
		}

		// 2. 根据请求路径选择限流策略
		limitConfig := m.selectLimitConfig(r.URL.Path, r.Method)

		// 3. 执行限流检查
		allowed, remaining, resetTime, err := m.checkRateLimit(r.Context(), clientIP, r.URL.Path, limitConfig)
		if err != nil {
			logx.Errorf("限流检查失败: clientIP=%s, error=%v", clientIP, err)
			// 限流检查失败，为了安全起见拒绝访问
			m.writeErrorResponse(w, "系统错误，请稍后重试", http.StatusInternalServerError)
			return
		}

		// 4. 设置限流响应头
		m.setRateLimitHeaders(w, limitConfig.Limit, remaining, resetTime)

		// 5. 检查是否超出限流
		if !allowed {
			logx.Infof("IP限流触发: clientIP=%s, path=%s, method=%s", clientIP, r.URL.Path, r.Method)
			m.writeRateLimitResponse(w, remaining, resetTime)
			return
		}

		// 6. 未超出限流，继续处理
		next(w, r)
	}
}

// getClientIP 获取客户端真实IP
func (m *IPRateLimitMiddleware) getClientIP(r *http.Request) string {
	// 1. 检查X-Forwarded-For头（代理/负载均衡器）
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// 取第一个IP（客户端真实IP）
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if m.isValidIP(ip) {
				return ip
			}
		}
	}

	// 2. 检查X-Real-IP头（Nginx代理）
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" && m.isValidIP(realIP) {
		return realIP
	}

	// 3. 检查X-Client-IP头
	clientIP := r.Header.Get("X-Client-IP")
	if clientIP != "" && m.isValidIP(clientIP) {
		return clientIP
	}

	// 4. 使用RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isValidIP 验证IP地址有效性
func (m *IPRateLimitMiddleware) isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	// 排除本地回环地址
	return !parsedIP.IsLoopback() && !parsedIP.IsUnspecified()
}

// selectLimitConfig 根据路径和方法选择限流配置
func (m *IPRateLimitMiddleware) selectLimitConfig(path, method string) cache.RateLimitConfig {
	// 登录接口特殊限流
	if strings.Contains(path, "/auth/login") {
		return cache.RateLimitConfig{
			Limit:  m.config.LoginRPS,
			Window: m.config.LoginWindow,
			Burst:  m.config.LoginBurst,
		}
	}

	// POST创建操作限流
	if method == "POST" && (strings.Contains(path, "/posts") ||
		strings.Contains(path, "/users") ||
		strings.Contains(path, "/pages")) {
		return cache.RateLimitConfig{
			Limit:  m.config.CreateRPS,
			Window: m.config.CreateWindow,
			Burst:  m.config.CreateBurst,
		}
	}

	// 通用限流配置
	return cache.RateLimitConfig{
		Limit:  m.config.GeneralRPS,
		Window: m.config.Window,
		Burst:  m.config.GeneralBurst,
	}
}

// checkRateLimit 执行限流检查
func (m *IPRateLimitMiddleware) checkRateLimit(ctx context.Context, clientIP, path string, config cache.RateLimitConfig) (allowed bool, remaining int, resetTime int64, err error) {
	// 构建限流键
	key := fmt.Sprintf("%s:%s", clientIP, m.normalizePathForRateLimit(path))

	// 执行限流检查
	var result *cache.RateLimitResult
	if strings.Contains(path, "/auth/login") {
		// 登录使用滑动窗口
		result, err = m.cache.CheckRateLimit(ctx, key, config)
	} else {
		// 其他使用固定窗口
		result, err = m.cache.CheckRateLimitFixed(ctx, key, config)
	}

	if err != nil {
		return false, 0, 0, err
	}

	return result.Allowed, result.Remaining, result.ResetAt.Unix(), nil
}

// normalizePathForRateLimit 标准化路径用于限流
func (m *IPRateLimitMiddleware) normalizePathForRateLimit(path string) string {
	// 将路径中的ID参数标准化，避免每个不同ID都创建新的限流键
	// 例如：/api/v1/admin/users/123 -> /api/v1/admin/users/{id}

	parts := strings.Split(path, "/")
	for i, part := range parts {
		// 检查是否为ID参数（纯数字或类似ObjectID格式）
		if m.looksLikeID(part) {
			parts[i] = "{id}"
		}
	}

	return strings.Join(parts, "/")
}

// looksLikeID 判断字符串是否像ID
func (m *IPRateLimitMiddleware) looksLikeID(s string) bool {
	// 检查是否为纯数字
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}

	// 检查是否为MongoDB ObjectID格式（24位十六进制）
	if len(s) == 24 {
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	return false
}

// setRateLimitHeaders 设置限流响应头
func (m *IPRateLimitMiddleware) setRateLimitHeaders(w http.ResponseWriter, limit, remaining int, resetTime int64) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
}

// writeErrorResponse 写入错误响应
func (m *IPRateLimitMiddleware) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"code":    statusCode,
		"message": message,
	}

	httpx.WriteJson(w, statusCode, response)
}

// writeRateLimitResponse 写入限流响应
func (m *IPRateLimitMiddleware) writeRateLimitResponse(w http.ResponseWriter, remaining int, resetTime int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"code":      429,
		"message":   "请求过于频繁，请稍后重试",
		"remaining": remaining,
		"resetTime": resetTime,
	}

	httpx.WriteJson(w, http.StatusTooManyRequests, response)
}
