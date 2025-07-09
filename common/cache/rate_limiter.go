package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

// RateLimiterCache 限流缓存服务
type RateLimiterCache struct {
	client redis.Cmdable
	prefix string
	logger logx.Logger
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Limit  int           // 限制次数
	Window time.Duration // 时间窗口
	Burst  int           // 突发限制（可选）
}

// RateLimitResult 限流检查结果
type RateLimitResult struct {
	Allowed    bool          // 是否允许请求
	Current    int           // 当前计数
	Limit      int           // 限制数量
	Remaining  int           // 剩余数量
	ResetAt    time.Time     // 重置时间
	RetryAfter time.Duration // 重试间隔
}

// NewRateLimiterCache 创建限流缓存实例
func NewRateLimiterCache(client redis.Cmdable, prefix string) *RateLimiterCache {
	if client == nil {
		panic("redis client cannot be nil")
	}
	if prefix == "" {
		prefix = "rate_limit:"
	}

	return &RateLimiterCache{
		client: client,
		prefix: prefix,
		logger: logx.WithContext(context.Background()),
	}
}

// CheckRateLimit 检查是否超过限流限制（滑动窗口算法）
func (r *RateLimiterCache) CheckRateLimit(ctx context.Context, key string, config RateLimitConfig) (*RateLimitResult, error) {
	if err := r.validateConfig(config); err != nil {
		return nil, err
	}

	redisKey := r.buildRateLimitKey(key)
	now := time.Now()
	windowStart := now.Add(-config.Window)

	// 使用Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_seconds = tonumber(ARGV[4])
		
		-- 清理过期的计数
		redis.call('zremrangebyscore', key, 0, window_start)
		
		-- 获取当前窗口内的计数
		local current = redis.call('zcard', key)
		
		-- 计算剩余数量
		local remaining = math.max(0, limit - current)
		
		-- 计算重置时间（窗口结束时间）
		local reset_at = now + window_seconds
		
		if current < limit then
			-- 未超限，添加新的计数记录
			redis.call('zadd', key, now, now)
			redis.call('expire', key, window_seconds)
			return {1, current + 1, limit, remaining - 1, reset_at}
		else
			-- 已超限
			return {0, current, limit, remaining, reset_at}
		end
	`

	result, err := r.client.Eval(ctx, luaScript, []string{redisKey},
		windowStart.Unix(), now.Unix(), config.Limit, int(config.Window.Seconds())).Result()
	if err != nil {
		r.logger.Errorf("Rate limit check failed for key %s: %v", key, err)
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	// 解析结果
	return r.parseRateLimitResult(result, config)
}

// CheckRateLimitFixed 固定窗口限流检查
func (r *RateLimiterCache) CheckRateLimitFixed(ctx context.Context, key string, config RateLimitConfig) (*RateLimitResult, error) {
	if err := r.validateConfig(config); err != nil {
		return nil, err
	}

	now := time.Now()
	windowKey := r.buildFixedWindowKey(key, now, config.Window)

	// 使用Lua脚本确保原子性
	luaScript := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_seconds = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		-- 获取当前计数
		local current = redis.call('get', key)
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end
		
		-- 计算剩余数量
		local remaining = math.max(0, limit - current)
		
		-- 计算重置时间
		local reset_at = now + window_seconds
		
		if current < limit then
			-- 未超限，增加计数
			local new_count = redis.call('incr', key)
			redis.call('expire', key, window_seconds)
			return {1, new_count, limit, remaining - 1, reset_at}
		else
			-- 已超限
			return {0, current, limit, remaining, reset_at}
		end
	`

	result, err := r.client.Eval(ctx, luaScript, []string{windowKey},
		config.Limit, int(config.Window.Seconds()), now.Unix()).Result()
	if err != nil {
		r.logger.Errorf("Fixed rate limit check failed for key %s: %v", key, err)
		return nil, fmt.Errorf("fixed rate limit check failed: %w", err)
	}

	return r.parseRateLimitResult(result, config)
}

// IncrementCounter 简单计数器递增
func (r *RateLimiterCache) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	redisKey := r.buildCounterKey(key)

	// 使用Pipeline确保原子性
	pipe := r.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Errorf("Failed to increment counter for key %s: %v", key, err)
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	return incrCmd.Val(), nil
}

// GetCounter 获取计数器值
func (r *RateLimiterCache) GetCounter(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	redisKey := r.buildCounterKey(key)

	result := r.client.Get(ctx, redisKey)
	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		r.logger.Errorf("Failed to get counter for key %s: %v", key, err)
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}

	count, err := strconv.ParseInt(result.Val(), 10, 64)
	if err != nil {
		r.logger.Errorf("Failed to parse counter value for key %s: %v", key, err)
		return 0, fmt.Errorf("failed to parse counter: %w", err)
	}

	return count, nil
}

// ResetCounter 重置计数器
func (r *RateLimiterCache) ResetCounter(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	redisKey := r.buildCounterKey(key)

	err := r.client.Del(ctx, redisKey).Err()
	if err != nil {
		r.logger.Errorf("Failed to reset counter for key %s: %v", key, err)
		return fmt.Errorf("failed to reset counter: %w", err)
	}

	return nil
}

// SetCustomLimit 为特定key设置自定义限制
func (r *RateLimiterCache) SetCustomLimit(ctx context.Context, key string, limit int, window time.Duration) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if limit <= 0 {
		return errors.New("limit must be positive")
	}
	if window <= 0 {
		return errors.New("window must be positive")
	}

	configKey := r.buildConfigKey(key)
	configValue := fmt.Sprintf("%d:%d", limit, int(window.Seconds()))

	err := r.client.Set(ctx, configKey, configValue, 24*time.Hour).Err()
	if err != nil {
		r.logger.Errorf("Failed to set custom limit for key %s: %v", key, err)
		return fmt.Errorf("failed to set custom limit: %w", err)
	}

	r.logger.Infof("Custom limit set for key %s: %d requests per %v", key, limit, window)
	return nil
}

// GetCustomLimit 获取特定key的自定义限制
func (r *RateLimiterCache) GetCustomLimit(ctx context.Context, key string) (*RateLimitConfig, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}

	configKey := r.buildConfigKey(key)

	result := r.client.Get(ctx, configKey)
	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // 没有自定义配置
		}
		r.logger.Errorf("Failed to get custom limit for key %s: %v", key, err)
		return nil, fmt.Errorf("failed to get custom limit: %w", err)
	}

	return r.parseCustomLimit(result.Val())
}

// ClearAllLimits 清理所有限流数据（谨慎使用）
func (r *RateLimiterCache) ClearAllLimits(ctx context.Context) error {
	pattern := r.prefix + "*"

	var cursor uint64
	var deletedCount int64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			r.logger.Errorf("Failed to scan rate limit keys: %v", err)
			return fmt.Errorf("failed to scan rate limit keys: %w", err)
		}

		if len(keys) > 0 {
			deleted, err := r.client.Del(ctx, keys...).Result()
			if err != nil {
				r.logger.Errorf("Failed to delete rate limit keys: %v", err)
				return fmt.Errorf("failed to delete rate limit keys: %w", err)
			}
			deletedCount += deleted
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	r.logger.Infof("Cleared %d rate limit keys", deletedCount)
	return nil
}

// validateConfig 验证限流配置
func (r *RateLimiterCache) validateConfig(config RateLimitConfig) error {
	if config.Limit <= 0 {
		return errors.New("limit must be positive")
	}
	if config.Window <= 0 {
		return errors.New("window must be positive")
	}
	if config.Window > 24*time.Hour {
		return errors.New("window too large")
	}
	return nil
}

// parseRateLimitResult 解析限流检查结果
func (r *RateLimiterCache) parseRateLimitResult(result interface{}, config RateLimitConfig) (*RateLimitResult, error) {
	values, ok := result.([]interface{})
	if !ok || len(values) != 5 {
		return nil, errors.New("invalid rate limit result format")
	}

	allowed, _ := values[0].(int64)
	current, _ := values[1].(int64)
	limit, _ := values[2].(int64)
	remaining, _ := values[3].(int64)
	resetAt, _ := values[4].(int64)

	retryAfter := time.Duration(0)
	if allowed == 0 {
		retryAfter = time.Until(time.Unix(resetAt, 0))
		if retryAfter < 0 {
			retryAfter = config.Window
		}
	}

	return &RateLimitResult{
		Allowed:    allowed == 1,
		Current:    int(current),
		Limit:      int(limit),
		Remaining:  int(remaining),
		ResetAt:    time.Unix(resetAt, 0),
		RetryAfter: retryAfter,
	}, nil
}

// parseCustomLimit 解析自定义限制配置
func (r *RateLimiterCache) parseCustomLimit(configValue string) (*RateLimitConfig, error) {
	var limit, windowSeconds int
	if _, err := fmt.Sscanf(configValue, "%d:%d", &limit, &windowSeconds); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	return &RateLimitConfig{
		Limit:  limit,
		Window: time.Duration(windowSeconds) * time.Second,
	}, nil
}

// buildRateLimitKey 构建限流键名
func (r *RateLimiterCache) buildRateLimitKey(key string) string {
	return r.prefix + "sliding:" + key
}

// buildFixedWindowKey 构建固定窗口键名
func (r *RateLimiterCache) buildFixedWindowKey(key string, now time.Time, window time.Duration) string {
	windowStart := now.Truncate(window).Unix()
	return fmt.Sprintf("%sfixed:%s:%d", r.prefix, key, windowStart)
}

// buildCounterKey 构建计数器键名
func (r *RateLimiterCache) buildCounterKey(key string) string {
	return r.prefix + "counter:" + key
}

// buildConfigKey 构建配置键名
func (r *RateLimiterCache) buildConfigKey(key string) string {
	return r.prefix + "config:" + key
}
