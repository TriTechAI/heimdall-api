package client

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis连接配置
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password,optional"`
	Database     int    `json:"database,optional"`
	MaxRetries   int    `json:"maxRetries,optional"`
	PoolSize     int    `json:"poolSize,optional"`
	MinIdleConns int    `json:"minIdleConns,optional"`
	Timeout      int    `json:"timeout,optional"`
}

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	config RedisConfig
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	// 设置默认值
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 1
	}
	if config.Timeout == 0 {
		config.Timeout = 5
	}

	// 创建Redis客户端选项
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  time.Duration(config.Timeout) * time.Second,
		ReadTimeout:  time.Duration(config.Timeout) * time.Second,
		WriteTimeout: time.Duration(config.Timeout) * time.Second,
	}

	// 创建客户端
	client := redis.NewClient(options)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: config,
	}, nil
}

// GetClient 获取Redis客户端实例
func (rc *RedisClient) GetClient() *redis.Client {
	return rc.client
}

// Ping 检查连接是否正常
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close 关闭连接
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// IsHealthy 检查Redis健康状态
func (rc *RedisClient) IsHealthy(ctx context.Context) bool {
	// 设置较短的超时时间用于健康检查
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// 执行ping命令
	err := rc.client.Ping(healthCtx).Err()
	if err != nil {
		return false
	}

	// 测试基本读写操作
	testKey := "health_check_test"
	err = rc.client.Set(healthCtx, testKey, "ok", time.Second).Err()
	if err != nil {
		return false
	}

	_, err = rc.client.Get(healthCtx, testKey).Result()
	if err != nil {
		return false
	}

	// 清理测试数据
	rc.client.Del(healthCtx, testKey)

	return true
}

// GetStats 获取Redis统计信息
func (rc *RedisClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取Redis INFO信息
	info, err := rc.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// 解析关键指标
	stats["info"] = info

	// 获取连接池统计
	poolStats := rc.client.PoolStats()
	stats["pool"] = map[string]interface{}{
		"hits":       poolStats.Hits,
		"misses":     poolStats.Misses,
		"timeouts":   poolStats.Timeouts,
		"totalConns": poolStats.TotalConns,
		"idleConns":  poolStats.IdleConns,
		"staleConns": poolStats.StaleConns,
	}

	// 获取数据库大小
	dbSize, err := rc.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	stats["dbSize"] = dbSize

	return stats, nil
}

// Set 设置键值对
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Del 删除键
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (rc *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.client.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.client.TTL(ctx, key).Result()
}

// Incr 递增键的值
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增键的值
func (rc *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return rc.client.IncrBy(ctx, key, value).Result()
}
