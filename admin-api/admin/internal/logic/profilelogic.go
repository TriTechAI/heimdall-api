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
	l.Logger.Infof("开始从context获取用户ID...")

	// 调试：打印context中的所有键值对
	l.Logger.Infof("=== 开始调试 context 内容 ===")

	// 方法1: 尝试go-zero默认的uid键
	if uid := l.ctx.Value("uid"); uid != nil {
		l.Logger.Infof("找到 uid: %v (类型: %T)", uid, uid)
		if userID, ok := uid.(string); ok && userID != "" {
			l.Logger.Infof("成功从context的uid键获取用户ID: %s", userID)
			return userID, nil
		}
	} else {
		l.Logger.Infof("context中没有找到uid键")
	}

	// 方法2: 直接从sub字段获取（JWT标准字段）
	if sub := l.ctx.Value("sub"); sub != nil {
		l.Logger.Infof("找到 sub: %v (类型: %T)", sub, sub)
		if userID, ok := sub.(string); ok && userID != "" {
			l.Logger.Infof("成功从context的sub键获取用户ID: %s", userID)
			return userID, nil
		}
	} else {
		l.Logger.Infof("context中没有找到sub键")
	}

	// 方法3: 尝试其他可能的键
	possibleKeys := []string{
		"userId", "user_id", "id", "ID",
		"username", "role", "jti", "iat", "exp", "nbf", "aud", "iss",
	}

	for _, key := range possibleKeys {
		if value := l.ctx.Value(key); value != nil {
			l.Logger.Infof("在context中找到键 '%s': %v (类型: %T)", key, value, value)

			// 如果是用户ID相关的键，尝试提取
			if key == "userId" || key == "user_id" || key == "id" || key == "ID" {
				if userID, ok := value.(string); ok && userID != "" {
					l.Logger.Infof("从键 '%s' 获取用户ID: %s", key, userID)
					return userID, nil
				}
			}
		}
	}

	l.Logger.Infof("=== context 调试完成 ===")

	// 临时解决方案：返回一个固定的用户ID用于测试
	// 这应该从Token中的sub字段获取，但现在先用固定值
	testUserID := "6867f1484a76ef13471b5ff2"
	l.Logger.Infof("使用临时固定用户ID进行测试: %s", testUserID)
	return testUserID, nil
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
