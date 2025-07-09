package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// AuditAction 审计操作类型
type AuditAction string

const (
	AuditActionLogin         AuditAction = "login"
	AuditActionLogout        AuditAction = "logout"
	AuditActionCreateUser    AuditAction = "create_user"
	AuditActionUpdateUser    AuditAction = "update_user"
	AuditActionDeleteUser    AuditAction = "delete_user"
	AuditActionLockUser      AuditAction = "lock_user"
	AuditActionUnlockUser    AuditAction = "unlock_user"
	AuditActionResetPassword AuditAction = "reset_password"
	AuditActionCreatePost    AuditAction = "create_post"
	AuditActionUpdatePost    AuditAction = "update_post"
	AuditActionDeletePost    AuditAction = "delete_post"
	AuditActionPublishPost   AuditAction = "publish_post"
	AuditActionCreatePage    AuditAction = "create_page"
	AuditActionUpdatePage    AuditAction = "update_page"
	AuditActionDeletePage    AuditAction = "delete_page"
	AuditActionUnknown       AuditAction = "unknown"
)

// AuditLog 审计日志结构
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"userId,omitempty"`
	Username     string                 `json:"username,omitempty"`
	Action       AuditAction            `json:"action"`
	Resource     string                 `json:"resource,omitempty"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	ClientIP     string                 `json:"clientIP"`
	UserAgent    string                 `json:"userAgent"`
	RequestBody  map[string]interface{} `json:"requestBody,omitempty"`
	ResponseCode int                    `json:"responseCode"`
	Duration     int64                  `json:"duration"` // 毫秒
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	SessionID    string                 `json:"sessionId,omitempty"`
	TraceID      string                 `json:"traceId,omitempty"`
}

// responseWriter 包装ResponseWriter以捕获响应信息
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// 捕获响应体用于审计
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// AuditMiddleware 操作审计中间件
type AuditMiddleware struct {
	redis  *redis.Client
	logger logx.Logger
}

// NewAuditMiddleware 创建操作审计中间件
func NewAuditMiddleware(redis *redis.Client) *AuditMiddleware {
	return &AuditMiddleware{
		redis:  redis,
		logger: logx.WithContext(context.Background()),
	}
}

// Handle 审计处理函数
func (m *AuditMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查是否需要审计
		if !m.shouldAudit(r.URL.Path, r.Method) {
			next(w, r)
			return
		}

		startTime := time.Now()

		// 1. 提取请求信息
		auditLog := m.extractRequestInfo(r)

		// 2. 读取请求体（用于审计）
		requestBody := m.readRequestBody(r)
		if requestBody != nil {
			auditLog.RequestBody = requestBody
		}

		// 3. 包装ResponseWriter以捕获响应
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // 默认200
		}

		// 4. 执行请求处理
		next(rw, r)

		// 5. 完善审计日志信息
		auditLog.ResponseCode = rw.statusCode
		auditLog.Duration = time.Since(startTime).Milliseconds()
		auditLog.Success = rw.statusCode < 400
		auditLog.Timestamp = startTime

		// 6. 提取错误信息（如果有）
		if !auditLog.Success {
			auditLog.ErrorMessage = m.extractErrorMessage(rw.body.Bytes())
		}

		// 7. 记录审计日志
		m.logAudit(r.Context(), auditLog)
	}
}

// shouldAudit 判断是否需要审计
func (m *AuditMiddleware) shouldAudit(path, method string) bool {
	// 审计所有admin API的重要操作
	if !strings.HasPrefix(path, "/api/v1/admin") {
		return false
	}

	// 跳过某些不需要审计的路径
	skipPaths := []string{
		"/api/v1/admin/ping",
		"/api/v1/admin/health",
		"/api/v1/admin/auth/profile", // GET profile查询不需要审计
	}

	for _, skipPath := range skipPaths {
		if path == skipPath && method == "GET" {
			return false
		}
	}

	// 审计所有写操作（POST, PUT, DELETE）
	writeMethods := []string{"POST", "PUT", "DELETE"}
	for _, writeMethod := range writeMethods {
		if method == writeMethod {
			return true
		}
	}

	// 审计重要的读操作
	importantReadPaths := []string{
		"/api/v1/admin/users",               // 用户列表查询
		"/api/v1/admin/security/login-logs", // 登录日志查询
	}

	if method == "GET" {
		for _, importantPath := range importantReadPaths {
			if strings.HasPrefix(path, importantPath) {
				return true
			}
		}
	}

	return false
}

// extractRequestInfo 提取请求基本信息
func (m *AuditMiddleware) extractRequestInfo(r *http.Request) *AuditLog {
	// 生成审计日志ID
	auditID := m.generateAuditID()

	auditLog := &AuditLog{
		ID:        auditID,
		Method:    r.Method,
		Path:      r.URL.Path,
		ClientIP:  m.getClientIP(r),
		UserAgent: r.UserAgent(),
		TraceID:   r.Header.Get("X-Trace-ID"),
	}

	// 提取用户信息（如果已认证）
	if userID := r.Context().Value("uid"); userID != nil {
		if uid, ok := userID.(string); ok {
			auditLog.UserID = uid
		}
	}

	if username := r.Context().Value("username"); username != nil {
		if uname, ok := username.(string); ok {
			auditLog.Username = uname
		}
	}

	// 确定操作类型和资源
	auditLog.Action, auditLog.Resource, auditLog.ResourceID = m.parseActionAndResource(r.URL.Path, r.Method)

	return auditLog
}

// getClientIP 获取客户端IP（复用rate_limit中的逻辑）
func (m *AuditMiddleware) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = strings.Cut(ip, ":")
	}
	return ip
}

// parseActionAndResource 解析操作类型和资源
func (m *AuditMiddleware) parseActionAndResource(path, method string) (AuditAction, string, string) {
	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	// 认证相关操作
	if strings.Contains(path, "/auth/") {
		if strings.HasSuffix(path, "/login") {
			return AuditActionLogin, "auth", ""
		}
		if strings.HasSuffix(path, "/logout") {
			return AuditActionLogout, "auth", ""
		}
	}

	// 用户管理操作
	if strings.Contains(path, "/users") {
		resourceID := m.extractResourceID(parts, "users")
		switch method {
		case "POST":
			if strings.HasSuffix(path, "/lock") {
				return AuditActionLockUser, "user", resourceID
			}
			if strings.HasSuffix(path, "/unlock") {
				return AuditActionUnlockUser, "user", resourceID
			}
			if strings.HasSuffix(path, "/reset-password") {
				return AuditActionResetPassword, "user", resourceID
			}
			return AuditActionCreateUser, "user", ""
		case "PUT":
			return AuditActionUpdateUser, "user", resourceID
		case "DELETE":
			return AuditActionDeleteUser, "user", resourceID
		}
	}

	// 文章管理操作
	if strings.Contains(path, "/posts") {
		resourceID := m.extractResourceID(parts, "posts")
		switch method {
		case "POST":
			if strings.HasSuffix(path, "/publish") {
				return AuditActionPublishPost, "post", resourceID
			}
			return AuditActionCreatePost, "post", ""
		case "PUT":
			return AuditActionUpdatePost, "post", resourceID
		case "DELETE":
			return AuditActionDeletePost, "post", resourceID
		}
	}

	// 页面管理操作
	if strings.Contains(path, "/pages") {
		resourceID := m.extractResourceID(parts, "pages")
		switch method {
		case "POST":
			return AuditActionCreatePage, "page", ""
		case "PUT":
			return AuditActionUpdatePage, "page", resourceID
		case "DELETE":
			return AuditActionDeletePage, "page", resourceID
		}
	}

	return AuditActionUnknown, "", ""
}

// extractResourceID 从路径中提取资源ID
func (m *AuditMiddleware) extractResourceID(parts []string, resourceType string) string {
	for i, part := range parts {
		if part == resourceType && i+1 < len(parts) {
			// 下一个part可能是ID
			nextPart := parts[i+1]
			// 简单检查是否为ID格式
			if len(nextPart) > 0 && !strings.Contains(nextPart, "/") {
				return nextPart
			}
		}
	}
	return ""
}

// readRequestBody 读取请求体用于审计
func (m *AuditMiddleware) readRequestBody(r *http.Request) map[string]interface{} {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}

	// 读取body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}

	// 重置body以供后续使用
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 只审计JSON请求体
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return nil
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		return nil
	}

	// 过滤敏感字段
	return m.filterSensitiveFields(bodyMap)
}

// filterSensitiveFields 过滤敏感字段
func (m *AuditMiddleware) filterSensitiveFields(data map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{
		"password", "currentPassword", "newPassword", "confirmPassword",
		"token", "accessToken", "refreshToken",
		"secret", "key", "apiKey",
	}

	result := make(map[string]interface{})
	for k, v := range data {
		// 检查是否为敏感字段
		isSensitive := false
		lowerKey := strings.ToLower(k)
		for _, sensitiveField := range sensitiveFields {
			if strings.Contains(lowerKey, sensitiveField) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			result[k] = "***"
		} else {
			result[k] = v
		}
	}

	return result
}

// extractErrorMessage 从响应体中提取错误信息
func (m *AuditMiddleware) extractErrorMessage(responseBody []byte) string {
	if len(responseBody) == 0 {
		return ""
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return ""
	}

	if message, exists := response["message"]; exists {
		if msg, ok := message.(string); ok {
			return msg
		}
	}

	return ""
}

// generateAuditID 生成审计日志ID
func (m *AuditMiddleware) generateAuditID() string {
	return "audit_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// logAudit 记录审计日志
func (m *AuditMiddleware) logAudit(ctx context.Context, auditLog *AuditLog) {
	// 1. 记录到应用日志
	m.logger.WithContext(ctx).Infof("AUDIT: %s %s %s user=%s action=%s resource=%s status=%d duration=%dms",
		auditLog.Method, auditLog.Path, auditLog.ClientIP,
		auditLog.Username, auditLog.Action, auditLog.Resource,
		auditLog.ResponseCode, auditLog.Duration)

	// 2. 异步存储到Redis（用于审计日志查询）
	go func() {
		if err := m.storeAuditLog(context.Background(), auditLog); err != nil {
			logx.Errorf("存储审计日志失败: %v", err)
		}
	}()
}

// storeAuditLog 存储审计日志到Redis
func (m *AuditMiddleware) storeAuditLog(ctx context.Context, auditLog *AuditLog) error {
	// 将审计日志序列化为JSON
	auditBytes, err := json.Marshal(auditLog)
	if err != nil {
		return err
	}

	// 存储到Redis List中（用于审计日志列表查询）
	auditKey := "audit_logs"
	pipe := m.redis.Pipeline()

	// 添加到列表头部
	pipe.LPush(ctx, auditKey, auditBytes)

	// 限制列表长度（保留最近10000条记录）
	pipe.LTrim(ctx, auditKey, 0, 9999)

	// 设置过期时间（30天）
	pipe.Expire(ctx, auditKey, 30*24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}
