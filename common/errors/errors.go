package errors

import (
	"fmt"
	"time"
)

// ====================
// 业务错误类型定义
// ====================

// BusinessError 业务错误结构
type BusinessError struct {
	Code      string      `json:"code"`
	Msg       string      `json:"msg"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("Code: %s, Msg: %s", e.Code, e.Msg)
}

// GetCode 获取错误码
func (e *BusinessError) GetCode() string {
	return e.Code
}

// GetMsg 获取错误消息
func (e *BusinessError) GetMsg() string {
	return e.Msg
}

// GetDetails 获取错误详情
func (e *BusinessError) GetDetails() interface{} {
	return e.Details
}

// WithDetails 添加错误详情
func (e *BusinessError) WithDetails(details interface{}) *BusinessError {
	e.Details = details
	return e
}

// ====================
// 错误构造函数
// ====================

// New 创建新的业务错误
func New(code, msg string) *BusinessError {
	return &BusinessError{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// NewWithCode 根据错误码创建业务错误
func NewWithCode(code string) *BusinessError {
	msg := GetErrorMessage(code)
	return &BusinessError{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// NewWithDetails 创建带详情的业务错误
func NewWithDetails(code, msg string, details interface{}) *BusinessError {
	return &BusinessError{
		Code:      code,
		Msg:       msg,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// NewWithCodeAndDetails 根据错误码创建带详情的业务错误
func NewWithCodeAndDetails(code string, details interface{}) *BusinessError {
	msg := GetErrorMessage(code)
	return &BusinessError{
		Code:      code,
		Msg:       msg,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// ====================
// 常用错误构造函数
// ====================

// InternalError 内部服务器错误
func InternalError() *BusinessError {
	return NewWithCode(CodeInternalError)
}

// InvalidRequest 请求参数无效
func InvalidRequest(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodeInvalidRequest)
	}
	return New(CodeInvalidRequest, msg)
}

// Unauthorized 未授权
func Unauthorized() *BusinessError {
	return NewWithCode(CodeUnauthorized)
}

// Forbidden 禁止访问
func Forbidden() *BusinessError {
	return NewWithCode(CodeForbidden)
}

// NotFound 资源不存在
func NotFound(resource string) *BusinessError {
	if resource == "" {
		return NewWithCode(CodeNotFound)
	}
	return New(CodeNotFound, fmt.Sprintf("%s不存在", resource))
}

// ValidationFailed 数据验证失败
func ValidationFailed(details interface{}) *BusinessError {
	return NewWithCodeAndDetails(CodeValidationFailed, details)
}

// TooManyRequests 请求过于频繁
func TooManyRequests() *BusinessError {
	return NewWithCode(CodeTooManyRequests)
}

// ====================
// 认证相关错误
// ====================

// AuthFailed 认证失败
func AuthFailed() *BusinessError {
	return NewWithCode(CodeAuthFailed)
}

// TokenInvalid Token无效
func TokenInvalid() *BusinessError {
	return NewWithCode(CodeTokenInvalid)
}

// TokenExpired Token已过期
func TokenExpired() *BusinessError {
	return NewWithCode(CodeTokenExpired)
}

// LoginFailed 登录失败
func LoginFailed(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodeLoginFailed)
	}
	return New(CodeLoginFailed, msg)
}

// PasswordIncorrect 密码错误
func PasswordIncorrect() *BusinessError {
	return NewWithCode(CodePasswordIncorrect)
}

// UserNotFound 用户不存在
func UserNotFound() *BusinessError {
	return NewWithCode(CodeUserNotFound)
}

// UserDisabled 用户已禁用
func UserDisabled() *BusinessError {
	return NewWithCode(CodeUserDisabled)
}

// UserLocked 用户已锁定
func UserLocked(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodeUserLocked)
	}
	return New(CodeUserLocked, msg)
}

// TooManyAttempts 登录尝试次数过多
func TooManyAttempts(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodeTooManyAttempts)
	}
	return New(CodeTooManyAttempts, msg)
}

// PermissionDenied 权限不足
func PermissionDenied() *BusinessError {
	return NewWithCode(CodePermissionDenied)
}

// ====================
// 内容管理相关错误
// ====================

// PostNotFound 文章不存在
func PostNotFound() *BusinessError {
	return NewWithCode(CodePostNotFound)
}

// PostCreateFailed 文章创建失败
func PostCreateFailed(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodePostCreateFailed)
	}
	return New(CodePostCreateFailed, msg)
}

// SlugExists Slug已存在
func SlugExists() *BusinessError {
	return NewWithCode(CodeSlugExists)
}

// PageNotFound 页面不存在
func PageNotFound() *BusinessError {
	return NewWithCode(CodePageNotFound)
}

// ====================
// 数据库相关错误
// ====================

// DatabaseError 数据库错误
func DatabaseError(msg string) *BusinessError {
	if msg == "" {
		return NewWithCode(CodeDatabaseError)
	}
	return New(CodeDatabaseError, msg)
}

// QueryFailed 查询失败
func QueryFailed() *BusinessError {
	return NewWithCode(CodeQueryFailed)
}

// InsertFailed 插入失败
func InsertFailed() *BusinessError {
	return NewWithCode(CodeInsertFailed)
}

// UpdateFailed 更新失败
func UpdateFailed() *BusinessError {
	return NewWithCode(CodeUpdateFailed)
}

// DeleteFailed 删除失败
func DeleteFailed() *BusinessError {
	return NewWithCode(CodeDeleteFailed)
}

// ====================
// 错误检查函数
// ====================

// IsBusinessError 检查是否为业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// AsBusinessError 转换为业务错误
func AsBusinessError(err error) (*BusinessError, bool) {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr, true
	}
	return nil, false
}

// WrapError 包装普通错误为业务错误
func WrapError(err error, code string) *BusinessError {
	if err == nil {
		return nil
	}

	// 如果已经是业务错误，直接返回
	if bizErr, ok := AsBusinessError(err); ok {
		return bizErr
	}

	// 包装为业务错误
	msg := GetErrorMessage(code)
	return NewWithDetails(code, msg, err.Error())
}

// WrapInternalError 包装为内部错误
func WrapInternalError(err error) *BusinessError {
	if err == nil {
		return nil
	}
	return WrapError(err, CodeInternalError)
}

// ====================
// 错误响应格式化
// ====================

// ToResponse 转换为响应格式
func (e *BusinessError) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"code":      e.Code,
		"msg":       e.Msg,
		"timestamp": e.Timestamp,
	}

	if e.Details != nil {
		response["details"] = e.Details
	}

	return response
}

// ToJSON 转换为JSON字符串
func (e *BusinessError) ToJSON() string {
	return fmt.Sprintf(`{"code":"%s","msg":"%s","timestamp":"%s"}`,
		e.Code, e.Msg, e.Timestamp)
}
