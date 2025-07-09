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
	// 1. 验证和安全检查
	if err := l.performSecurityChecks(req); err != nil {
		return nil, err
	}

	// 2. 用户认证
	clientIP := l.getClientIP()
	user, err := l.authenticateUser(req.Username, req.Password, clientIP)
	if err != nil {
		return nil, err
	}

	// 3. 生成令牌并处理登录成功
	tokens, err := l.generateTokens(user)
	if err != nil {
		return nil, err
	}
	l.handleLoginSuccess(user, clientIP)

	// 4. 构造响应
	return l.buildLoginResponse(user, tokens), nil
}

// performSecurityChecks 执行安全检查
func (l *LoginLogic) performSecurityChecks(req *types.LoginRequest) error {
	// 基础验证
	if err := l.validateRequest(req); err != nil {
		return err
	}

	// 检查登录尝试次数
	clientIP := l.getClientIP()
	if err := l.checkLoginAttempts(req.Username, clientIP); err != nil {
		return err
	}

	return nil
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

// getClientIP 从context中获取客户端IP地址
func (l *LoginLogic) getClientIP() string {
	if ip, ok := l.ctx.Value("client_ip").(string); ok && ip != "" {
		return ip
	}
	// 兜底方案：如果无法从context获取，返回unknown
	l.Logger.Error("无法从context获取客户端IP，使用unknown作为兜底")
	return "unknown"
}

// checkLoginAttempts 检查登录失败次数限制
func (l *LoginLogic) checkLoginAttempts(username, clientIP string) error {
	// 获取当前失败次数
	attempts, err := l.getLoginAttempts(username, clientIP)
	if err != nil {
		return err
	}

	// 检查是否超过最大失败次数
	if attempts >= l.svcCtx.Config.Security.MaxLoginAttempts {
		lockoutMinutes := l.svcCtx.Config.Security.LoginLockoutDuration / 60
		return fmt.Errorf("登录失败次数过多，请%d分钟后再试", lockoutMinutes)
	}

	return nil
}

// getLoginAttempts 获取登录失败次数
func (l *LoginLogic) getLoginAttempts(username, clientIP string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)
	result := l.svcCtx.Redis.Get(l.ctx, key)

	if result.Err() != nil {
		if errors.Is(result.Err(), redis.Nil) {
			return 0, nil // 没有失败记录
		}
		l.Logger.Errorf("获取登录失败次数失败: %v", result.Err())
		// Redis故障时采用严格安全策略：拒绝登录
		return 0, errors.New("安全服务暂时不可用，请稍后再试")
	}

	attempts, err := strconv.Atoi(result.Val())
	if err != nil {
		l.Logger.Errorf("解析登录失败次数失败: %v", err)
		// 数据格式错误时采用严格策略
		return 0, errors.New("安全验证数据异常，请联系管理员")
	}

	return attempts, nil
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
	ttl := time.Duration(l.svcCtx.Config.Cache.LoginAttempts.TTL) * time.Second

	// 执行原子操作
	count, err := l.executeAtomicIncrement(key, ttl)
	if err != nil {
		l.Logger.Errorf("增加登录失败次数失败: %v", err)
		return
	}

	l.Logger.Infof("用户 %s IP %s 登录失败次数: %d", username, clientIP, count)
}

// executeAtomicIncrement 执行原子递增操作
func (l *LoginLogic) executeAtomicIncrement(key string, ttl time.Duration) (int64, error) {
	// 使用Redis Pipeline确保原子性
	pipe := l.svcCtx.Redis.TxPipeline()

	// 增加计数
	incrCmd := pipe.Incr(l.ctx, key)

	// 设置过期时间
	pipe.Expire(l.ctx, key, ttl)

	// 执行管道操作
	_, err := pipe.Exec(l.ctx)
	if err != nil {
		return 0, err
	}

	return incrCmd.Val(), nil
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
	// 获取当前失败次数
	attempts := l.getCurrentAttempts(username, clientIP)

	// 如果达到最大失败次数，锁定账户
	if attempts >= l.svcCtx.Config.Security.MaxLoginAttempts {
		return l.lockUserAccount(user)
	}

	return nil
}

// getCurrentAttempts 获取当前Redis中的失败次数
func (l *LoginLogic) getCurrentAttempts(username, clientIP string) int {
	key := fmt.Sprintf("%s%s:%s", l.svcCtx.Config.Cache.LoginAttempts.Prefix, username, clientIP)
	result := l.svcCtx.Redis.Get(l.ctx, key)

	if result.Err() != nil {
		return 1 // 默认返回1
	}

	attempts, err := strconv.Atoi(result.Val())
	if err != nil {
		return 1 // 解析失败返回1
	}

	return attempts
}

// lockUserAccount 锁定用户账户
func (l *LoginLogic) lockUserAccount(user *model.User) error {
	lockDuration := time.Duration(l.svcCtx.Config.Security.LoginLockoutDuration) * time.Second
	lockUntil := time.Now().Add(lockDuration)

	if err := l.svcCtx.UserDAO.LockUser(l.ctx, user.ID.Hex(), lockUntil); err != nil {
		l.Logger.Errorf("锁定用户账户失败: %v", err)
	}

	lockoutMinutes := l.svcCtx.Config.Security.LoginLockoutDuration / 60
	return fmt.Errorf("登录失败次数过多，账户已被锁定%d分钟", lockoutMinutes)
}

// recordLoginFailure 记录登录失败日志
func (l *LoginLogic) recordLoginFailure(username, clientIP, reason string) {
	loginLog := &model.LoginLog{
		Username:    username,
		IPAddress:   clientIP,
		UserAgent:   "Admin Panel", // 简化处理
		LoginMethod: "password",    // 添加登录方式
		Status:      constants.LoginStatusFailed,
		FailReason:  reason,
		LoginAt:     time.Now(),
	}

	if err := l.svcCtx.LoginLogDAO.Create(l.ctx, loginLog); err != nil {
		l.Logger.Errorf("记录登录失败日志失败: %v", err)
	}
}

// recordLoginSuccess 记录登录成功日志
func (l *LoginLogic) recordLoginSuccess(user *model.User, clientIP string) {
	loginLog := &model.LoginLog{
		UserID:      &user.ID,
		Username:    user.Username,
		IPAddress:   clientIP,
		UserAgent:   "Admin Panel", // 简化处理
		LoginMethod: "password",    // 添加登录方式
		Status:      constants.LoginStatusSuccess,
		LoginAt:     time.Now(),
	}

	if err := l.svcCtx.LoginLogDAO.Create(l.ctx, loginLog); err != nil {
		l.Logger.Errorf("记录登录成功日志失败: %v", err)
	}
}

// ===============================
// 新增的原子化方法
// ===============================

// authenticateUser 用户认证
func (l *LoginLogic) authenticateUser(username, password, clientIP string) (*model.User, error) {
	// 获取用户信息
	user, err := l.fetchUser(username)
	if err != nil {
		return nil, err
	}

	// 验证用户凭证
	failureReason := l.verifyUserCredentials(user, password)
	if failureReason != "" {
		return nil, l.handleAuthenticationFailure(user, username, password, clientIP, failureReason)
	}

	return user, nil
}

// fetchUser 获取用户信息
func (l *LoginLogic) fetchUser(username string) (*model.User, error) {
	user, err := l.svcCtx.UserDAO.GetByUsername(l.ctx, username)
	if err != nil {
		l.Logger.Errorf("获取用户信息失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}
	return user, nil
}

// verifyUserCredentials 验证用户凭证
func (l *LoginLogic) verifyUserCredentials(user *model.User, password string) string {
	if user == nil {
		return "用户不存在"
	}

	// 检查用户状态
	if err := l.checkUserStatus(user); err != nil {
		return err.Error()
	}

	// 验证密码
	if err := utils.VerifyPassword(password, user.PasswordHash); err != nil {
		return "密码错误"
	}

	return ""
}

// handleAuthenticationFailure 处理认证失败
func (l *LoginLogic) handleAuthenticationFailure(user *model.User, username, password, clientIP, failureReason string) error {
	// 执行防时间攻击的密码验证
	if user == nil || failureReason != "密码错误" {
		l.performDummyPasswordCheck(password)
	}

	// 记录失败日志和增加失败次数
	l.recordLoginFailure(username, clientIP, failureReason)
	l.incrementLoginAttempts(username, clientIP)

	// 更新用户登录失败计数
	if user != nil && failureReason == "密码错误" {
		if err := l.svcCtx.UserDAO.IncrementLoginFailCount(l.ctx, user.ID.Hex()); err != nil {
			l.Logger.Errorf("更新用户登录失败次数失败: %v", err)
		}
	}

	// 统一返回相同的错误消息，防止用户枚举
	return errors.New("用户名或密码错误")
}

// performDummyPasswordCheck 执行虚拟密码检查，防止时间差异攻击
func (l *LoginLogic) performDummyPasswordCheck(password string) {
	// 使用固定的哈希值进行虚拟验证，确保执行时间与真实密码验证相同
	// 这个哈希值对应密码 "dummy_password_for_timing_attack_prevention"
	dummyHash := "$2a$12$KTGlrWkOKRASHE7e9hKzI.zyPmhNnD8sHk7wLZqN4KRCz3TnQKl8m"
	utils.VerifyPassword(password, dummyHash)
}

// TokenPair 令牌对结构
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// generateTokens 生成令牌
func (l *LoginLogic) generateTokens(user *model.User) (*TokenPair, error) {
	jwtManager := utils.NewJWTManager(l.svcCtx.Config.Auth.AccessSecret, "heimdall-admin")

	// 生成访问令牌
	accessToken, err := jwtManager.GenerateGoZeroCompatibleToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		l.Logger.Errorf("生成JWT Token失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	// 生成刷新令牌
	tokenPair, err := jwtManager.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		l.Logger.Errorf("生成Refresh Token失败: %v", err)
		return nil, errors.New("系统错误，请稍后重试")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int(tokenPair.ExpiresAt.Sub(time.Now()).Seconds()),
	}, nil
}

// handleLoginSuccess 处理登录成功后的操作
func (l *LoginLogic) handleLoginSuccess(user *model.User, clientIP string) {
	// 清除登录失败次数缓存
	l.clearLoginAttempts(user.Username, clientIP)

	// 更新用户登录信息
	if err := l.svcCtx.UserDAO.UpdateLoginInfo(l.ctx, user.ID.Hex(), clientIP); err != nil {
		l.Logger.Errorf("更新用户登录信息失败: %v", err)
		// 这个错误不阻止登录流程
	}

	// 记录成功登录日志
	l.recordLoginSuccess(user, clientIP)
}

// buildLoginResponse 构建登录响应
func (l *LoginLogic) buildLoginResponse(user *model.User, tokens *TokenPair) *types.LoginResponse {
	return &types.LoginResponse{
		Code:      200,
		Message:   "登录成功",
		Timestamp: time.Now().Format(constants.DefaultTimeFormat),
		Data: types.LoginData{
			Token:        tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresIn:    tokens.ExpiresIn,
			User: types.UserInfo{
				ID:          user.ID.Hex(),
				Username:    user.Username,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				Role:        user.Role,
				Status:      user.Status,
				CreatedAt:   user.CreatedAt.Format(constants.DefaultTimeFormat),
			},
		},
	}
}
