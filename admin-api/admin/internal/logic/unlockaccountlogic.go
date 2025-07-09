package logic

import (
	"context"
	"errors"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解锁用户账户
func NewUnlockAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockAccountLogic {
	return &UnlockAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnlockAccountLogic) UnlockAccount(req *types.UnlockAccountRequest) (resp *types.UnlockAccountResponse, err error) {
	// 1. 权限检查
	if err := l.checkPermission(); err != nil {
		return nil, err
	}

	// 2. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 3. 检查目标用户是否存在
	targetUser, err := l.getTargetUser(req.UserID)
	if err != nil {
		return nil, err
	}

	// 4. 检查是否可以解锁该用户
	if err := l.checkUnlockPermission(targetUser); err != nil {
		return nil, err
	}

	// 5. 执行解锁操作
	if err := l.performUnlock(req.UserID); err != nil {
		return nil, err
	}

	// 6. 重置登录失败计数
	l.resetLoginFailureCount(req.UserID)

	// 7. 构造响应
	resp = &types.UnlockAccountResponse{
		Code:      200,
		Message:   "账户解锁成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("账户解锁成功: userID=%s", req.UserID)
	return resp, nil
}

// checkPermission 检查解锁权限
func (l *UnlockAccountLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以解锁用户
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以解锁用户")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *UnlockAccountLogic) validateRequest(req *types.UnlockAccountRequest) error {
	if req.UserID == "" {
		return errors.New("用户ID不能为空")
	}
	return nil
}

// getTargetUser 获取目标用户
func (l *UnlockAccountLogic) getTargetUser(userID string) (*model.User, error) {
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

// checkUnlockPermission 检查是否可以解锁该用户
func (l *UnlockAccountLogic) checkUnlockPermission(targetUser *model.User) error {
	// 获取当前操作用户ID
	currentUserID := l.ctx.Value("uid")
	if currentUserID == nil {
		return errors.New("无法获取当前用户信息")
	}

	currentUID, ok := currentUserID.(string)
	if !ok {
		return errors.New("当前用户ID无效")
	}

	// 不能解锁自己
	if currentUID == targetUser.ID.Hex() {
		return errors.New("不能解锁自己的账户")
	}

	// 不能解锁其他管理员
	if targetUser.Role == "admin" {
		return errors.New("不能解锁管理员账户")
	}

	// 检查用户是否已经被锁定
	if !targetUser.IsLocked() {
		return errors.New("用户账户未被锁定")
	}

	return nil
}

// performUnlock 执行解锁操作
func (l *UnlockAccountLogic) performUnlock(userID string) error {
	updates := map[string]interface{}{
		"lockedUntil":     nil,
		"loginFailCount":  0,
		"lastFailedLogin": nil,
		"updatedAt":       time.Now(),
	}

	// 如果用户状态是locked，同时恢复为active
	user, _ := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if user != nil && user.Status == "locked" {
		updates["status"] = "active"
	}

	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("解锁用户账户失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}

	return nil
}

// resetLoginFailureCount 重置登录失败计数（Redis中的计数）
func (l *UnlockAccountLogic) resetLoginFailureCount(userID string) {
	// 清除Redis中的登录失败计数
	key := "login_fail:" + userID
	if err := l.svcCtx.Redis.Del(l.ctx, key).Err(); err != nil {
		l.Logger.Errorf("清除Redis登录失败计数失败: userID=%s, error=%v", userID, err)
		// 这里不返回错误，因为这不是关键操作
	}
}
