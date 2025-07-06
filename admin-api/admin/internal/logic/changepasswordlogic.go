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

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改密码
func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (resp *types.ChangePasswordResponse, err error) {
	// 1. 获取当前用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 2. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 3. 获取用户信息
	user, err := l.getUserInfo(userID)
	if err != nil {
		return nil, err
	}

	// 4. 验证当前密码
	if err := l.verifyCurrentPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return nil, err
	}

	// 5. 验证新密码强度
	if err := l.validatePasswordStrength(req.NewPassword); err != nil {
		return nil, err
	}

	// 6. 检查新密码与旧密码是否相同
	if err := l.checkPasswordDifference(user.PasswordHash, req.NewPassword); err != nil {
		return nil, err
	}

	// 7. 更新密码
	if err := l.updateUserPassword(userID, req.NewPassword); err != nil {
		return nil, err
	}

	// 8. 构造成功响应
	resp = &types.ChangePasswordResponse{
		Code:      200,
		Message:   "密码修改成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户密码修改成功: userID=%s", userID)
	return resp, nil
}

// getUserIDFromContext 从context获取用户ID
func (l *ChangePasswordLogic) getUserIDFromContext() (string, error) {
	uid := l.ctx.Value("uid")
	if uid == nil {
		return "", errors.New("用户未认证")
	}

	userID, ok := uid.(string)
	if !ok || userID == "" {
		return "", errors.New("用户ID无效")
	}

	return userID, nil
}

// validateRequest 验证请求参数
func (l *ChangePasswordLogic) validateRequest(req *types.ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return errors.New("当前密码不能为空")
	}
	if req.NewPassword == "" {
		return errors.New("新密码不能为空")
	}
	if req.ConfirmPassword == "" {
		return errors.New("确认密码不能为空")
	}
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("新密码与确认密码不一致")
	}
	return nil
}

// getUserInfo 获取用户信息
func (l *ChangePasswordLogic) getUserInfo(userID string) (*model.User, error) {
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		l.Logger.Errorf("获取用户信息失败: userID=%s, error=%v", userID, err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, errors.New("账户已被禁用")
	}

	return user, nil
}

// verifyCurrentPassword 验证当前密码
func (l *ChangePasswordLogic) verifyCurrentPassword(hashedPassword, currentPassword string) error {
	if err := utils.VerifyPassword(currentPassword, hashedPassword); err != nil {
		l.Logger.Info("用户当前密码验证失败")
		return errors.New("当前密码错误")
	}
	return nil
}

// validatePasswordStrength 验证密码强度
func (l *ChangePasswordLogic) validatePasswordStrength(password string) error {
	// 基本长度检查
	if len(password) < 8 {
		return errors.New("新密码长度至少为8位")
	}
	if len(password) > 50 {
		return errors.New("新密码长度不能超过50位")
	}

	// 密码复杂度检查
	checks := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{regexp.MustCompile(`[a-z]`), "新密码必须包含小写字母"},
		{regexp.MustCompile(`[A-Z]`), "新密码必须包含大写字母"},
		{regexp.MustCompile(`[0-9]`), "新密码必须包含数字"},
		{regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>?~]`), "新密码必须包含特殊字符"},
	}

	for _, check := range checks {
		if !check.pattern.MatchString(password) {
			return errors.New(check.message)
		}
	}

	// 检查是否包含常见弱密码
	weakPasswords := []string{
		"password", "123456", "admin", "user", "guest",
		"qwerty", "abc123", "password123", "admin123",
	}

	passwordLower := password
	for _, weak := range weakPasswords {
		if passwordLower == weak {
			return errors.New("新密码不能使用常见弱密码")
		}
	}

	return nil
}

// checkPasswordDifference 检查新密码与旧密码是否相同
func (l *ChangePasswordLogic) checkPasswordDifference(hashedPassword, newPassword string) error {
	// 验证新密码是否与当前密码相同
	if err := utils.VerifyPassword(newPassword, hashedPassword); err == nil {
		return errors.New("新密码不能与当前密码相同")
	}
	return nil
}

// updateUserPassword 更新用户密码
func (l *ChangePasswordLogic) updateUserPassword(userID, newPassword string) error {
	// 生成新密码哈希
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		l.Logger.Errorf("生成密码哈希失败: %v", err)
		return errors.New("系统错误，请稍后重试")
	}

	// 更新用户密码
	updates := map[string]interface{}{
		"passwordHash": hashedPassword,
		"updatedAt":    time.Now(),
	}

	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("更新用户密码失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}

	return nil
}
