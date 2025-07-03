package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"msg"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationData 分页数据结构
type PaginationData struct {
	List       interface{}        `json:"list"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata 分页元数据
type PaginationMetadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// 常用状态码
const (
	StatusOK                  = http.StatusOK                  // 200
	StatusCreated             = http.StatusCreated             // 201
	StatusNoContent           = http.StatusNoContent           // 204
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusConflict            = http.StatusConflict            // 409
	StatusTooManyRequests     = http.StatusTooManyRequests     // 429
	StatusInternalServerError = http.StatusInternalServerError // 500
)

// 常用错误码
const (
	ErrCodeSuccess           = "success"
	ErrCodeValidationFailed  = "validation_failed"
	ErrCodeUnauthorized      = "unauthorized"
	ErrCodeForbidden         = "forbidden"
	ErrCodeNotFound          = "resource_not_found"
	ErrCodeConflict          = "resource_conflict"
	ErrCodeTooManyRequests   = "too_many_requests"
	ErrCodeInternalError     = "internal_error"
	ErrCodeInvalidToken      = "invalid_token"
	ErrCodeTokenExpired      = "token_expired"
	ErrCodeUsernameExists    = "username_exists"
	ErrCodeEmailExists       = "email_exists"
	ErrCodeWeakPassword      = "weak_password"
	ErrCodeLoginFailed       = "login_failed"
	ErrCodeAccountLocked     = "account_locked"
	ErrCodePermissionDenied  = "permission_denied"
	ErrCodeRateLimitExceeded = "rate_limit_exceeded"
)

// Success 返回成功响应
func Success(w http.ResponseWriter, data interface{}) {
	response := Response{
		Code:      StatusOK,
		Message:   "Success",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSONResponse(w, StatusOK, response)
}

// Created 返回创建成功响应
func Created(w http.ResponseWriter, data interface{}) {
	response := Response{
		Code:      StatusCreated,
		Message:   "Created",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSONResponse(w, StatusCreated, response)
}

// NoContent 返回无内容响应
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(StatusNoContent)
}

// Error 返回错误响应
func Error(w http.ResponseWriter, statusCode int, errCode, message string, details interface{}) {
	response := ErrorResponse{
		Code:    errCode,
		Message: message,
		Details: details,
	}
	writeJSONResponse(w, statusCode, response)
}

// BadRequest 返回400错误响应
func BadRequest(w http.ResponseWriter, message string, details interface{}) {
	Error(w, StatusBadRequest, ErrCodeValidationFailed, message, details)
}

// Unauthorized 返回401错误响应
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	Error(w, StatusUnauthorized, ErrCodeUnauthorized, message, nil)
}

// Forbidden 返回403错误响应
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Permission denied"
	}
	Error(w, StatusForbidden, ErrCodeForbidden, message, nil)
}

// NotFound 返回404错误响应
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, StatusNotFound, ErrCodeNotFound, message, nil)
}

// Conflict 返回409错误响应
func Conflict(w http.ResponseWriter, message string, details interface{}) {
	Error(w, StatusConflict, ErrCodeConflict, message, details)
}

// TooManyRequests 返回429错误响应
func TooManyRequests(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Too many requests"
	}
	Error(w, StatusTooManyRequests, ErrCodeTooManyRequests, message, nil)
}

// InternalError 返回500错误响应
func InternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	Error(w, StatusInternalServerError, ErrCodeInternalError, message, nil)
}

// SuccessWithPagination 返回带分页的成功响应
func SuccessWithPagination(w http.ResponseWriter, list interface{}, page, limit int, total int64) {
	totalPages := calculateTotalPages(total, limit)

	paginationData := PaginationData{
		List: list,
		Pagination: PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	response := Response{
		Code:      StatusOK,
		Message:   "Success",
		Data:      paginationData,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSONResponse(w, StatusOK, response)
}

// ValidationError 返回验证错误响应
func ValidationError(w http.ResponseWriter, errors map[string][]string) {
	BadRequest(w, "Validation failed", errors)
}

// CustomError 返回自定义错误响应
func CustomError(w http.ResponseWriter, statusCode int, errCode, message string, details interface{}) {
	Error(w, statusCode, errCode, message, details)
}

// LoginFailed 返回登录失败响应
func LoginFailed(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Invalid username or password"
	}
	Error(w, StatusUnauthorized, ErrCodeLoginFailed, message, nil)
}

// TokenExpired 返回令牌过期响应
func TokenExpired(w http.ResponseWriter) {
	Error(w, StatusUnauthorized, ErrCodeTokenExpired, "Token has expired", nil)
}

// TokenInvalid 返回令牌无效响应
func TokenInvalid(w http.ResponseWriter) {
	Error(w, StatusUnauthorized, ErrCodeInvalidToken, "Invalid token", nil)
}

// AccountLocked 返回账户锁定响应
func AccountLocked(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Account is locked due to too many failed login attempts"
	}
	Error(w, StatusForbidden, ErrCodeAccountLocked, message, nil)
}

// UsernameExists 返回用户名已存在响应
func UsernameExists(w http.ResponseWriter) {
	Error(w, StatusConflict, ErrCodeUsernameExists, "Username already exists", map[string]string{"field": "username"})
}

// EmailExists 返回邮箱已存在响应
func EmailExists(w http.ResponseWriter) {
	Error(w, StatusConflict, ErrCodeEmailExists, "Email already exists", map[string]string{"field": "email"})
}

// WeakPassword 返回密码强度不足响应
func WeakPassword(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Password does not meet strength requirements"
	}
	BadRequest(w, message, map[string]string{"field": "password"})
}

// RateLimitExceeded 返回限流超出响应
func RateLimitExceeded(w http.ResponseWriter, retryAfter string) {
	w.Header().Set("Retry-After", retryAfter)
	Error(w, StatusTooManyRequests, ErrCodeRateLimitExceeded, "Rate limit exceeded", nil)
}

// writeJSONResponse 写入JSON响应
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// 如果JSON编码失败，返回纯文本错误
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}

// calculateTotalPages 计算总页数
func calculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return totalPages
}

// CreatePaginationMetadata 创建分页元数据
func CreatePaginationMetadata(page, limit int, total int64) PaginationMetadata {
	totalPages := calculateTotalPages(total, limit)

	return PaginationMetadata{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetContentType 获取内容类型
func GetContentType(w http.ResponseWriter) string {
	return w.Header().Get("Content-Type")
}

// SetCacheHeaders 设置缓存头
func SetCacheHeaders(w http.ResponseWriter, maxAge int) {
	if maxAge > 0 {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}

// SetSecurityHeaders 设置安全头
func SetSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// SetCORSHeaders 设置CORS头
func SetCORSHeaders(w http.ResponseWriter, allowOrigin string) {
	if allowOrigin == "" {
		allowOrigin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

// HandlePreflight 处理预检请求
func HandlePreflight(w http.ResponseWriter) {
	SetCORSHeaders(w, "*")
	w.WriteHeader(StatusNoContent)
}

// ResponseMiddleware 响应中间件，自动设置通用头
type ResponseMiddleware struct {
	enableSecurity bool
	enableCORS     bool
	corsOrigin     string
}

// NewResponseMiddleware 创建响应中间件
func NewResponseMiddleware(enableSecurity, enableCORS bool, corsOrigin string) *ResponseMiddleware {
	return &ResponseMiddleware{
		enableSecurity: enableSecurity,
		enableCORS:     enableCORS,
		corsOrigin:     corsOrigin,
	}
}

// Wrap 包装响应写入器
func (rm *ResponseMiddleware) Wrap(w http.ResponseWriter) http.ResponseWriter {
	if rm.enableSecurity {
		SetSecurityHeaders(w)
	}

	if rm.enableCORS {
		SetCORSHeaders(w, rm.corsOrigin)
	}

	return w
}

// APIResponse 通用API响应接口
type APIResponse interface {
	WriteResponse(w http.ResponseWriter)
}

// SuccessResponse 成功响应实现
type SuccessResponse struct {
	Data interface{}
}

// WriteResponse 实现APIResponse接口
func (r SuccessResponse) WriteResponse(w http.ResponseWriter) {
	Success(w, r.Data)
}

// ErrorResponseImpl 错误响应实现
type ErrorResponseImpl struct {
	StatusCode int
	ErrCode    string
	Message    string
	Details    interface{}
}

// WriteResponse 实现APIResponse接口
func (r ErrorResponseImpl) WriteResponse(w http.ResponseWriter) {
	Error(w, r.StatusCode, r.ErrCode, r.Message, r.Details)
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) APIResponse {
	return SuccessResponse{Data: data}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(statusCode int, errCode, message string, details interface{}) APIResponse {
	return ErrorResponseImpl{
		StatusCode: statusCode,
		ErrCode:    errCode,
		Message:    message,
		Details:    details,
	}
}
