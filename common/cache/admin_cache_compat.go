package cache

import (
	redisv8 "github.com/go-redis/redis/v8"
	redisv9 "github.com/redis/go-redis/v9"
)

// NewAdminCacheManagerCompat 创建与v8 Redis客户端兼容的管理员缓存管理器
func NewAdminCacheManagerCompat(clientV8 *redisv8.Client, prefix string) *AdminCacheManager {
	// 创建v9客户端的适配器
	clientV9 := &redisv9.Client{}
	
	// TODO: 在后续任务中完成Redis客户端版本统一
	// 暂时返回nil，需要在实际使用时处理
	return &AdminCacheManager{
		CacheManager: NewCacheManager(clientV9, prefix),
	}
}