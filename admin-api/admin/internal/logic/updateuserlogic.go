package logic

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户信息
func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UserUpdateRequest) (resp *types.UserUpdateResponse, err error) {
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

	// 4. 检查唯一性约束
	if err := l.checkUniqueness(req, targetUser); err != nil {
		return nil, err
	}

	// 5. 构建更新数据
	updates, err := l.buildUpdateData(req)
	if err != nil {
		return nil, err
	}

	// 6. 执行更新操作
	if err := l.performUpdate(req.ID, updates); err != nil {
		return nil, err
	}

	// 7. 获取更新后的用户信息
	updatedUser, err := l.getUpdatedUser(req.ID)
	if err != nil {
		return nil, err
	}

	// 8. 构造响应
	resp = &types.UserUpdateResponse{
		Code:      200,
		Message:   "用户信息更新成功",
		Data:      l.buildUserInfo(updatedUser),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户信息更新成功: userID=%s", req.ID)
	return resp, nil
}

// checkPermission 检查更新权限
func (l *UpdateUserLogic) checkPermission() error {
	role := l.ctx.Value("role")
	if role == nil {
		return errors.New("用户未认证")
	}

	userRole, ok := role.(string)
	if !ok {
		return errors.New("用户角色无效")
	}

	// 只有管理员可以更新用户信息
	if userRole != "admin" {
		return errors.New("权限不足，只有管理员可以更新用户信息")
	}

	return nil
}

// validateRequest 验证请求参数
func (l *UpdateUserLogic) validateRequest(req *types.UserUpdateRequest) error {
	if req.ID == "" {
		return errors.New("用户ID不能为空")
	}

	// 验证邮箱格式（如果提供）
	if req.Email != "" {
		if err := l.validateEmail(req.Email); err != nil {
			return err
		}
	}

	// 验证显示名称（如果提供）
	if req.DisplayName != "" {
		if len(req.DisplayName) > 100 {
			return errors.New("显示名称长度不能超过100个字符")
		}
	}

	// 验证简介（如果提供）
	if len(req.Bio) > 500 {
		return errors.New("简介长度不能超过500个字符")
	}

	return nil
}

// validateEmail 验证邮箱格式
func (l *UpdateUserLogic) validateEmail(email string) error {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		return errors.New("邮箱格式无效")
	}
	return nil
}

// getTargetUser 获取目标用户
func (l *UpdateUserLogic) getTargetUser(userID string) (*model.User, error) {
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

// checkUniqueness 检查唯一性约束
func (l *UpdateUserLogic) checkUniqueness(req *types.UserUpdateRequest, currentUser *model.User) error {
	// 检查邮箱唯一性（如果要更新邮箱）
	if req.Email != "" && req.Email != currentUser.Email {
		existingUser, err := l.svcCtx.UserDAO.GetByEmail(l.ctx, req.Email)
		if err != nil {
			l.Logger.Errorf("检查邮箱唯一性失败: %v", err)
			return errors.New("系统错误，请稍后重试")
		}
		if existingUser != nil {
			return errors.New("邮箱已被其他用户使用")
		}
	}

	return nil
}

// buildUpdateData 构建更新数据
func (l *UpdateUserLogic) buildUpdateData(req *types.UserUpdateRequest) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// 基础信息更新
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.DisplayName != "" {
		updates["displayName"] = req.DisplayName
	}
	if req.ProfileImage != "" {
		updates["profileImage"] = req.ProfileImage
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}
	if req.Twitter != "" {
		updates["twitter"] = req.Twitter
	}
	if req.Facebook != "" {
		updates["facebook"] = req.Facebook
	}

	// 添加更新时间
	updates["updatedAt"] = time.Now()

	if len(updates) == 1 { // 只有updatedAt，说明没有实际更新
		return nil, errors.New("没有提供要更新的字段")
	}

	return updates, nil
}

// performUpdate 执行更新操作
func (l *UpdateUserLogic) performUpdate(userID string, updates map[string]interface{}) error {
	if err := l.svcCtx.UserDAO.Update(l.ctx, userID, updates); err != nil {
		l.Logger.Errorf("更新用户信息失败: userID=%s, error=%v", userID, err)
		return errors.New("系统错误，请稍后重试")
	}
	return nil
}

// getUpdatedUser 获取更新后的用户信息
func (l *UpdateUserLogic) getUpdatedUser(userID string) (*model.User, error) {
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		l.Logger.Errorf("获取更新后的用户信息失败: userID=%s, error=%v", userID, err)
		return nil, errors.New("系统错误，请稍后重试")
	}
	return user, nil
}

// buildUserInfo 构建用户信息响应
func (l *UpdateUserLogic) buildUserInfo(user *model.User) types.UserInfo {
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
