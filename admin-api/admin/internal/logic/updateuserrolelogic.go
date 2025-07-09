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

type UpdateUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户角色
func NewUpdateUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleLogic {
	return &UpdateUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserRoleLogic) UpdateUserRole(req *types.UserRoleRequest) (resp *types.UserRoleResponse, err error) {
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

	// 4. 检查角色变更权限
	if err := l.checkRoleChangePermission(targetUser, req.Role); err != nil {
		return nil, err
	}

	// 5. 构建更新数据
	updates, err := l.buildRoleUpdate(req)
	if err != nil {
		return nil, err
	}

	// 6. 执行角色更新
	if err := l.performRoleUpdate(req.ID, updates); err != nil {
		return nil, err
	}

	// 7. 构造响应
	resp = &types.UserRoleResponse{
		Code:      200,
		Message:   "用户角色更新成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户角色更新成功: userID=%s, role=%s", req.ID, req.Role)
	return resp, nil
}

// checkPermission 检查角色更新权限
func (l *UpdateUserRoleLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以更新用户角色
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以更新用户角色")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *UpdateUserRoleLogic) validateRequest(req *types.UserRoleRequest) error {
	if req.ID == "" {
		return errors.New("用户ID不能为空")
	}

	if req.Role == "" {
		return errors.New("用户角色不能为空")
	}

	// 验证角色值的有效性
	if err := l.validateRole(req.Role); err != nil {
		return err
	}

	return nil
}

// validateRole 验证角色值
func (l *UpdateUserRoleLogic) validateRole(role string) error {
	validRoles := []string{"admin", "editor", "author"}
	for _, validRole := range validRoles {
		if role == validRole {
			return nil
		}
	}
	return errors.New("无效的用户角色，有效角色为: admin, editor, author")
}

// getTargetUser 获取目标用户
func (l *UpdateUserRoleLogic) getTargetUser(userID string) (*model.User, error) {
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

// checkRoleChangePermission 检查角色变更权限
func (l *UpdateUserRoleLogic) checkRoleChangePermission(targetUser *model.User, newRole string) error {
	// 获取当前操作用户ID
	currentUserID := l.ctx.Value("uid")
	if currentUserID == nil {
		return errors.New("无法获取当前用户信息")
	}

	currentUID, ok := currentUserID.(string)
	if !ok {
		return errors.New("当前用户ID无效")
	}

	// 不能修改自己的角色
	if currentUID == targetUser.ID.Hex() {
		return errors.New("不能修改自己的账户角色")
	}

	// 检查角色是否需要变更
	if targetUser.Role == newRole {
		return errors.New("用户角色无需变更")
	}

	// 特殊权限控制：管理员角色变更
	if targetUser.Role == "admin" {
		return errors.New("不能修改管理员账户角色")
	}

	// 特殊权限控制：不能将用户提升为管理员（需要更高权限）
	if newRole == "admin" {
		return errors.New("不能将用户提升为管理员角色")
	}

	// 特殊业务规则：被锁定的用户不能变更角色
	if targetUser.IsLocked() {
		return errors.New("用户账户被锁定，无法变更角色")
	}

	// 特殊业务规则：非活跃用户不能变更角色
	if !targetUser.IsActive() {
		return errors.New("用户账户未激活，无法变更角色")
	}

	return nil
}

// buildRoleUpdate 构建角色更新数据
func (l *UpdateUserRoleLogic) buildRoleUpdate(req *types.UserRoleRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{
		"role":      req.Role,
		"updatedAt": time.Now(),
	}

	// 如果有角色变更原因，记录下来
	if req.Reason != "" {
		updates["roleReason"] = req.Reason
	}

	// 记录角色变更时间
	updates["roleChangedAt"] = time.Now()

	return updates, nil
}

// performRoleUpdate 执行角色更新
func (l *UpdateUserRoleLogic) performRoleUpdate(userID string, updates map[string]interface{}) error {
	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("更新用户角色失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}
	return nil
}
