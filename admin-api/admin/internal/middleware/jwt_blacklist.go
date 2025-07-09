package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/heimdall-api/common/cache"
	"github.com/heimdall-api/common/utils"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// JWTBlacklistMiddleware JWT黑名单检查中间件
type JWTBlacklistMiddleware struct {
	redis     *redis.Client
	jwtSecret string
	issuer    string
}

// NewJWTBlacklistMiddleware 创建JWT黑名单中间件
func NewJWTBlacklistMiddleware(redis *redis.Client, jwtSecret, issuer string) *JWTBlacklistMiddleware {
	return &JWTBlacklistMiddleware{
		redis:     redis,
		jwtSecret: jwtSecret,
		issuer:    issuer,
	}
}

// Handle JWT黑名单检查处理函数
func (m *JWTBlacklistMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 提取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// 没有Authorization头，让后续的JWT中间件处理
			next(w, r)
			return
		}

		// 2. 解析Bearer token
		token, err := m.extractToken(authHeader)
		if err != nil {
			// token格式错误，让后续的JWT中间件处理
			next(w, r)
			return
		}

		// 3. 检查token是否在黑名单中
		isBlacklisted, err := m.checkBlacklist(r.Context(), token)
		if err != nil {
			logx.Errorf("检查JWT黑名单失败: %v", err)
			// 检查失败，为了安全起见拒绝访问
			m.writeErrorResponse(w, "系统错误，请稍后重试")
			return
		}

		if isBlacklisted {
			logx.Infof("检测到黑名单Token访问: %s", m.maskToken(token))
			m.writeErrorResponse(w, "Token已失效，请重新登录")
			return
		}

		// 4. token不在黑名单中，继续处理
		next(w, r)
	}
}

// extractToken 从Authorization头中提取token
func (m *JWTBlacklistMiddleware) extractToken(authHeader string) (string, error) {
	return utils.ParseAuthHeader(authHeader)
}

// checkBlacklist 检查token是否在黑名单中
func (m *JWTBlacklistMiddleware) checkBlacklist(ctx context.Context, token string) (bool, error) {
	// 1. 解析token获取token ID
	jwtManager := utils.NewJWTManager(m.jwtSecret, m.issuer)
	tokenID, err := jwtManager.ExtractTokenIDFromToken(token)
	if err != nil {
		// 无法解析token ID，可能是无效token，让JWT中间件处理
		return false, nil
	}

	// 2. 检查黑名单缓存
	blacklistCache := cache.NewJWTBlacklistCache(m.redis, "jwt:blacklist:")
	return blacklistCache.IsTokenBlacklisted(ctx, tokenID)
}

// maskToken 对token进行脱敏处理用于日志
func (m *JWTBlacklistMiddleware) maskToken(token string) string {
	if len(token) <= 20 {
		return "***"
	}
	return token[:10] + "***" + token[len(token)-10:]
}

// writeErrorResponse 写入错误响应
func (m *JWTBlacklistMiddleware) writeErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"code":    401,
		"message": message,
	}

	httpx.WriteJson(w, http.StatusUnauthorized, response)
}

// SkipPaths 定义不需要检查JWT黑名单的路径
func (m *JWTBlacklistMiddleware) SkipPaths() []string {
	return []string{
		"/api/v1/admin/auth/login", // 登录接口
		"/api/v1/admin/ping",       // 健康检查
		"/api/v1/admin/health",     // 健康检查
	}
}

// ShouldSkip 判断当前路径是否应该跳过检查
func (m *JWTBlacklistMiddleware) ShouldSkip(path string) bool {
	skipPaths := m.SkipPaths()
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// HandleWithSkip 带路径跳过功能的处理函数
func (m *JWTBlacklistMiddleware) HandleWithSkip(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否应该跳过
		if m.ShouldSkip(r.URL.Path) {
			next(w, r)
			return
		}

		// 执行黑名单检查
		m.Handle(next)(w, r)
	}
}
