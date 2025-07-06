package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/cache"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 刷新Token
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	// 1. 验证refresh token
	userInfo, err := l.validateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 检查用户状态
	user, err := l.validateUser(userInfo.UserID)
	if err != nil {
		return nil, err
	}

	// 3. 将旧的refresh token加入黑名单
	if err := l.blacklistOldToken(req.RefreshToken); err != nil {
		l.Logger.Errorf("将旧refresh token加入黑名单失败: %v", err)
		// 继续流程，不中断
	}

	// 4. 生成新的token对
	tokens, err := l.generateNewTokens(user)
	if err != nil {
		return nil, err
	}

	// 5. 更新用户会话
	l.updateUserSession(user.ID.Hex(), tokens.Token)

	// 6. 构造成功响应
	resp = &types.RefreshTokenResponse{
		Code:      200,
		Message:   "Token刷新成功",
		Data:      *tokens,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("Token刷新成功: userID=%s", user.ID.Hex())
	return resp, nil
}

// TokenInfo 从token解析的用户信息
type TokenInfo struct {
	UserID   string
	Username string
	Role     string
	TokenID  string
}

// validateRefreshToken 验证refresh token
func (l *RefreshTokenLogic) validateRefreshToken(tokenString string) (*TokenInfo, error) {
	if tokenString == "" {
		return nil, errors.New("refresh token不能为空")
	}

	// 验证token格式和签名
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	claims, err := jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("refresh token无效: %w", err)
	}

	// 检查token是否在黑名单中
	tokenID, err := jwtManager.ExtractTokenIDFromToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("提取token ID失败: %w", err)
	}

	blacklistCache := cache.NewJWTBlacklistCache(l.svcCtx.Redis, "jwt:blacklist:")
	isBlacklisted, err := blacklistCache.IsTokenBlacklisted(l.ctx, tokenID)
	if err != nil {
		l.Logger.Errorf("检查token黑名单状态失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	if isBlacklisted {
		return nil, errors.New("refresh token已失效")
	}

	return &TokenInfo{
		UserID:   claims.Subject,
		Username: claims.Username,
		Role:     claims.Role,
		TokenID:  tokenID,
	}, nil
}

// validateUser 验证用户状态
func (l *RefreshTokenLogic) validateUser(userID string) (*model.User, error) {
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

	if user.IsLocked() {
		if user.LockedUntil != nil {
			remainingMinutes := int(time.Until(*user.LockedUntil).Minutes())
			if remainingMinutes > 0 {
				return nil, fmt.Errorf("账户已被锁定，还需等待%d分钟", remainingMinutes)
			}
		}
		return nil, errors.New("账户已被锁定")
	}

	return user, nil
}

// blacklistOldToken 将旧token加入黑名单
func (l *RefreshTokenLogic) blacklistOldToken(tokenString string) error {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	
	tokenID, err := jwtManager.ExtractTokenIDFromToken(tokenString)
	if err != nil {
		return fmt.Errorf("提取token ID失败: %w", err)
	}

	remainingTime, err := jwtManager.GetTokenRemainingTime(tokenString)
	if err != nil {
		if errors.Is(err, utils.ErrTokenExpired) {
			return nil // token已过期，无需加入黑名单
		}
		return fmt.Errorf("获取token剩余时间失败: %w", err)
	}

	blacklistCache := cache.NewJWTBlacklistCache(l.svcCtx.Redis, "jwt:blacklist:")
	return blacklistCache.AddToken(l.ctx, tokenID, remainingTime)
}

// generateNewTokens 生成新的token对
func (l *RefreshTokenLogic) generateNewTokens(user *model.User) (*types.LoginData, error) {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")

	// 生成访问令牌
	accessToken, err := jwtManager.GenerateGoZeroCompatibleToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		l.Logger.Errorf("生成访问令牌失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 生成刷新令牌
	tokenPair, err := jwtManager.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		l.Logger.Errorf("生成刷新令牌失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	return &types.LoginData{
		Token:        accessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
		User: types.UserInfo{
			ID:          user.ID.Hex(),
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        user.Role,
			Status:      user.Status,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// updateUserSession 更新用户会话
func (l *RefreshTokenLogic) updateUserSession(userID, accessToken string) {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	tokenID, err := jwtManager.ExtractTokenIDFromToken(accessToken)
	if err != nil {
		l.Logger.Errorf("提取token ID失败: %v", err)
		return
	}
	// 使用tokenID作为sessionID
	sessionID := tokenID

	sessionCache := cache.NewSessionCache(l.svcCtx.Redis, "session:")
	sessionData, err := sessionCache.GetSession(l.ctx, sessionID)
	if err != nil {
		l.Logger.Errorf("获取用户会话失败: %v", err)
		return
	}

	if sessionData != nil {
		// 更新会话最后活动时间
		if err := sessionCache.TouchSession(l.ctx, sessionID); err != nil {
			l.Logger.Errorf("更新会话活动时间失败: %v", err)
		}
	}
}
