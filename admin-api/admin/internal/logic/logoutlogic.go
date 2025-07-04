package logic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
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
		l.Logger.Errorf("获取用户ID失败: %v", err)
		return nil, err
	}

	// 2. 获取当前访问token
	accessToken, err := l.getAccessTokenFromContext()
	if err != nil {
		l.Logger.Errorf("获取访问token失败: %v", err)
		return nil, err
	}

	// 3. 提取access token信息
	accessTokenID, err := l.extractTokenID(accessToken)
	if err != nil {
		l.Logger.Errorf("提取access token ID失败: %v", err)
		return nil, err
	}

	// 4. 将access token加入黑名单
	if err := l.addTokenToBlacklist(accessTokenID, accessToken); err != nil {
		l.Logger.Errorf("将access token加入黑名单失败: %v", err)
		return nil, err
	}

	// 5. 处理refresh token（如果提供）
	if req.RefreshToken != "" {
		if err := l.handleRefreshToken(req.RefreshToken); err != nil {
			l.Logger.Errorf("处理refresh token失败: %v", err)
			// refresh token处理失败不影响整个登出流程
		}
	}

	// 6. 清除用户会话缓存
	if err := l.clearUserSession(userID, accessTokenID); err != nil {
		l.Logger.Errorf("清除用户会话失败: %v", err)
		// 会话清理失败不影响登出
	}

	// 7. 记录登出日志
	go l.recordLogoutLog(userID, accessTokenID)

	// 8. 构造成功响应
	resp = &types.LogoutResponse{
		Code:      200,
		Message:   "登出成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	l.Logger.Infof("用户登出成功: userID=%s, tokenID=%s", userID, accessTokenID)
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
			token, err := utils.ParseAuthHeader(authHeader)
			if err != nil {
				return "", fmt.Errorf("解析Authorization头失败: %w", err)
			}
			return token, nil
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

// extractTokenID 提取token ID
func (l *LogoutLogic) extractTokenID(tokenString string) (string, error) {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	tokenID, err := jwtManager.ExtractTokenIDFromToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("提取token ID失败: %w", err)
	}
	return tokenID, nil
}

// addTokenToBlacklist 将token加入黑名单
func (l *LogoutLogic) addTokenToBlacklist(tokenID, tokenString string) error {
	// 计算token剩余有效时间
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	remainingTime, err := jwtManager.GetTokenRemainingTime(tokenString)
	if err != nil {
		// 如果token已过期，直接返回成功
		if errors.Is(err, utils.ErrTokenExpired) {
			return nil
		}
		return fmt.Errorf("获取token剩余时间失败: %w", err)
	}

	// 将token ID加入黑名单，过期时间与token剩余时间一致
	blacklistKey := utils.GenerateBlacklistKey(tokenID)
	return l.svcCtx.Redis.Set(l.ctx, blacklistKey, "1", remainingTime).Err()
}

// handleRefreshToken 处理refresh token
func (l *LogoutLogic) handleRefreshToken(refreshToken string) error {
	// 提取refresh token ID
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	refreshTokenID, err := jwtManager.ExtractTokenIDFromToken(refreshToken)
	if err != nil {
		return fmt.Errorf("提取refresh token ID失败: %w", err)
	}

	// 将refresh token加入黑名单
	remainingTime, err := jwtManager.GetTokenRemainingTime(refreshToken)
	if err != nil {
		if errors.Is(err, utils.ErrTokenExpired) {
			return nil
		}
		return fmt.Errorf("获取refresh token剩余时间失败: %w", err)
	}

	blacklistKey := utils.GenerateBlacklistKey(refreshTokenID)
	return l.svcCtx.Redis.Set(l.ctx, blacklistKey, "1", remainingTime).Err()
}

// clearUserSession 清除用户会话缓存
func (l *LogoutLogic) clearUserSession(userID, tokenID string) error {
	sessionKey := utils.GenerateSessionKey(userID, tokenID)
	return l.svcCtx.Redis.Del(l.ctx, sessionKey).Err()
}

// recordLogoutLog 记录登出日志
func (l *LogoutLogic) recordLogoutLog(userID, tokenID string) {
	// 转换userID为ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		l.Logger.Errorf("转换用户ID失败: userID=%s, error=%v", userID, err)
		return
	}

	// 创建登出日志
	logoutLog := &model.LoginLog{
		UserID:      &objectID,  // 使用ObjectID指针
		LoginMethod: "username", // 使用有效的登录方式
		IPAddress:   l.getClientIP(),
		UserAgent:   l.getUserAgent(),
		Status:      constants.LoginStatusSuccess,
		SessionID:   tokenID,
		LoginAt:     time.Now(),
		LogoutAt:    &[]time.Time{time.Now()}[0], // 设置登出时间
	}

	// 异步记录日志，不影响主流程
	if err := l.svcCtx.LoginLogDAO.Create(l.ctx, logoutLog); err != nil {
		l.Logger.Errorf("记录登出日志失败: userID=%s, error=%v", userID, err)
	}
}

// getClientIP 获取客户端IP
func (l *LogoutLogic) getClientIP() string {
	if req, ok := l.ctx.Value("httpRequest").(*http.Request); ok {
		// 尝试从代理头获取真实IP
		if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
			return forwarded
		}
		if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
			return realIP
		}
		return req.RemoteAddr
	}
	return "unknown"
}

// getUserAgent 获取用户代理
func (l *LogoutLogic) getUserAgent() string {
	if req, ok := l.ctx.Value("httpRequest").(*http.Request); ok {
		return req.Header.Get("User-Agent")
	}
	return "unknown"
}
