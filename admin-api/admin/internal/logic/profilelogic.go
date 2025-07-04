package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户信息
func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile() (resp *types.ProfileResponse, err error) {
	// 1. 从context获取用户ID
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 2. 验证用户ID格式
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		l.Logger.Errorf("用户ID格式无效: %s, error: %v", userID, err)
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
		l.Logger.Errorf("用户不存在: %s", userID)
		return nil, errors.New("用户不存在")
	}

	// 5. 检查用户状态
	if err := l.checkUserStatus(user); err != nil {
		return nil, err
	}

	// 6. 构造响应
	resp = &types.ProfileResponse{
		Code:      200,
		Message:   "获取用户信息成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      l.buildUserInfo(user),
	}

	return resp, nil
}

// getUserIDFromContext 从context获取用户ID
func (l *ProfileLogic) getUserIDFromContext() (string, error) {
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

// checkUserStatus 检查用户状态
func (l *ProfileLogic) checkUserStatus(user *model.User) error {
	// 检查用户是否为活跃状态
	if !user.IsActive() {
		switch user.Status {
		case constants.UserStatusInactive:
			return errors.New("账户已被禁用")
		case constants.UserStatusLocked:
			return errors.New("账户已被锁定")
		case constants.UserStatusSuspended:
			return errors.New("账户已被暂停")
		default:
			return errors.New("账户状态异常")
		}
	}

	// 检查用户是否临时锁定
	if user.IsLocked() {
		if user.LockedUntil != nil {
			return fmt.Errorf("账户已被锁定，解锁时间：%s", user.LockedUntil.Format("2006-01-02 15:04:05"))
		}
		return errors.New("账户已被锁定")
	}

	return nil
}

// buildUserInfo 构建用户信息响应
func (l *ProfileLogic) buildUserInfo(user *model.User) types.UserInfo {
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
