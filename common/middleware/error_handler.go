package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/heimdall-api/common/errors"
)

// ====================
// 错误处理中间件
// ====================

// ErrorHandlerMiddleware 统一错误处理中间件
func ErrorHandlerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置响应头
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			// 创建自定义ResponseWriter来捕获错误
			wrapper := &responseWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 执行下一个处理器
			next.ServeHTTP(wrapper, r)
		})
	}
}

// responseWrapper 响应包装器
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader 重写WriteHeader方法
func (w *responseWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if !w.written {
		w.ResponseWriter.WriteHeader(statusCode)
		w.written = true
	}
}

// Write 重写Write方法
func (w *responseWrapper) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// ====================
// 错误响应处理函数
// ====================

// HandleError 统一错误处理函数
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	logx.WithContext(r.Context()).Errorf("Request error: %v", err)

	var bizError *errors.BusinessError

	// 检查是否为业务错误
	if errors.IsBusinessError(err) {
		bizError, _ = errors.AsBusinessError(err)
	} else {
		// 包装为内部错误
		bizError = errors.WrapInternalError(err)
	}

	// 根据错误码确定HTTP状态码
	httpStatus := getHTTPStatusFromErrorCode(bizError.GetCode())

	// 构造错误响应
	errorResponse := ErrorResponse{
		Code:      bizError.GetCode(),
		Msg:       bizError.GetMsg(),
		Details:   bizError.GetDetails(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 返回错误响应
	httpx.WriteJson(w, httpStatus, errorResponse)
}

// HandleBusinessError 处理业务错误
func HandleBusinessError(w http.ResponseWriter, r *http.Request, bizError *errors.BusinessError) {
	logx.WithContext(r.Context()).Errorf("Business error: %s - %s", bizError.GetCode(), bizError.GetMsg())

	// 根据错误码确定HTTP状态码
	httpStatus := getHTTPStatusFromErrorCode(bizError.GetCode())

	// 构造错误响应
	errorResponse := ErrorResponse{
		Code:      bizError.GetCode(),
		Msg:       bizError.GetMsg(),
		Details:   bizError.GetDetails(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 返回错误响应
	httpx.WriteJson(w, httpStatus, errorResponse)
}

// HandleSuccess 处理成功响应
func HandleSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	successResponse := SuccessResponse{
		Code:      errors.CodeSuccess,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	httpx.WriteJson(w, http.StatusOK, successResponse)
}

// ====================
// 响应结构体定义
// ====================

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      string      `json:"code"`
	Msg       string      `json:"msg"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

// ====================
// HTTP状态码映射
// ====================

// getHTTPStatusFromErrorCode 根据错误码获取HTTP状态码
func getHTTPStatusFromErrorCode(errorCode string) int {
	switch errorCode {
	// 成功
	case errors.CodeSuccess:
		return http.StatusOK

	// 客户端错误 4xx
	case errors.CodeInvalidRequest, errors.CodeValidationFailed:
		return http.StatusBadRequest

	case errors.CodeUnauthorized, errors.CodeAuthFailed, errors.CodeTokenInvalid,
		errors.CodeTokenExpired, errors.CodeLoginFailed, errors.CodePasswordIncorrect:
		return http.StatusUnauthorized

	case errors.CodeForbidden, errors.CodePermissionDenied, errors.CodeUserDisabled,
		errors.CodeUserLocked:
		return http.StatusForbidden

	case errors.CodeNotFound, errors.CodeUserNotFound, errors.CodePostNotFound,
		errors.CodePageNotFound, errors.CodeCommentNotFound, errors.CodeTagNotFound,
		errors.CodeMediaNotFound, errors.CodeSettingNotFound:
		return http.StatusNotFound

	case errors.CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed

	case errors.CodeConflict, errors.CodeUserExists, errors.CodeUsernameExists,
		errors.CodeEmailExists, errors.CodeSlugExists, errors.CodeTagExists:
		return http.StatusConflict

	case errors.CodeTooManyRequests, errors.CodeTooManyAttempts:
		return http.StatusTooManyRequests

	case errors.CodeContentTooLong, errors.CodeFileTooLarge, errors.CodeTooManyTags:
		return http.StatusRequestEntityTooLarge

	case errors.CodeFileTypeNotAllowed:
		return http.StatusUnsupportedMediaType

	// 服务器错误 5xx
	case errors.CodeInternalError, errors.CodeDatabaseError, errors.CodeCacheError,
		errors.CodeExternalServiceError, errors.CodeBusinessLogicError:
		return http.StatusInternalServerError

	case errors.CodeServiceUnavailable:
		return http.StatusServiceUnavailable

	case errors.CodeTimeout:
		return http.StatusGatewayTimeout

	// 默认为内部服务器错误
	default:
		return http.StatusInternalServerError
	}
}

// ====================
// 便捷函数
// ====================

// WriteErrorResponse 写入错误响应
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, code, msg string) {
	bizError := errors.New(code, msg)
	HandleBusinessError(w, r, bizError)
}

// WriteErrorResponseWithDetails 写入带详情的错误响应
func WriteErrorResponseWithDetails(w http.ResponseWriter, r *http.Request, code, msg string, details interface{}) {
	bizError := errors.NewWithDetails(code, msg, details)
	HandleBusinessError(w, r, bizError)
}

// WriteSuccessResponse 写入成功响应
func WriteSuccessResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	HandleSuccess(w, r, data)
}

// WriteInternalError 写入内部错误响应
func WriteInternalError(w http.ResponseWriter, r *http.Request) {
	bizError := errors.InternalError()
	HandleBusinessError(w, r, bizError)
}

// WriteUnauthorized 写入未授权响应
func WriteUnauthorized(w http.ResponseWriter, r *http.Request) {
	bizError := errors.Unauthorized()
	HandleBusinessError(w, r, bizError)
}

// WriteForbidden 写入禁止访问响应
func WriteForbidden(w http.ResponseWriter, r *http.Request) {
	bizError := errors.Forbidden()
	HandleBusinessError(w, r, bizError)
}

// WriteNotFound 写入资源不存在响应
func WriteNotFound(w http.ResponseWriter, r *http.Request, resource string) {
	bizError := errors.NotFound(resource)
	HandleBusinessError(w, r, bizError)
}

// WriteValidationError 写入验证错误响应
func WriteValidationError(w http.ResponseWriter, r *http.Request, details interface{}) {
	bizError := errors.ValidationFailed(details)
	HandleBusinessError(w, r, bizError)
}

// ====================
// 错误恢复中间件
// ====================

// RecoveryMiddleware panic恢复中间件
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logx.WithContext(r.Context()).Errorf("Panic recovered: %v", err)
					WriteInternalError(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
