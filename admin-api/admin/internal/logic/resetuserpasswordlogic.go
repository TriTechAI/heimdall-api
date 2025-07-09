package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetUserPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重置用户密码
func NewResetUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserPasswordLogic {
	return &ResetUserPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetUserPasswordLogic) ResetUserPassword(req *types.ResetUserPasswordRequest) (resp *types.ResetUserPasswordResponse, err error) {
	// 1. 权限检查
	if err := l.checkPermission(); err != nil {
		return nil, err
	}

	// 2. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 3. 检查目标用户是否存在
	targetUser, err := l.getTargetUser(req.ID)
	if err != nil {
		return nil, err
	}

	// 4. 检查是否可以重置该用户密码
	if err := l.checkResetPermission(targetUser); err != nil {
		return nil, err
	}

	// 5. 生成新密码或使用提供的密码
	newPassword, err := l.generateOrValidatePassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	// 6. 执行密码重置
	if err := l.performPasswordReset(req.ID, newPassword); err != nil {
		return nil, err
	}

	// 7. 重置相关安全信息
	l.resetSecurityInfo(req.ID)

	// 8. 构造响应
	resp = &types.ResetUserPasswordResponse{
		Code:      200,
		Message:   "密码重置成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户密码重置成功: userID=%s", req.ID)
	return resp, nil
}

// checkPermission 检查密码重置权限
func (l *ResetUserPasswordLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以重置用户密码
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以重置用户密码")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *ResetUserPasswordLogic) validateRequest(req *types.ResetUserPasswordRequest) error {
	if req.ID == "" {
		return errors.New("用户ID不能为空")
	}

	// 如果提供了新密码，验证密码强度
	if req.NewPassword != "" {
		if err := l.validatePasswordStrength(req.NewPassword); err != nil {
			return err
		}
	}

	return nil
}

// validatePasswordStrength 验证密码强度
func (l *ResetUserPasswordLogic) validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少为8位")
	}
	if len(password) > 50 {
		return errors.New("密码长度不能超过50位")
	}
	return nil
}

// getTargetUser 获取目标用户
func (l *ResetUserPasswordLogic) getTargetUser(userID string) (*model.User, error) {
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		l.Logger.Errorf("获取用户信息失败: userID=%s, error=%v", userID, err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	return user, nil
}

// checkResetPermission 检查是否可以重置该用户密码
func (l *ResetUserPasswordLogic) checkResetPermission(targetUser *model.User) error {
	// 获取当前操作用户ID
	currentUserID := l.ctx.Value("uid")
	if currentUserID == nil {
		return errors.New("无法获取当前用户信息")
	}

	currentUID, ok := currentUserID.(string)
	if !ok {
		return errors.New("当前用户ID无效")
	}

	// 不能重置自己的密码（应该使用修改密码功能）
	if currentUID == targetUser.ID.Hex() {
		return errors.New("不能重置自己的密码，请使用修改密码功能")
	}

	// 不能重置其他管理员的密码
	if targetUser.Role == "admin" {
		return errors.New("不能重置管理员账户密码")
	}

	return nil
}

// generateOrValidatePassword 生成新密码或验证提供的密码
func (l *ResetUserPasswordLogic) generateOrValidatePassword(providedPassword string) (string, error) {
	if providedPassword != "" {
		// 使用提供的密码
		return providedPassword, nil
	}

	// 自动生成强密码
	return l.generateStrongPassword()
}

// generateStrongPassword 生成强密码
func (l *ResetUserPasswordLogic) generateStrongPassword() (string, error) {
	const (
		length      = 12
		lowerChars  = "abcdefghijklmnopqrstuvwxyz"
		upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digitChars  = "0123456789"
		symbolChars = "!@#$%^&*"
	)

	// 确保密码包含各种字符类型
	password := make([]byte, 0, length)

	// 至少一个小写字母
	if char, err := l.randomChar(lowerChars); err != nil {
		return "", err
	} else {
		password = append(password, char)
	}

	// 至少一个大写字母
	if char, err := l.randomChar(upperChars); err != nil {
		return "", err
	} else {
		password = append(password, char)
	}

	// 至少一个数字
	if char, err := l.randomChar(digitChars); err != nil {
		return "", err
	} else {
		password = append(password, char)
	}

	// 至少一个特殊字符
	if char, err := l.randomChar(symbolChars); err != nil {
		return "", err
	} else {
		password = append(password, char)
	}

	// 填充剩余长度
	allChars := lowerChars + upperChars + digitChars + symbolChars
	for len(password) < length {
		if char, err := l.randomChar(allChars); err != nil {
			return "", err
		} else {
			password = append(password, char)
		}
	}

	// 打乱密码字符顺序
	for i := len(password) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}

// randomChar 从字符集中随机选择一个字符
func (l *ResetUserPasswordLogic) randomChar(charset string) (byte, error) {
	max := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// performPasswordReset 执行密码重置
func (l *ResetUserPasswordLogic) performPasswordReset(userID, newPassword string) error {
	// 生成密码哈希
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		l.Logger.Errorf("生成密码哈希失败: %v", err)
		return errors.New("系统错误，请稍后重试")
	}

	// 更新密码
	updates := map[string]interface{}{
		"passwordHash": hashedPassword,
		"updatedAt":    time.Now(),
	}

	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("重置用户密码失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}

	return nil
}

// resetSecurityInfo 重置相关安全信息
func (l *ResetUserPasswordLogic) resetSecurityInfo(userID string) {
	// 重置登录失败计数
	securityUpdates := map[string]interface{}{
		"loginFailCount":  0,
		"lastFailedLogin": nil,
	}

	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, securityUpdates); err != nil {
		l.Logger.Errorf("重置安全信息失败: userID=%s, error=%v", userID, err)
		// 不返回错误，因为密码重置已经成功了
	}

	// 清除Redis中的登录失败计数
	key := "login_fail:" + userID
	if err := l.svcCtx.Redis.Del(l.ctx, key).Err(); err != nil {
		l.Logger.Errorf("清除Redis登录失败计数失败: userID=%s, error=%v", userID, err)
		// 不返回错误，因为密码重置已经成功了
	}
}
