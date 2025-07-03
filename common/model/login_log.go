package model

import (
	"time"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID      *primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`         // 可能为空（登录失败时）
	Username    string              `bson:"username" json:"username"`                         // 登录时使用的用户名
	LoginMethod string              `bson:"loginMethod" json:"loginMethod"`                   // 登录方式：username, email
	IPAddress   string              `bson:"ipAddress" json:"ipAddress"`                       // IP地址
	UserAgent   string              `bson:"userAgent" json:"userAgent"`                       // 用户代理
	Status      string              `bson:"status" json:"status"`                             // 登录状态：success, failed
	FailReason  string              `bson:"failReason,omitempty" json:"failReason,omitempty"` // 失败原因
	SessionID   string              `bson:"sessionId,omitempty" json:"sessionId,omitempty"`   // 会话ID（成功时）
	LoginAt     time.Time           `bson:"loginAt" json:"loginAt"`                           // 登录时间
	LogoutAt    *time.Time          `bson:"logoutAt,omitempty" json:"logoutAt,omitempty"`     // 登出时间
	Duration    *int64              `bson:"duration,omitempty" json:"duration,omitempty"`     // 会话持续时间（秒）
	Country     string              `bson:"country,omitempty" json:"country,omitempty"`       // 国家
	Region      string              `bson:"region,omitempty" json:"region,omitempty"`         // 地区
	City        string              `bson:"city,omitempty" json:"city,omitempty"`             // 城市
	DeviceType  string              `bson:"deviceType,omitempty" json:"deviceType,omitempty"` // 设备类型
	Browser     string              `bson:"browser,omitempty" json:"browser,omitempty"`       // 浏览器
	OS          string              `bson:"os,omitempty" json:"os,omitempty"`                 // 操作系统
	CreatedAt   time.Time           `bson:"createdAt" json:"createdAt"`                       // 记录创建时间
}

// LoginLogCreateRequest 创建登录日志请求
type LoginLogCreateRequest struct {
	UserID      *primitive.ObjectID `json:"userId,omitempty"`
	Username    string              `json:"username" validate:"required"`
	LoginMethod string              `json:"loginMethod" validate:"required,oneof=username email"`
	IPAddress   string              `json:"ipAddress" validate:"required,ip"`
	UserAgent   string              `json:"userAgent" validate:"required"`
	Status      string              `json:"status" validate:"required,oneof=success failed"`
	FailReason  string              `json:"failReason,omitempty"`
	SessionID   string              `json:"sessionId,omitempty"`
	Country     string              `json:"country,omitempty"`
	Region      string              `json:"region,omitempty"`
	City        string              `json:"city,omitempty"`
	DeviceType  string              `json:"deviceType,omitempty"`
	Browser     string              `json:"browser,omitempty"`
	OS          string              `json:"os,omitempty"`
}

// LoginLogFilter 登录日志过滤器
type LoginLogFilter struct {
	UserID     string     `json:"userId"`     // 用户ID
	Username   string     `json:"username"`   // 用户名（模糊搜索）
	Status     string     `json:"status"`     // 登录状态
	IPAddress  string     `json:"ipAddress"`  // IP地址
	StartTime  *time.Time `json:"startTime"`  // 开始时间
	EndTime    *time.Time `json:"endTime"`    // 结束时间
	Country    string     `json:"country"`    // 国家
	Region     string     `json:"region"`     // 地区
	City       string     `json:"city"`       // 城市
	DeviceType string     `json:"deviceType"` // 设备类型
	Browser    string     `json:"browser"`    // 浏览器
	OS         string     `json:"os"`         // 操作系统
	Page       int        `json:"page"`       // 页码
	Limit      int        `json:"limit"`      // 每页数量
	SortBy     string     `json:"sortBy"`     // 排序字段
	SortDesc   bool       `json:"sortDesc"`   // 是否降序
}

// LoginLogListItem 登录日志列表项
type LoginLogListItem struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId,omitempty"`
	Username    string     `json:"username"`
	LoginMethod string     `json:"loginMethod"`
	IPAddress   string     `json:"ipAddress"`
	Status      string     `json:"status"`
	FailReason  string     `json:"failReason,omitempty"`
	Country     string     `json:"country,omitempty"`
	Region      string     `json:"region,omitempty"`
	City        string     `json:"city,omitempty"`
	DeviceType  string     `json:"deviceType,omitempty"`
	Browser     string     `json:"browser,omitempty"`
	OS          string     `json:"os,omitempty"`
	LoginAt     time.Time  `json:"loginAt"`
	LogoutAt    *time.Time `json:"logoutAt,omitempty"`
	Duration    *int64     `json:"duration,omitempty"`
}

// LoginStatistics 登录统计信息
type LoginStatistics struct {
	TotalLogins   int64   `json:"totalLogins"`   // 总登录次数
	SuccessLogins int64   `json:"successLogins"` // 成功登录次数
	FailedLogins  int64   `json:"failedLogins"`  // 失败登录次数
	UniqueUsers   int64   `json:"uniqueUsers"`   // 独立用户数
	UniqueIPs     int64   `json:"uniqueIPs"`     // 独立IP数
	SuccessRate   float64 `json:"successRate"`   // 成功率
}

// ===============================
// 验证方法
// ===============================

// ValidateForCreate 验证登录日志创建数据
func (l *LoginLog) ValidateForCreate() error {
	// 验证必填字段
	if l.Username == "" {
		return NewValidationError("username", "用户名不能为空")
	}
	if l.LoginMethod == "" {
		return NewValidationError("loginMethod", "登录方式不能为空")
	}
	if l.IPAddress == "" {
		return NewValidationError("ipAddress", "IP地址不能为空")
	}
	if l.UserAgent == "" {
		return NewValidationError("userAgent", "用户代理不能为空")
	}
	if l.Status == "" {
		return NewValidationError("status", "登录状态不能为空")
	}

	// 验证登录方式
	if !isValidLoginMethod(l.LoginMethod) {
		return NewValidationError("loginMethod", "无效的登录方式")
	}

	// 验证登录状态
	if !isValidLoginStatus(l.Status) {
		return NewValidationError("status", "无效的登录状态")
	}

	// 验证失败原因
	if l.Status == constants.LoginStatusFailed && l.FailReason == "" {
		return NewValidationError("failReason", "登录失败时必须提供失败原因")
	}

	// 验证成功登录必须有用户ID
	if l.Status == constants.LoginStatusSuccess && (l.UserID == nil || l.UserID.IsZero()) {
		return NewValidationError("userId", "登录成功时必须提供用户ID")
	}

	// 验证字段长度
	if len(l.Username) > 64 {
		return NewValidationError("username", "用户名长度不能超过64字符")
	}
	if len(l.IPAddress) > 45 { // IPv6最大长度
		return NewValidationError("ipAddress", "IP地址格式错误")
	}
	if len(l.UserAgent) > 512 {
		return NewValidationError("userAgent", "用户代理长度不能超过512字符")
	}
	if len(l.Country) > 100 {
		return NewValidationError("country", "国家名称长度不能超过100字符")
	}
	if len(l.Region) > 100 {
		return NewValidationError("region", "地区名称长度不能超过100字符")
	}
	if len(l.City) > 100 {
		return NewValidationError("city", "城市名称长度不能超过100字符")
	}

	return nil
}

// ===============================
// 状态检查方法
// ===============================

// IsSuccess 检查是否登录成功
func (l *LoginLog) IsSuccess() bool {
	return l.Status == constants.LoginStatusSuccess
}

// IsFailed 检查是否登录失败
func (l *LoginLog) IsFailed() bool {
	return l.Status == constants.LoginStatusFailed
}

// IsActiveSession 检查会话是否仍然活跃
func (l *LoginLog) IsActiveSession() bool {
	return l.IsSuccess() && l.LogoutAt == nil
}

// GetSessionDuration 获取会话持续时间（秒）
func (l *LoginLog) GetSessionDuration() int64 {
	if l.Duration != nil {
		return *l.Duration
	}

	if l.LogoutAt != nil {
		return int64(l.LogoutAt.Sub(l.LoginAt).Seconds())
	}

	// 如果会话仍然活跃，返回当前持续时间
	if l.IsActiveSession() {
		return int64(time.Since(l.LoginAt).Seconds())
	}

	return 0
}

// ===============================
// 转换方法
// ===============================

// ToListItem 转换为列表项
func (l *LoginLog) ToListItem() *LoginLogListItem {
	item := &LoginLogListItem{
		ID:          l.ID.Hex(),
		Username:    l.Username,
		LoginMethod: l.LoginMethod,
		IPAddress:   l.IPAddress,
		Status:      l.Status,
		FailReason:  l.FailReason,
		Country:     l.Country,
		Region:      l.Region,
		City:        l.City,
		DeviceType:  l.DeviceType,
		Browser:     l.Browser,
		OS:          l.OS,
		LoginAt:     l.LoginAt,
		LogoutAt:    l.LogoutAt,
		Duration:    l.Duration,
	}

	if l.UserID != nil {
		item.UserID = l.UserID.Hex()
	}

	return item
}

// ===============================
// 工厂方法
// ===============================

// NewLoginLog 创建新的登录日志
func NewLoginLog(username, loginMethod, ipAddress, userAgent, status string) *LoginLog {
	now := time.Now()

	log := &LoginLog{
		ID:          primitive.NewObjectID(),
		Username:    username,
		LoginMethod: loginMethod,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      status,
		LoginAt:     now,
		CreatedAt:   now,
	}

	return log
}

// NewSuccessLoginLog 创建成功登录日志
func NewSuccessLoginLog(userID primitive.ObjectID, username, loginMethod, ipAddress, userAgent, sessionID string) *LoginLog {
	log := NewLoginLog(username, loginMethod, ipAddress, userAgent, constants.LoginStatusSuccess)
	log.UserID = &userID
	log.SessionID = sessionID
	return log
}

// NewFailedLoginLog 创建失败登录日志
func NewFailedLoginLog(username, loginMethod, ipAddress, userAgent, failReason string) *LoginLog {
	log := NewLoginLog(username, loginMethod, ipAddress, userAgent, constants.LoginStatusFailed)
	log.FailReason = failReason
	return log
}

// NewLoginLogFromRequest 从请求创建登录日志
func NewLoginLogFromRequest(req *LoginLogCreateRequest) *LoginLog {
	log := NewLoginLog(req.Username, req.LoginMethod, req.IPAddress, req.UserAgent, req.Status)

	// 设置可选字段
	log.UserID = req.UserID
	log.FailReason = req.FailReason
	log.SessionID = req.SessionID
	log.Country = req.Country
	log.Region = req.Region
	log.City = req.City
	log.DeviceType = req.DeviceType
	log.Browser = req.Browser
	log.OS = req.OS

	return log
}

// ===============================
// 数据库操作辅助方法
// ===============================

// PrepareForInsert 准备插入数据库
func (l *LoginLog) PrepareForInsert() {
	now := time.Now()
	if l.ID.IsZero() {
		l.ID = primitive.NewObjectID()
	}
	if l.LoginAt.IsZero() {
		l.LoginAt = now
	}
	l.CreatedAt = now
}

// MarkLogout 标记登出
func (l *LoginLog) MarkLogout() {
	now := time.Now()
	l.LogoutAt = &now

	// 计算会话持续时间
	duration := int64(now.Sub(l.LoginAt).Seconds())
	l.Duration = &duration
}

// UpdateLocation 更新地理位置信息
func (l *LoginLog) UpdateLocation(country, region, city string) {
	l.Country = country
	l.Region = region
	l.City = city
}

// UpdateDeviceInfo 更新设备信息
func (l *LoginLog) UpdateDeviceInfo(deviceType, browser, os string) {
	l.DeviceType = deviceType
	l.Browser = browser
	l.OS = os
}

// ===============================
// 验证辅助函数
// ===============================

// isValidLoginMethod 验证登录方式是否有效
func isValidLoginMethod(method string) bool {
	validMethods := []string{"username", "email"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// isValidLoginStatus 验证登录状态是否有效
func isValidLoginStatus(status string) bool {
	validStatuses := []string{constants.LoginStatusSuccess, constants.LoginStatusFailed}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}
