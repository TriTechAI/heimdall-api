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

type LockAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 锁定用户账户
func NewLockAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockAccountLogic {
	return &LockAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LockAccountLogic) LockAccount(req *types.LockAccountRequest) (resp *types.LockAccountResponse, err error) {
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

	// 4. 检查是否可以锁定该用户
	if err := l.checkLockPermission(targetUser); err != nil {
		return nil, err
	}

	// 5. 计算锁定截止时间
	lockedUntil := l.calculateLockTime(req.LockDuration)

	// 6. 执行锁定操作
	if err := l.performLock(req.UserID, lockedUntil); err != nil {
		return nil, err
	}

	// 7. 构造响应
	resp = &types.LockAccountResponse{
		Code:      200,
		Message:   "账户锁定成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("账户锁定成功: userID=%s, duration=%d分钟, reason=%s", 
		req.UserID, req.LockDuration, req.LockReason)
	return resp, nil
}

// checkPermission 检查锁定权限
func (l *LockAccountLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以锁定用户
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以锁定用户")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *LockAccountLogic) validateRequest(req *types.LockAccountRequest) error {
	if req.UserID == "" {
		return errors.New("用户ID不能为空")
	}
	if req.LockDuration <= 0 {
		return errors.New("锁定时长必须大于0")
	}
	if req.LockDuration > 525600 { // 1年 = 365 * 24 * 60
		return errors.New("锁定时长不能超过1年")
	}
	return nil
}

// getTargetUser 获取目标用户
func (l *LockAccountLogic) getTargetUser(userID string) (*model.User, error) {
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

// checkLockPermission 检查是否可以锁定该用户
func (l *LockAccountLogic) checkLockPermission(targetUser *model.User) error {
	// 获取当前操作用户ID
	currentUserID := l.ctx.Value("uid")
	if currentUserID == nil {
		return errors.New("无法获取当前用户信息")
	}

	currentUID, ok := currentUserID.(string)
	if !ok {
		return errors.New("当前用户ID无效")
	}

	// 不能锁定自己
	if currentUID == targetUser.ID.Hex() {
		return errors.New("不能锁定自己的账户")
	}

	// 不能锁定其他管理员
	if targetUser.Role == "admin" {
		return errors.New("不能锁定管理员账户")
	}

	// 检查用户是否已经被锁定
	if targetUser.IsLocked() {
		return errors.New("用户账户已经被锁定")
	}

	return nil
}

// calculateLockTime 计算锁定截止时间
func (l *LockAccountLogic) calculateLockTime(durationMinutes int) time.Time {
	return time.Now().Add(time.Duration(durationMinutes) * time.Minute)
}

// performLock 执行锁定操作
func (l *LockAccountLogic) performLock(userID string, lockedUntil time.Time) error {
	if err := l.svcCtx.UserDAO.LockUser(l.ctx, userID, lockedUntil); err != nil {
		l.Logger.Errorf("锁定用户失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}
	return nil
}
