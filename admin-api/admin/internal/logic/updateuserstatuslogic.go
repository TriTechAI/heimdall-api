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

type UpdateUserStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户状态
func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserStatusLogic) UpdateUserStatus(req *types.UserStatusRequest) (resp *types.UserStatusResponse, err error) {
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

	// 4. 检查状态变更权限
	if err := l.checkStatusChangePermission(targetUser, req.Status); err != nil {
		return nil, err
	}

	// 5. 构建更新数据
	updates, err := l.buildStatusUpdate(req)
	if err != nil {
		return nil, err
	}

	// 6. 执行状态更新
	if err := l.performStatusUpdate(req.ID, updates); err != nil {
		return nil, err
	}

	// 7. 构造响应
	resp = &types.UserStatusResponse{
		Code:      200,
		Message:   "用户状态更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户状态更新成功: userID=%s, status=%s", req.ID, req.Status)
	return resp, nil
}

// checkPermission 检查状态更新权限
func (l *UpdateUserStatusLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以更新用户状态
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以更新用户状态")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *UpdateUserStatusLogic) validateRequest(req *types.UserStatusRequest) error {
	if req.ID == "" {
		return errors.New("用户ID不能为空")
	}

	if req.Status == "" {
		return errors.New("用户状态不能为空")
	}

	// 验证状态值的有效性
	if err := l.validateStatus(req.Status); err != nil {
		return err
	}

	return nil
}

// validateStatus 验证状态值
func (l *UpdateUserStatusLogic) validateStatus(status string) error {
	validStatuses := []string{"active", "inactive", "suspended"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return errors.New("无效的用户状态，有效状态为: active, inactive, suspended")
}

// getTargetUser 获取目标用户
func (l *UpdateUserStatusLogic) getTargetUser(userID string) (*model.User, error) {
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

// checkStatusChangePermission 检查状态变更权限
func (l *UpdateUserStatusLogic) checkStatusChangePermission(targetUser *model.User, newStatus string) error {
	// 获取当前操作用户ID
	currentUserID := l.ctx.Value("uid")
	if currentUserID == nil {
		return errors.New("无法获取当前用户信息")
	}

	currentUID, ok := currentUserID.(string)
	if !ok {
		return errors.New("当前用户ID无效")
	}

	// 不能修改自己的状态
	if currentUID == targetUser.ID.Hex() {
		return errors.New("不能修改自己的账户状态")
	}

	// 不能修改其他管理员的状态
	if targetUser.Role == "admin" {
		return errors.New("不能修改管理员账户状态")
	}

	// 检查状态是否需要变更
	if targetUser.Status == newStatus {
		return errors.New("用户状态无需变更")
	}

	// 特殊业务规则：如果用户被锁定，需要先解锁再修改状态
	if targetUser.IsLocked() && newStatus == "active" {
		return errors.New("用户账户被锁定，请先解锁后再激活")
	}

	return nil
}

// buildStatusUpdate 构建状态更新数据
func (l *UpdateUserStatusLogic) buildStatusUpdate(req *types.UserStatusRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{
		"status":    req.Status,
		"updatedAt": time.Now(),
	}

	// 如果有状态变更原因，记录下来
	if req.Reason != "" {
		updates["statusReason"] = req.Reason
	}

	// 如果状态变为inactive或suspended，记录变更时间
	if req.Status == "inactive" || req.Status == "suspended" {
		updates["statusChangedAt"] = time.Now()
	}

	// 如果状态变为active，清除之前的状态变更信息
	if req.Status == "active" {
		updates["statusReason"] = ""
		updates["statusChangedAt"] = nil
	}

	return updates, nil
}

// performStatusUpdate 执行状态更新
func (l *UpdateUserStatusLogic) performStatusUpdate(userID string, updates map[string]interface{}) error {
	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("更新用户状态失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}
	return nil
}
