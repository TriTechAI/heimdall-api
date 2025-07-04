package logic

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户详情
func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserDetailLogic) GetUserDetail(req *types.UserDetailRequest) (resp *types.UserDetailResponse, err error) {
	// 1. 参数验证
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 验证用户ID格式
	objectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		l.Logger.Errorf("用户ID格式无效: %s, error: %v", req.ID, err)
		return nil, errors.New("用户ID格式无效")
	}

	// 3. 获取用户信息
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, objectID.Hex())
	if err != nil {
		l.Logger.Errorf("获取用户信息失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 4. 检查用户是否存在
	if user == nil {
		l.Logger.Errorf("用户不存在: %s", req.ID)
		return nil, errors.New("用户不存在")
	}

	// 5. 权限检查（可选：管理员可以查看所有用户，普通用户只能查看自己）
	if err := l.checkViewPermission(user); err != nil {
		return nil, err
	}

	// 6. 构造响应
	resp = &types.UserDetailResponse{
		Code:      200,
		Message:   "获取用户详情成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      l.buildUserInfo(user),
	}

	return resp, nil
}

// validateRequest 验证请求参数
func (l *GetUserDetailLogic) validateRequest(req *types.UserDetailRequest) error {
	if req == nil {
		return errors.New("请求参数不能为空")
	}

	if req.ID == "" {
		return errors.New("用户ID不能为空")
	}

	return nil
}

// checkViewPermission 检查查看权限
func (l *GetUserDetailLogic) checkViewPermission(user *model.User) error {
	// 获取当前操作用户ID
	currentUserID := l.getCurrentUserID()
	if currentUserID == "" {
		return errors.New("用户未认证")
	}

	// 如果查看自己的信息，直接允许
	if currentUserID == user.ID.Hex() {
		return nil
	}

	// 如果是管理员，允许查看所有用户信息
	currentUser, err := l.svcCtx.UserDAO.GetByID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.Errorf("获取当前用户信息失败: %v", err)
		return errors.New("权限验证失败")
	}

	if currentUser != nil && (currentUser.Role == "admin" || currentUser.Role == "owner") {
		return nil
	}

	// 其他情况拒绝访问
	return errors.New("权限不足")
}

// getCurrentUserID 从context获取当前用户ID
func (l *GetUserDetailLogic) getCurrentUserID() string {
	uid := l.ctx.Value("uid")
	if uid == nil {
		return ""
	}

	userID, ok := uid.(string)
	if !ok || userID == "" {
		return ""
	}

	return userID
}

// buildUserInfo 构建用户信息响应（复用ProfileLogic的逻辑）
func (l *GetUserDetailLogic) buildUserInfo(user *model.User) types.UserInfo {
	userInfo := types.UserInfo{
		ID:          user.ID.Hex(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}

	// 设置可选字段
	if user.ProfileImage != "" {
		userInfo.ProfileImage = user.ProfileImage
	}
	if user.Bio != "" {
		userInfo.Bio = user.Bio
	}
	if user.Location != "" {
		userInfo.Location = user.Location
	}
	if user.Website != "" {
		userInfo.Website = user.Website
	}
	if user.Twitter != "" {
		userInfo.Twitter = user.Twitter
	}
	if user.Facebook != "" {
		userInfo.Facebook = user.Facebook
	}
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	return userInfo
}
