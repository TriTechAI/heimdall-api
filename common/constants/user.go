package constants

// UserRole 用户角色常量
const (
	UserRoleOwner  = "owner"  // 博客所有者，拥有所有权限
	UserRoleAdmin  = "admin"  // 管理员，除所有者设置外的所有权限
	UserRoleEditor = "editor" // 编辑，可管理所有内容和评论
	UserRoleAuthor = "author" // 作者，只能管理自己创建的内容
)

// UserStatus 用户状态常量
const (
	UserStatusActive    = "active"    // 正常状态
	UserStatusInactive  = "inactive"  // 非活跃状态
	UserStatusLocked    = "locked"    // 锁定状态
	UserStatusSuspended = "suspended" // 暂停状态
)

// LoginStatus 登录状态常量
const (
	LoginStatusSuccess = "success" // 登录成功
	LoginStatusFailed  = "failed"  // 登录失败
)

// LoginFailReason 登录失败原因常量
const (
	LoginFailReasonInvalidPassword = "invalid_password"  // 密码错误
	LoginFailReasonUserNotFound    = "user_not_found"    // 用户不存在
	LoginFailReasonUserLocked      = "user_locked"       // 账号被锁定
	LoginFailReasonUserInactive    = "user_inactive"     // 账号未激活
	LoginFailReasonUserSuspended   = "user_suspended"    // 账号被暂停
	LoginFailReasonTooManyAttempts = "too_many_attempts" // 尝试次数过多
)

// AccountLockDuration 账号锁定时长（分钟）
const (
	LockDuration3Failures  = 15   // 3次失败锁定15分钟
	LockDuration5Failures  = 60   // 5次失败锁定1小时
	LockDuration10Failures = 1440 // 10次失败锁定24小时
)

// PasswordPolicy 密码策略常量
const (
	PasswordMinLength    = 8  // 密码最小长度
	PasswordMaxLength    = 64 // 密码最大长度
	PasswordBcryptCost   = 12 // bcrypt成本因子
	PasswordHistoryCount = 5  // 密码历史记录数量
	PasswordExpireDays   = 90 // 密码过期天数
)

// SessionLimits 会话限制常量
const (
	MaxConcurrentSessions = 3     // 单用户最大并发会话数
	SessionTimeoutMinutes = 120   // 会话超时时间(分钟)
	SessionRefreshMinutes = 10080 // 刷新令牌有效期(7天)
)

// UserValidation 用户验证相关常量
const (
	UsernameMinLength      = 3   // 用户名最小长度
	UsernameMaxLength      = 32  // 用户名最大长度
	DisplayNameMaxLength   = 64  // 显示名最大长度
	BioMaxLength           = 500 // 简介最大长度
	LocationMaxLength      = 100 // 地址最大长度
	WebsiteMaxLength       = 255 // 网站URL最大长度
	SocialAccountMaxLength = 50  // 社交账号最大长度
)

// GetAllUserRoles 返回所有用户角色
func GetAllUserRoles() []string {
	return []string{
		UserRoleOwner,
		UserRoleAdmin,
		UserRoleEditor,
		UserRoleAuthor,
	}
}

// GetAllUserStatuses 返回所有用户状态
func GetAllUserStatuses() []string {
	return []string{
		UserStatusActive,
		UserStatusInactive,
		UserStatusLocked,
		UserStatusSuspended,
	}
}

// IsValidUserRole 验证用户角色是否有效
func IsValidUserRole(role string) bool {
	validRoles := GetAllUserRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsValidUserStatus 验证用户状态是否有效
func IsValidUserStatus(status string) bool {
	validStatuses := GetAllUserStatuses()
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// GetLockDurationByFailCount 根据失败次数获取锁定时长（分钟）
func GetLockDurationByFailCount(failCount int) int {
	switch {
	case failCount >= 10:
		return LockDuration10Failures
	case failCount >= 5:
		return LockDuration5Failures
	case failCount >= 3:
		return LockDuration3Failures
	default:
		return 0
	}
}
