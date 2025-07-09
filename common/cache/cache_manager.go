package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager 缓存管理器
type CacheManager struct {
	client *redis.Client
	prefix string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL    time.Duration // 缓存过期时间
	Prefix string        // 缓存键前缀
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(client *redis.Client, prefix string) *CacheManager {
	return &CacheManager{
		client: client,
		prefix: prefix,
	}
}

// buildKey 构建缓存键
func (cm *CacheManager) buildKey(key string) string {
	if cm.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", cm.prefix, key)
}

// Set 设置缓存
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	cacheKey := cm.buildKey(key)
	return cm.client.Set(ctx, cacheKey, data, ttl).Err()
}

// Get 获取缓存
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	cacheKey := cm.buildKey(key)
	data, err := cm.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// Delete 删除缓存
func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}

	return cm.client.Del(ctx, cacheKeys...).Err()
}

// DeletePattern 按模式删除缓存
func (cm *CacheManager) DeletePattern(ctx context.Context, pattern string) error {
	// 构建完整的模式
	fullPattern := cm.buildKey(pattern)

	// 使用SCAN命令获取匹配的键
	var cursor uint64
	var keys []string

	for {
		result, newCursor, err := cm.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, result...)
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	// 如果有匹配的键，则删除它们
	if len(keys) > 0 {
		return cm.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists 检查缓存是否存在
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := cm.buildKey(key)
	count, err := cm.client.Exists(ctx, cacheKey).Result()
	return count > 0, err
}

// TTL 获取缓存剩余生存时间
func (cm *CacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	cacheKey := cm.buildKey(key)
	return cm.client.TTL(ctx, cacheKey).Result()
}

// Expire 设置缓存过期时间
func (cm *CacheManager) Expire(ctx context.Context, key string, ttl time.Duration) error {
	cacheKey := cm.buildKey(key)
	return cm.client.Expire(ctx, cacheKey, ttl).Err()
}

// GetOrSet 获取缓存，如果不存在则设置
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, setter func() (interface{}, error)) error {
	// 尝试获取缓存
	err := cm.Get(ctx, key, dest)
	if err == nil {
		return nil // 缓存存在且获取成功
	}

	// 如果是Redis连接错误，直接返回错误
	if err != redis.Nil {
		return fmt.Errorf("cache get error: %w", err)
	}

	// 缓存不存在，调用setter获取数据
	value, err := setter()
	if err != nil {
		return fmt.Errorf("setter function error: %w", err)
	}

	// 设置缓存
	if err := cm.Set(ctx, key, value, ttl); err != nil {
		// 设置缓存失败不影响数据返回，只记录错误
		// 在生产环境中可以添加日志记录
	}

	// 将获取的值赋给dest
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return json.Unmarshal(data, dest)
}

// Increment 递增计数器
func (cm *CacheManager) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	cacheKey := cm.buildKey(key)
	return cm.client.IncrBy(ctx, cacheKey, delta).Result()
}

// SetWithExpiration 设置带有具体过期时间的缓存
func (cm *CacheManager) SetWithExpiration(ctx context.Context, key string, value interface{}, expireAt time.Time) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	cacheKey := cm.buildKey(key)
	return cm.client.SetEx(ctx, cacheKey, data, time.Until(expireAt)).Err()
}

// GetMultiple 批量获取缓存
func (cm *CacheManager) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}

	results, err := cm.client.MGet(ctx, cacheKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple cache values: %w", err)
	}

	resultMap := make(map[string]string)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				resultMap[keys[i]] = str
			}
		}
	}

	return resultMap, nil
}

// SetMultiple 批量设置缓存
func (cm *CacheManager) SetMultiple(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := cm.client.Pipeline()
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value for key %s: %w", key, err)
		}

		cacheKey := cm.buildKey(key)
		pipe.Set(ctx, cacheKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// InvalidateByTags 根据标签失效缓存
func (cm *CacheManager) InvalidateByTags(ctx context.Context, tags ...string) error {
	for _, tag := range tags {
		pattern := fmt.Sprintf("*:%s:*", tag)
		if err := cm.DeletePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate cache by tag %s: %w", tag, err)
		}
	}
	return nil
}

// Clear 清空所有缓存（慎用）
func (cm *CacheManager) Clear(ctx context.Context) error {
	if cm.prefix == "" {
		return fmt.Errorf("cannot clear cache without prefix - too dangerous")
	}

	pattern := cm.prefix + ":*"
	return cm.DeletePattern(ctx, pattern)
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := cm.client.Info(ctx, "memory", "stats", "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	stats := make(map[string]interface{})

	// 解析INFO命令结果
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				stats[parts[0]] = parts[1]
			}
		}
	}

	// 添加连接池统计
	poolStats := cm.client.PoolStats()
	stats["pool_stats"] = map[string]interface{}{
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
		"timeouts":    poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}

	return stats, nil
}
