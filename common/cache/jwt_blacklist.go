package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// JWTBlacklistCache JWT黑名单缓存服务
type JWTBlacklistCache struct {
	client redis.Cmdable
	prefix string
	logger logx.Logger
}

// NewJWTBlacklistCache 创建JWT黑名单缓存实例
func NewJWTBlacklistCache(client redis.Cmdable, prefix string) *JWTBlacklistCache {
	if client == nil {
		panic("redis client cannot be nil")
	}
	if prefix == "" {
		prefix = "jwt:blacklist:"
	}
	
	return &JWTBlacklistCache{
		client: client,
		prefix: prefix,
		logger: logx.WithContext(context.Background()),
	}
}

// AddToken 将JWT token添加到黑名单
func (j *JWTBlacklistCache) AddToken(ctx context.Context, tokenID string, expiration time.Duration) error {
	if err := j.validateTokenID(tokenID); err != nil {
		return err
	}
	
	key := j.buildTokenKey(tokenID)
	
	// 设置token为已注销状态，过期时间为JWT的剩余有效期
	err := j.client.Set(ctx, key, "revoked", expiration).Err()
	if err != nil {
		j.logger.Errorf("Failed to add token to blacklist: %v", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	
	j.logger.Infof("Token %s added to blacklist, expires in %v", tokenID, expiration)
	return nil
}

// IsTokenBlacklisted 检查JWT token是否在黑名单中
func (j *JWTBlacklistCache) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if err := j.validateTokenID(tokenID); err != nil {
		return false, err
	}
	
	key := j.buildTokenKey(tokenID)
	
	result := j.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			// Token不在黑名单中
			return false, nil
		}
		j.logger.Errorf("Failed to check token blacklist status: %v", err)
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	
	// Token存在于黑名单中
	return true, nil
}

// RemoveToken 从黑名单中移除JWT token（一般不需要，除非管理员手动操作）
func (j *JWTBlacklistCache) RemoveToken(ctx context.Context, tokenID string) error {
	if err := j.validateTokenID(tokenID); err != nil {
		return err
	}
	
	key := j.buildTokenKey(tokenID)
	
	err := j.client.Del(ctx, key).Err()
	if err != nil {
		j.logger.Errorf("Failed to remove token from blacklist: %v", err)
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}
	
	j.logger.Infof("Token %s removed from blacklist", tokenID)
	return nil
}

// GetBlacklistedTokensCount 获取当前黑名单中的token数量
func (j *JWTBlacklistCache) GetBlacklistedTokensCount(ctx context.Context) (int64, error) {
	pattern := j.prefix + "*"
	
	// 使用SCAN命令获取所有匹配的键
	var cursor uint64
	var count int64
	
	for {
		keys, nextCursor, err := j.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			j.logger.Errorf("Failed to scan blacklisted tokens: %v", err)
			return 0, fmt.Errorf("failed to count blacklisted tokens: %w", err)
		}
		
		count += int64(len(keys))
		cursor = nextCursor
		
		if cursor == 0 {
			break
		}
	}
	
	return count, nil
}

// CleanupExpiredTokens 清理已过期的黑名单token（Redis会自动清理，此方法主要用于监控）
func (j *JWTBlacklistCache) CleanupExpiredTokens(ctx context.Context) error {
	// Redis会自动清理过期的键，这里主要是为了日志记录
	count, err := j.GetBlacklistedTokensCount(ctx)
	if err != nil {
		return err
	}
	
	j.logger.Infof("Current blacklisted tokens count: %d", count)
	return nil
}

// AddTokenWithTTL 添加token到黑名单并设置精确的过期时间
func (j *JWTBlacklistCache) AddTokenWithTTL(ctx context.Context, tokenID string, expireAt time.Time) error {
	if err := j.validateTokenID(tokenID); err != nil {
		return err
	}
	
	now := time.Now()
	if expireAt.Before(now) {
		return errors.New("expiration time cannot be in the past")
	}
	
	ttl := expireAt.Sub(now)
	return j.AddToken(ctx, tokenID, ttl)
}

// BatchAddTokens 批量添加多个token到黑名单
func (j *JWTBlacklistCache) BatchAddTokens(ctx context.Context, tokens map[string]time.Duration) error {
	if len(tokens) == 0 {
		return errors.New("tokens map cannot be empty")
	}
	
	// 使用Pipeline进行批量操作
	pipe := j.client.TxPipeline()
	
	for tokenID, expiration := range tokens {
		if err := j.validateTokenID(tokenID); err != nil {
			return fmt.Errorf("invalid token %s: %w", tokenID, err)
		}
		
		key := j.buildTokenKey(tokenID)
		pipe.Set(ctx, key, "revoked", expiration)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		j.logger.Errorf("Failed to batch add tokens to blacklist: %v", err)
		return fmt.Errorf("failed to batch blacklist tokens: %w", err)
	}
	
	j.logger.Infof("Successfully blacklisted %d tokens", len(tokens))
	return nil
}

// BatchCheckTokens 批量检查多个token的黑名单状态
func (j *JWTBlacklistCache) BatchCheckTokens(ctx context.Context, tokenIDs []string) (map[string]bool, error) {
	if len(tokenIDs) == 0 {
		return make(map[string]bool), nil
	}
	
	// 构建所有的key
	keys := make([]string, len(tokenIDs))
	for i, tokenID := range tokenIDs {
		if err := j.validateTokenID(tokenID); err != nil {
			return nil, fmt.Errorf("invalid token %s: %w", tokenID, err)
		}
		keys[i] = j.buildTokenKey(tokenID)
	}
	
	// 使用Pipeline批量检查
	pipe := j.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		j.logger.Errorf("Failed to batch check tokens: %v", err)
		return nil, fmt.Errorf("failed to batch check tokens: %w", err)
	}
	
	// 解析结果
	result := make(map[string]bool)
	for i, cmd := range cmds {
		tokenID := tokenIDs[i]
		if err := cmd.Err(); err != nil {
			if errors.Is(err, redis.Nil) {
				result[tokenID] = false // 不在黑名单中
			} else {
				j.logger.Errorf("Failed to check token %s: %v", tokenID, err)
				return nil, fmt.Errorf("failed to check token %s: %w", tokenID, err)
			}
		} else {
			result[tokenID] = true // 在黑名单中
		}
	}
	
	return result, nil
}

// validateTokenID 验证token ID的有效性
func (j *JWTBlacklistCache) validateTokenID(tokenID string) error {
	if tokenID == "" {
		return errors.New("token ID cannot be empty")
	}
	if len(tokenID) < 10 {
		return errors.New("token ID too short")
	}
	if len(tokenID) > 200 {
		return errors.New("token ID too long")
	}
	return nil
}

// buildTokenKey 构建Redis中的token键名
func (j *JWTBlacklistCache) buildTokenKey(tokenID string) string {
	return j.prefix + tokenID
}