package logic

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建用户
func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.UserCreateRequest) (resp *types.UserCreateResponse, err error) {
	// 1. 权限检查
	if err := l.checkPermission(); err != nil {
		return nil, err
	}

	// 2. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 3. 检查用户名和邮箱唯一性
	if err := l.checkUniqueness(req.Username, req.Email); err != nil {
		return nil, err
	}

	// 4. 验证密码强度
	if err := l.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// 5. 创建用户对象
	user, err := l.buildUser(req)
	if err != nil {
		return nil, err
	}

	// 6. 保存用户到数据库
	if err := l.saveUser(user); err != nil {
		return nil, err
	}

	// 7. 构造响应
	resp = &types.UserCreateResponse{
		Code:      201,
		Message:   "用户创建成功",
		Data:      l.buildUserInfo(user),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户创建成功: username=%s, email=%s", req.Username, req.Email)
	return resp, nil
}

// checkPermission 检查创建用户权限
func (l *CreateUserLogic) checkPermission() error {
	// 获取当前用户角色
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以创建用户
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以创建用户")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *CreateUserLogic) validateRequest(req *types.UserCreateRequest) error {
	if req.Username == "" {
		return errors.New("用户名不能为空")
	}
	if req.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if req.Password == "" {
		return errors.New("密码不能为空")
	}
	if req.DisplayName == "" {
		return errors.New("显示名称不能为空")
	}
	if req.Role == "" {
		return errors.New("用户角色不能为空")
	}

	// 验证用户名格式
	if err := l.validateUsername(req.Username); err != nil {
		return err
	}

	// 验证邮箱格式
	if err := l.validateEmail(req.Email); err != nil {
		return err
	}

	// 验证角色有效性
	if err := l.validateRole(req.Role); err != nil {
		return err
	}

	return nil
}

// validateUsername 验证用户名格式
func (l *CreateUserLogic) validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return errors.New("用户名长度必须在3-50个字符之间")
	}

	// 用户名只能包含字母、数字和下划线
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return errors.New("用户名只能包含字母、数字和下划线")
	}

	return nil
}

// validateEmail 验证邮箱格式
func (l *CreateUserLogic) validateEmail(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		return errors.New("邮箱格式无效")
	}
	return nil
}

// validateRole 验证用户角色
func (l *CreateUserLogic) validateRole(role string) error {
	validRoles := []string{"admin", "editor", "author"}
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}
	return errors.New("无效的用户角色，有效角色为: admin, editor, author")
}

// checkUniqueness 检查用户名和邮箱唯一性
func (l *CreateUserLogic) checkUniqueness(username, email string) error {
	// 检查用户名是否已存在
	existingUser, err := l.svcCtx.UserDAO.GetByUsername(l.ctx, username)
	if err != nil {
		l.Logger.Errorf("检查用户名唯一性失败: %v", err)
		return errors.New("系统错误，请稍后重试")
	}
	if existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = l.svcCtx.UserDAO.GetByEmail(l.ctx, email)
	if err != nil {
		l.Logger.Errorf("检查邮箱唯一性失败: %v", err)
		return errors.New("系统错误，请稍后重试")
	}
	if existingUser != nil {
		return errors.New("邮箱已存在")
	}

	return nil
}

// validatePassword 验证密码强度
func (l *CreateUserLogic) validatePassword(password string) error {
	if len(password) < 8 || len(password) > 50 {
		return errors.New("密码长度必须在8-50个字符之间")
	}

	// 检查密码复杂度
	checks := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{regexp.MustCompile(`[a-z]`), "密码必须包含小写字母"},
		{regexp.MustCompile(`[A-Z]`), "密码必须包含大写字母"},
		{regexp.MustCompile(`[0-9]`), "密码必须包含数字"},
	}

	for _, check := range checks {
		if !check.pattern.MatchString(password) {
			return errors.New(check.message)
		}
	}

	return nil
}

// buildUser 构建用户对象
func (l *CreateUserLogic) buildUser(req *types.UserCreateRequest) (*model.User, error) {
	// 生成密码哈希
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		l.Logger.Errorf("生成密码哈希失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	now := time.Now()
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         req.Role,
		Status:       req.Status,
		ProfileImage: req.ProfileImage,
		Bio:          req.Bio,
		Location:     req.Location,
		Website:      req.Website,
		Twitter:      req.Twitter,
		Facebook:     req.Facebook,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 如果状态为空，设置默认值
	if user.Status == "" {
		user.Status = "active"
	}

	return user, nil
}

// saveUser 保存用户到数据库
func (l *CreateUserLogic) saveUser(user *model.User) error {
	if err := l.svcCtx.UserDAO.Create(l.ctx, user); err != nil {
		l.Logger.Errorf("保存用户失败: %v", err)
		return errors.New("系统错误，请稍后重试")
	}
	return nil
}

// buildUserInfo 构建用户信息响应
func (l *CreateUserLogic) buildUserInfo(user *model.User) types.UserInfo {
	return types.UserInfo{
		ID:           user.ID.Hex(),
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		Role:         user.Role,
		ProfileImage: user.ProfileImage,
		Bio:          user.Bio,
		Location:     user.Location,
		Website:      user.Website,
		Twitter:      user.Twitter,
		Facebook:     user.Facebook,
		Status:       user.Status,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
	}
}
