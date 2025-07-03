package model

import (
	"time"

	"github.com/heimdall-api/common/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户模型
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	Email          string             `bson:"email" json:"email"`
	PasswordHash   string             `bson:"passwordHash" json:"-"` // 不在JSON中返回
	DisplayName    string             `bson:"displayName" json:"displayName"`
	Role           string             `bson:"role" json:"role"`
	ProfileImage   string             `bson:"profileImage" json:"profileImage"`
	CoverImage     string             `bson:"coverImage" json:"coverImage"`
	Bio            string             `bson:"bio" json:"bio"`
	Location       string             `bson:"location" json:"location"`
	Website        string             `bson:"website" json:"website"`
	Twitter        string             `bson:"twitter" json:"twitter"`
	Facebook       string             `bson:"facebook" json:"facebook"`
	Status         string             `bson:"status" json:"status"`
	LoginFailCount int                `bson:"loginFailCount" json:"loginFailCount"`
	LockedUntil    *time.Time         `bson:"lockedUntil,omitempty" json:"lockedUntil,omitempty"`
	LastLoginAt    *time.Time         `bson:"lastLoginAt,omitempty" json:"lastLoginAt,omitempty"`
	LastLoginIP    string             `bson:"lastLoginIP" json:"lastLoginIP"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=32"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8,max=64"`
	DisplayName string `json:"displayName" validate:"required,max=64"`
	Role        string `json:"role" validate:"required,oneof=owner admin editor author"`
	Bio         string `json:"bio" validate:"max=500"`
	Location    string `json:"location" validate:"max=100"`
	Website     string `json:"website" validate:"omitempty,url,max=255"`
	Twitter     string `json:"twitter" validate:"max=50"`
	Facebook    string `json:"facebook" validate:"max=50"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	DisplayName  string `json:"displayName" validate:"omitempty,max=64"`
	Bio          string `json:"bio" validate:"max=500"`
	Location     string `json:"location" validate:"max=100"`
	Website      string `json:"website" validate:"omitempty,url,max=255"`
	Twitter      string `json:"twitter" validate:"max=50"`
	Facebook     string `json:"facebook" validate:"max=50"`
	ProfileImage string `json:"profileImage" validate:"omitempty,url"`
	CoverImage   string `json:"coverImage" validate:"omitempty,url"`
}

// UserProfileResponse 用户档案响应
type UserProfileResponse struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"displayName"`
	Role         string     `json:"role"`
	ProfileImage string     `json:"profileImage"`
	CoverImage   string     `json:"coverImage"`
	Bio          string     `json:"bio"`
	Location     string     `json:"location"`
	Website      string     `json:"website"`
	Twitter      string     `json:"twitter"`
	Facebook     string     `json:"facebook"`
	Status       string     `json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// UserListItem 用户列表项
type UserListItem struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"displayName"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	ProfileImage string     `json:"profileImage"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// UserFilter 用户过滤器
type UserFilter struct {
	Role     string `json:"role"`
	Status   string `json:"status"`
	Keyword  string `json:"keyword"` // 搜索用户名、邮箱、显示名
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	SortBy   string `json:"sortBy"`   // created_at, last_login_at, username
	SortDesc bool   `json:"sortDesc"` // 是否降序
}

// AuthorInfo 作者信息（用于文章显示）
type AuthorInfo struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"displayName"`
	ProfileImage string `json:"profileImage"`
	Bio          string `json:"bio"`
}

// ===============================
// 验证方法
// ===============================

// ValidateForCreate 验证用户创建数据
func (u *User) ValidateForCreate() error {
	// 验证必填字段
	if u.Username == "" {
		return NewValidationError("username", "用户名不能为空")
	}
	if u.Email == "" {
		return NewValidationError("email", "邮箱不能为空")
	}
	if u.DisplayName == "" {
		return NewValidationError("displayName", "显示名不能为空")
	}
	if u.Role == "" {
		return NewValidationError("role", "角色不能为空")
	}

	// 验证字段长度
	if len(u.Username) < constants.UsernameMinLength {
		return NewValidationError("username", "用户名长度不能少于3字符")
	}
	if len(u.Username) > constants.UsernameMaxLength {
		return NewValidationError("username", "用户名长度不能超过32字符")
	}
	if len(u.DisplayName) > constants.DisplayNameMaxLength {
		return NewValidationError("displayName", "显示名长度不能超过64字符")
	}
	if len(u.Bio) > constants.BioMaxLength {
		return NewValidationError("bio", "简介长度不能超过500字符")
	}
	if len(u.Location) > constants.LocationMaxLength {
		return NewValidationError("location", "地址长度不能超过100字符")
	}
	if len(u.Website) > constants.WebsiteMaxLength {
		return NewValidationError("website", "网站URL长度不能超过255字符")
	}
	if len(u.Twitter) > constants.SocialAccountMaxLength {
		return NewValidationError("twitter", "Twitter账号长度不能超过50字符")
	}
	if len(u.Facebook) > constants.SocialAccountMaxLength {
		return NewValidationError("facebook", "Facebook账号长度不能超过50字符")
	}

	// 验证用户角色
	if !constants.IsValidUserRole(u.Role) {
		return NewValidationError("role", "无效的用户角色")
	}

	// 验证用户状态
	if u.Status != "" && !constants.IsValidUserStatus(u.Status) {
		return NewValidationError("status", "无效的用户状态")
	}

	return nil
}

// ValidateForUpdate 验证用户更新数据
func (u *User) ValidateForUpdate() error {
	// 对于更新，只验证非空字段
	if u.DisplayName != "" && len(u.DisplayName) > constants.DisplayNameMaxLength {
		return NewValidationError("displayName", "显示名长度不能超过64字符")
	}
	if len(u.Bio) > constants.BioMaxLength {
		return NewValidationError("bio", "简介长度不能超过500字符")
	}
	if len(u.Location) > constants.LocationMaxLength {
		return NewValidationError("location", "地址长度不能超过100字符")
	}
	if len(u.Website) > constants.WebsiteMaxLength {
		return NewValidationError("website", "网站URL长度不能超过255字符")
	}
	if len(u.Twitter) > constants.SocialAccountMaxLength {
		return NewValidationError("twitter", "Twitter账号长度不能超过50字符")
	}
	if len(u.Facebook) > constants.SocialAccountMaxLength {
		return NewValidationError("facebook", "Facebook账号长度不能超过50字符")
	}

	if u.Role != "" && !constants.IsValidUserRole(u.Role) {
		return NewValidationError("role", "无效的用户角色")
	}
	if u.Status != "" && !constants.IsValidUserStatus(u.Status) {
		return NewValidationError("status", "无效的用户状态")
	}

	return nil
}

// ===============================
// 状态检查方法
// ===============================

// IsActive 检查用户是否为活跃状态
func (u *User) IsActive() bool {
	return u.Status == constants.UserStatusActive
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	if u.Status == constants.UserStatusLocked {
		return true
	}

	// 检查是否有临时锁定
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return true
	}

	return false
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsLocked()
}

// IsOwner 检查是否为所有者
func (u *User) IsOwner() bool {
	return u.Role == constants.UserRoleOwner
}

// IsAdmin 检查是否为管理员或更高权限
func (u *User) IsAdmin() bool {
	return u.Role == constants.UserRoleOwner || u.Role == constants.UserRoleAdmin
}

// IsEditor 检查是否为编辑或更高权限
func (u *User) IsEditor() bool {
	return u.IsAdmin() || u.Role == constants.UserRoleEditor
}

// CanManageUser 检查是否可以管理用户
func (u *User) CanManageUser() bool {
	return u.IsAdmin()
}

// CanManageAllPosts 检查是否可以管理所有文章
func (u *User) CanManageAllPosts() bool {
	return u.IsEditor()
}

// CanManageComments 检查是否可以管理评论
func (u *User) CanManageComments() bool {
	return u.IsEditor()
}

// ===============================
// 转换方法
// ===============================

// ToProfileResponse 转换为用户档案响应
func (u *User) ToProfileResponse() *UserProfileResponse {
	return &UserProfileResponse{
		ID:           u.ID.Hex(),
		Username:     u.Username,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		Role:         u.Role,
		ProfileImage: u.ProfileImage,
		CoverImage:   u.CoverImage,
		Bio:          u.Bio,
		Location:     u.Location,
		Website:      u.Website,
		Twitter:      u.Twitter,
		Facebook:     u.Facebook,
		Status:       u.Status,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// ToListItem 转换为用户列表项
func (u *User) ToListItem() *UserListItem {
	return &UserListItem{
		ID:           u.ID.Hex(),
		Username:     u.Username,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		Role:         u.Role,
		Status:       u.Status,
		ProfileImage: u.ProfileImage,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
	}
}

// ToAuthorInfo 转换为作者信息
func (u *User) ToAuthorInfo() *AuthorInfo {
	return &AuthorInfo{
		ID:           u.ID.Hex(),
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		ProfileImage: u.ProfileImage,
		Bio:          u.Bio,
	}
}

// ===============================
// 工厂方法
// ===============================

// NewUser 创建新用户
func NewUser(username, email, passwordHash, displayName, role string) *User {
	now := time.Now()

	user := &User{
		ID:             primitive.NewObjectID(),
		Username:       username,
		Email:          email,
		PasswordHash:   passwordHash,
		DisplayName:    displayName,
		Role:           role,
		Status:         constants.UserStatusActive,
		LoginFailCount: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return user
}

// NewUserFromCreateRequest 从创建请求创建用户
func NewUserFromCreateRequest(req *UserCreateRequest, passwordHash string) *User {
	user := NewUser(req.Username, req.Email, passwordHash, req.DisplayName, req.Role)

	// 设置可选字段
	user.Bio = req.Bio
	user.Location = req.Location
	user.Website = req.Website
	user.Twitter = req.Twitter
	user.Facebook = req.Facebook

	return user
}

// ===============================
// 数据库操作辅助方法
// ===============================

// PrepareForInsert 准备插入数据库
func (u *User) PrepareForInsert() {
	now := time.Now()
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	u.CreatedAt = now
	u.UpdatedAt = now

	// 设置默认状态
	if u.Status == "" {
		u.Status = constants.UserStatusActive
	}
}

// PrepareForUpdate 准备更新数据库
func (u *User) PrepareForUpdate() {
	u.UpdatedAt = time.Now()
}

// IncrementLoginFailCount 增加登录失败次数
func (u *User) IncrementLoginFailCount() {
	u.LoginFailCount++
	u.UpdatedAt = time.Now()

	// 根据失败次数设置锁定时间
	lockDuration := constants.GetLockDurationByFailCount(u.LoginFailCount)
	if lockDuration > 0 {
		lockUntil := time.Now().Add(time.Duration(lockDuration) * time.Minute)
		u.LockedUntil = &lockUntil
		u.Status = constants.UserStatusLocked
	}
}

// ResetLoginFailCount 重置登录失败次数
func (u *User) ResetLoginFailCount() {
	u.LoginFailCount = 0
	u.LockedUntil = nil
	if u.Status == constants.UserStatusLocked {
		u.Status = constants.UserStatusActive
	}
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin 更新最后登录信息
func (u *User) UpdateLastLogin(ipAddress string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = ipAddress
	u.UpdatedAt = now

	// 登录成功时重置失败计数
	u.ResetLoginFailCount()
}

// ===============================
// 验证错误类型
// ===============================

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
