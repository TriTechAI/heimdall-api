package utils

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	// ErrValidationFailed 验证失败错误
	ErrValidationFailed = errors.New("validation failed")
	// ErrRequiredField 必填字段错误
	ErrRequiredField = errors.New("required field is missing")
	// ErrInvalidFormat 格式错误
	ErrInvalidFormat = errors.New("invalid format")
	// ErrOutOfRange 超出范围错误
	ErrOutOfRange = errors.New("value out of range")
)

// 预编译的正则表达式
var (
	// 用户名格式：字母开头，可包含字母、数字、下划线，3-30位
	usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,29}$`)
	// 标签格式：字母开头，可包含字母、数字、连字符，2-50位
	tagRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{1,49}$`)
	// URL格式（简化版）
	urlRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	// IPv4地址格式
	ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	// 手机号格式（中国）
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	// 身份证号格式（中国）
	idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)?\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)
)

// ValidationRule 验证规则
type ValidationRule struct {
	Field    string
	Value    interface{}
	Rules    []string
	Required bool
	Message  string
}

// Validator 验证器
type Validator struct {
	errors map[string][]string
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make(map[string][]string),
	}
}

// AddError 添加错误信息
func (v *Validator) AddError(field, message string) {
	v.errors[field] = append(v.errors[field], message)
}

// HasErrors 检查是否有验证错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors 获取所有错误
func (v *Validator) GetErrors() map[string][]string {
	return v.errors
}

// GetFirstError 获取第一个错误信息
func (v *Validator) GetFirstError() string {
	for _, messages := range v.errors {
		if len(messages) > 0 {
			return messages[0]
		}
	}
	return ""
}

// Clear 清空验证错误
func (v *Validator) Clear() {
	v.errors = make(map[string][]string)
}

// Required 必填字段验证
func (v *Validator) Required(field string, value interface{}) *Validator {
	if isEmpty(value) {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
	return v
}

// Email 邮箱格式验证
func (v *Validator) Email(field string, value string) *Validator {
	if value == "" {
		return v
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid email address", field))
	}
	return v
}

// Username 用户名格式验证
func (v *Validator) Username(field string, value string) *Validator {
	if value == "" {
		return v
	}
	if !usernameRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must start with a letter and contain only letters, numbers, and underscores (3-30 characters)", field))
	}
	return v
}

// Length 长度验证
func (v *Validator) Length(field string, value string, min, max int) *Validator {
	length := len([]rune(value)) // 使用rune计算字符长度，支持Unicode
	if length < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters long", field, min))
	}
	if length > max {
		v.AddError(field, fmt.Sprintf("%s must be no more than %d characters long", field, max))
	}
	return v
}

// Range 数值范围验证
func (v *Validator) Range(field string, value, min, max int) *Validator {
	if value < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d", field, min))
	}
	if value > max {
		v.AddError(field, fmt.Sprintf("%s must be no more than %d", field, max))
	}
	return v
}

// In 枚举值验证
func (v *Validator) In(field string, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return v
		}
	}
	v.AddError(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
	return v
}

// URL URL格式验证
func (v *Validator) URL(field string, value string) *Validator {
	if value == "" {
		return v
	}
	if !urlRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must be a valid URL", field))
	}
	return v
}

// Phone 手机号验证（中国）
func (v *Validator) Phone(field string, value string) *Validator {
	if value == "" {
		return v
	}
	if !phoneRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must be a valid Chinese phone number", field))
	}
	return v
}

// IPAddress IP地址验证
func (v *Validator) IPAddress(field string, value string) *Validator {
	if value == "" {
		return v
	}
	if !ipv4Regex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must be a valid IP address", field))
	}
	return v
}

// DateTime 日期时间格式验证
func (v *Validator) DateTime(field string, value string, layout string) *Validator {
	if value == "" {
		return v
	}
	_, err := time.Parse(layout, value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid datetime in format %s", field, layout))
	}
	return v
}

// Regex 正则表达式验证
func (v *Validator) Regex(field string, value string, pattern string, message string) *Validator {
	if value == "" {
		return v
	}
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		if message == "" {
			message = fmt.Sprintf("%s format is invalid", field)
		}
		v.AddError(field, message)
	}
	return v
}

// Slug URL友好字符串验证
func (v *Validator) Slug(field string, value string) *Validator {
	if value == "" {
		return v
	}
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !slugRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must be a valid slug (lowercase letters, numbers, and hyphens only)", field))
	}
	return v
}

// Tag 标签格式验证
func (v *Validator) Tag(field string, value string) *Validator {
	if value == "" {
		return v
	}
	if !tagRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s must start with a letter and contain only letters, numbers, and hyphens (2-50 characters)", field))
	}
	return v
}

// Custom 自定义验证
func (v *Validator) Custom(field string, value interface{}, validator func(interface{}) bool, message string) *Validator {
	if !validator(value) {
		v.AddError(field, message)
	}
	return v
}

// 工具函数

// isEmpty 检查值是否为空
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case bool:
		return !v
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// ValidateRequired 必填字段验证
func ValidateRequired(field string, value interface{}) error {
	if isEmpty(value) {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// ValidateEmail 邮箱格式验证
func ValidateEmail(email string) error {
	if email == "" {
		return nil
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidFormat
	}
	return nil
}

// ValidateUsername 用户名格式验证
func ValidateUsername(username string) error {
	if username == "" {
		return nil
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username must start with a letter and contain only letters, numbers, and underscores (3-30 characters)")
	}
	return nil
}

// ValidateStringLength 字符串长度验证
func ValidateStringLength(value string, min, max int) error {
	length := len([]rune(value))
	if length < min {
		return fmt.Errorf("value must be at least %d characters long", min)
	}
	if length > max {
		return fmt.Errorf("value must be no more than %d characters long", max)
	}
	return nil
}

// ValidateIntRange 整数范围验证
func ValidateIntRange(value, min, max int) error {
	if value < min {
		return fmt.Errorf("value must be at least %d", min)
	}
	if value > max {
		return fmt.Errorf("value must be no more than %d", max)
	}
	return nil
}

// ValidateEnum 枚举值验证
func ValidateEnum(value string, allowed []string) error {
	if value == "" {
		return nil
	}
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}
	return fmt.Errorf("value must be one of: %s", strings.Join(allowed, ", "))
}

// ValidateURL URL格式验证
func ValidateURL(url string) error {
	if url == "" {
		return nil
	}
	if !urlRegex.MatchString(url) {
		return ErrInvalidFormat
	}
	return nil
}

// ValidateSlug URL友好字符串验证
func ValidateSlug(slug string) error {
	if slug == "" {
		return nil
	}
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	if !slugRegex.MatchString(slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

// ValidateTag 标签格式验证
func ValidateTag(tag string) error {
	if tag == "" {
		return nil
	}
	if !tagRegex.MatchString(tag) {
		return errors.New("tag must start with a letter and contain only letters, numbers, and hyphens (2-50 characters)")
	}
	return nil
}

// SanitizeString 清理字符串，移除危险字符
func SanitizeString(input string) string {
	// 移除控制字符
	var result strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}

	// 转义HTML特殊字符 - 先转义&，避免重复转义
	sanitized := result.String()
	sanitized = strings.ReplaceAll(sanitized, "&", "&amp;")
	sanitized = strings.ReplaceAll(sanitized, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
	sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")

	return strings.TrimSpace(sanitized)
}

// ValidatePageParams 验证分页参数
func ValidatePageParams(page, limit int) (int, int, error) {
	// 默认值
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// 范围检查
	if limit > 100 {
		limit = 100
	}

	return page, limit, nil
}

// ParseAndValidateInt 解析并验证整数
func ParseAndValidateInt(value string, min, max int) (int, error) {
	if value == "" {
		return 0, errors.New("value is required")
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrInvalidFormat
	}

	if err := ValidateIntRange(parsed, min, max); err != nil {
		return 0, err
	}

	return parsed, nil
}

// ParseAndValidateID 解析并验证ID（MongoDB ObjectID或UUID）
func ParseAndValidateID(id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	// MongoDB ObjectID格式（24位十六进制字符）
	if len(id) == 24 {
		matched, _ := regexp.MatchString(`^[a-fA-F0-9]{24}$`, id)
		if matched {
			return nil
		}
	}

	// UUID格式
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	if uuidRegex.MatchString(id) {
		return nil
	}

	return errors.New("id must be a valid MongoDB ObjectID or UUID")
}

// ValidateBatchSize 验证批量操作大小
func ValidateBatchSize(size int) error {
	const maxBatchSize = 1000
	if size <= 0 {
		return errors.New("batch size must be greater than 0")
	}
	if size > maxBatchSize {
		return fmt.Errorf("batch size must not exceed %d", maxBatchSize)
	}
	return nil
}
