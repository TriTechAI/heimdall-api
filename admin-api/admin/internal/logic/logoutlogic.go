package logic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/cache"
	"github.com/heimdall-api/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登出
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutRequest) (resp *types.LogoutResponse, err error) {
	// 1. 获取当前用户信息
	userID, err := l.getUserIDFromContext()
	if err != nil {
		return nil, err
	}

	// 2. 获取当前访问token
	accessToken, err := l.getAccessTokenFromContext()
	if err != nil {
		return nil, err
	}

	// 3. 将访问token加入黑名单
	if err := l.blacklistAccessToken(accessToken); err != nil {
		l.Logger.Errorf("将访问token加入黑名单失败: %v", err)
		return nil, errors.New("登出失败，请稍后重试")
	}

	// 4. 处理refresh token（如果提供）
	if req.RefreshToken != "" {
		if err := l.blacklistRefreshToken(req.RefreshToken); err != nil {
			l.Logger.Errorf("将刷新token加入黑名单失败: %v", err)
			// refresh token处理失败不影响整个登出流程
		}
	}

	// 5. 清除用户会话
	sessionID := l.extractSessionID(accessToken)
	if sessionID != "" {
		l.clearUserSession(userID, sessionID)
	}

	// 6. 构造成功响应
	resp = &types.LogoutResponse{
		Code:      200,
		Message:   "登出成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户登出成功: userID=%s", userID)
	return resp, nil
}

// getUserIDFromContext 从context获取用户ID
func (l *LogoutLogic) getUserIDFromContext() (string, error) {
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

// getAccessTokenFromContext 从context获取访问token
func (l *LogoutLogic) getAccessTokenFromContext() (string, error) {
	// 尝试从HTTP请求头获取token
	if req, ok := l.ctx.Value("httpRequest").(*http.Request); ok {
		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			return utils.ParseAuthHeader(authHeader)
		}
	}

	// 如果无法从请求头获取，尝试从context获取
	if token := l.ctx.Value("token"); token != nil {
		if tokenStr, ok := token.(string); ok && tokenStr != "" {
			return tokenStr, nil
		}
	}

	return "", errors.New("无法获取访问token")
}

// blacklistAccessToken 将访问token加入黑名单
func (l *LogoutLogic) blacklistAccessToken(tokenString string) error {
	// 提取token ID和剩余时间
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

	// 使用缓存服务加入黑名单
	blacklistCache := cache.NewJWTBlacklistCache(l.svcCtx.Redis, "jwt:blacklist:")
	return blacklistCache.AddToken(l.ctx, tokenID, remainingTime)
}

// blacklistRefreshToken 将刷新token加入黑名单
func (l *LogoutLogic) blacklistRefreshToken(tokenString string) error {
	// 提取token ID和剩余时间
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	tokenID, err := jwtManager.ExtractTokenIDFromToken(tokenString)
	if err != nil {
		return fmt.Errorf("提取refresh token ID失败: %w", err)
	}

	remainingTime, err := jwtManager.GetTokenRemainingTime(tokenString)
	if err != nil {
		if errors.Is(err, utils.ErrTokenExpired) {
			return nil // token已过期，无需加入黑名单
		}
		return fmt.Errorf("获取refresh token剩余时间失败: %w", err)
	}

	// 使用缓存服务加入黑名单
	blacklistCache := cache.NewJWTBlacklistCache(l.svcCtx.Redis, "jwt:blacklist:")
	return blacklistCache.AddToken(l.ctx, tokenID, remainingTime)
}

// extractSessionID 从token中提取会话ID
func (l *LogoutLogic) extractSessionID(tokenString string) string {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	tokenID, err := jwtManager.ExtractTokenIDFromToken(tokenString)
	if err != nil {
		l.Logger.Errorf("提取token ID失败: %v", err)
		return ""
	}
	// 使用tokenID作为sessionID
	sessionID := tokenID
	return sessionID
}

// clearUserSession 清除用户会话
func (l *LogoutLogic) clearUserSession(userID, sessionID string) {
	if sessionID == "" {
		return
	}

	sessionCache := cache.NewSessionCache(l.svcCtx.Redis, "session:")
	if err := sessionCache.DeleteSession(l.ctx, sessionID); err != nil {
		l.Logger.Errorf("删除用户会话失败: userID=%s, sessionID=%s, error=%v", userID, sessionID, err)
	}
}
