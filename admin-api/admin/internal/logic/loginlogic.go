package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// 1. 参数验证
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 获取客户端IP地址
	clientIP := l.getClientIP()

	// 3. 检查登录失败次数限制
	if err := l.checkLoginAttempts(req.Username, clientIP); err != nil {
		return nil, err
	}

	// 4. 获取用户信息
	user, err := l.svcCtx.UserDAO.GetByUsername(l.ctx, req.Username)
	if err != nil {
		l.Logger.Errorf("获取用户信息失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 5. 验证用户存在性和状态
	if user == nil {
		// 记录失败日志并增加失败次数
		l.recordLoginFailure(req.Username, clientIP, "用户不存在")
		l.incrementLoginAttempts(req.Username, clientIP)
		return nil, errors.New("用户名或密码错误")
	}

	// 6. 检查用户状态
	if err := l.checkUserStatus(user); err != nil {
		l.recordLoginFailure(user.Username, clientIP, err.Error())
		return nil, err
	}

	// 7. 验证密码
	if utils.VerifyPassword(req.Password, user.PasswordHash) != nil {
		// 记录失败日志并增加失败次数
		l.recordLoginFailure(user.Username, clientIP, "密码错误")
		l.incrementLoginAttempts(req.Username, clientIP)

		// 增加用户的登录失败计数
		if err := l.svcCtx.UserDAO.IncrementLoginFailCount(l.ctx, user.ID.Hex()); err != nil {
			l.Logger.Errorf("更新用户登录失败次数失败: %v", err)
		}

		// 检查是否需要锁定账户
		if err := l.checkAndLockAccount(user, req.Username, clientIP); err != nil {
			return nil, err
		}

		return nil, errors.New("用户名或密码错误")
	}

	// 8. 登录成功，生成JWT Token
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")
	tokenPair, err := jwtManager.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		l.Logger.Errorf("生成JWT Token失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 9. 清除登录失败次数缓存
	l.clearLoginAttempts(req.Username, clientIP)

	// 10. 更新用户登录信息
	if err := l.svcCtx.UserDAO.UpdateLoginInfo(l.ctx, user.ID.Hex(), clientIP); err != nil {
		l.Logger.Errorf("更新用户登录信息失败: %v", err)
		// 这个错误不阻止登录流程
	}

	// 11. 记录成功登录日志
	l.recordLoginSuccess(user, clientIP)

	// 12. 构造响应
	resp = &types.LoginResponse{
		Code:      200,
		Message:   "登录成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.LoginData{
			Token:        tokenPair.AccessToken,
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
		},
	}

	return resp, nil
}

// validateRequest 验证登录请求参数
func (l *LoginLogic) validateRequest(req *types.LoginRequest) error {
	if req == nil {
		return errors.New("登录请求不能为空")
	}
	if req.Username == "" {
		return errors.New("用户名不能为空")
	}
	if req.Password == "" {
		return errors.New("密码不能为空")
	}
	return nil
}

// getClientIP 获取客户端IP地址
func (l *LoginLogic) getClientIP() string {
	// 这里简化处理，实际应该从HTTP头中获取
	// 在go-zero中可以通过logx.FromContext获取
	return "127.0.0.1" // 临时固定值
}

// checkLoginAttempts 检查登录失败次数限制
func (l *LoginLogic) checkLoginAttempts(username, clientIP string) error {
	// 构造Redis键
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)

	// 获取当前失败次数
	result := l.svcCtx.Redis.Get(l.ctx, key)
	if result.Err() != nil && result.Err() != redis.Nil {
		l.Logger.Errorf("获取登录失败次数失败: %v", result.Err())
		return nil // 不因为Redis错误阻止登录
	}

	if result.Err() == redis.Nil {
		return nil // 没有失败记录
	}

	attempts, err := strconv.Atoi(result.Val())
	if err != nil {
		l.Logger.Errorf("解析登录失败次数失败: %v", err)
		return nil
	}

	// 检查是否超过最大失败次数
	if attempts >= l.svcCtx.Config.Security.MaxLoginAttempts {
		lockoutMinutes := l.svcCtx.Config.Security.LoginLockoutDuration / 60
		return fmt.Errorf("登录失败次数过多，请%d分钟后再试", lockoutMinutes)
	}

	return nil
}

// checkUserStatus 检查用户状态
func (l *LoginLogic) checkUserStatus(user *model.User) error {
	// 检查用户是否已被禁用
	if !user.IsActive() {
		return errors.New("账户已被禁用")
	}

	// 检查用户是否被锁定
	if user.IsLocked() {
		if user.LockedUntil != nil {
			remainingMinutes := int(time.Until(*user.LockedUntil).Minutes())
			if remainingMinutes > 0 {
				return fmt.Errorf("账户已被锁定，还需等待%d分钟", remainingMinutes)
			}
		}
		return errors.New("账户已被锁定")
	}

	return nil
}

// incrementLoginAttempts 增加登录失败次数
func (l *LoginLogic) incrementLoginAttempts(username, clientIP string) {
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)

	// 增加计数
	result := l.svcCtx.Redis.Incr(l.ctx, key)
	if result.Err() != nil {
		l.Logger.Errorf("增加登录失败次数失败: %v", result.Err())
		return
	}

	// 设置过期时间
	l.svcCtx.Redis.Expire(l.ctx, key, time.Duration(l.svcCtx.Config.Cache.LoginAttempts.TTL)*time.Second)
}

// clearLoginAttempts 清除登录失败次数
func (l *LoginLogic) clearLoginAttempts(username, clientIP string) {
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)

	if err := l.svcCtx.Redis.Del(l.ctx, key).Err(); err != nil {
		l.Logger.Errorf("清除登录失败次数失败: %v", err)
	}
}

// checkAndLockAccount 检查并锁定账户
func (l *LoginLogic) checkAndLockAccount(user *model.User, username, clientIP string) error {
	// 获取当前Redis中的失败次数
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)
	result := l.svcCtx.Redis.Get(l.ctx, key)

	var attempts int
	if result.Err() == nil {
		var err error
		attempts, err = strconv.Atoi(result.Val())
		if err != nil {
			attempts = 1
		}
	} else {
		attempts = 1
	}

	// 如果达到最大失败次数，锁定账户
	if attempts >= l.svcCtx.Config.Security.MaxLoginAttempts {
		lockDuration := time.Duration(l.svcCtx.Config.Security.LoginLockoutDuration) * time.Second
		lockUntil := time.Now().Add(lockDuration)

		if err := l.svcCtx.UserDAO.LockUser(l.ctx, user.ID.Hex(), lockUntil); err != nil {
			l.Logger.Errorf("锁定用户账户失败: %v", err)
		}

		lockoutMinutes := l.svcCtx.Config.Security.LoginLockoutDuration / 60
		return fmt.Errorf("登录失败次数过多，账户已被锁定%d分钟", lockoutMinutes)
	}

	return nil
}

// recordLoginFailure 记录登录失败日志
func (l *LoginLogic) recordLoginFailure(username, clientIP, reason string) {
	loginLog := &model.LoginLog{
		Username:   username,
		IPAddress:  clientIP,
		UserAgent:  "Admin Panel", // 简化处理
		Status:     constants.LoginStatusFailed,
		FailReason: reason,
		LoginAt:    time.Now(),
	}

	if err := l.svcCtx.LoginLogDAO.Create(l.ctx, loginLog); err != nil {
		l.Logger.Errorf("记录登录失败日志失败: %v", err)
	}
}

// recordLoginSuccess 记录登录成功日志
func (l *LoginLogic) recordLoginSuccess(user *model.User, clientIP string) {
	loginLog := &model.LoginLog{
		UserID:    &user.ID,
		Username:  user.Username,
		IPAddress: clientIP,
		UserAgent: "Admin Panel", // 简化处理
		Status:    constants.LoginStatusSuccess,
		LoginAt:   time.Now(),
	}

	if err := l.svcCtx.LoginLogDAO.Create(l.ctx, loginLog); err != nil {
		l.Logger.Errorf("记录登录成功日志失败: %v", err)
	}
}
