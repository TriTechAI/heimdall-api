package cache

import (
	"context"

	redisv8 "github.com/go-redis/redis/v8"
)

// CacheInvalidator 缓存失效器 - 兼容Redis v8客户端
type CacheInvalidator struct {
	client *redisv8.Client
	prefix string
}

// NewCacheInvalidator 创建缓存失效器
func NewCacheInvalidator(client *redisv8.Client, prefix string) *CacheInvalidator {
	return &CacheInvalidator{
		client: client,
		prefix: prefix,
	}
}

// InvalidatePattern 按模式失效缓存
func (ci *CacheInvalidator) InvalidatePattern(ctx context.Context, pattern string) error {
	// 构建完整的模式
	fullPattern := ci.prefix + ":" + pattern

	// 使用SCAN命令获取匹配的键
	var cursor uint64
	var keys []string

	for {
		result, newCursor, err := ci.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, result...)
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	// 如果有匹配的键，则删除它们
	if len(keys) > 0 {
		return ci.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateKeys 删除指定的缓存键
func (ci *CacheInvalidator) InvalidateKeys(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = ci.prefix + ":" + key
	}

	return ci.client.Del(ctx, fullKeys...).Err()
}

// InvalidatePostRelated 失效文章相关缓存
func (ci *CacheInvalidator) InvalidatePostRelated(ctx context.Context, postSlug string) error {
	patterns := []string{
		"post:detail:" + postSlug,
		"post:list:*",
		"search:*",
	}

	for _, pattern := range patterns {
		if err := ci.InvalidatePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateTagRelated 失效标签相关缓存
func (ci *CacheInvalidator) InvalidateTagRelated(ctx context.Context, tagSlug string) error {
	patterns := []string{
		"tag:detail:" + tagSlug,
		"tag:list:*",
		"post:list:*", // 标签变更可能影响文章列表
	}

	for _, pattern := range patterns {
		if err := ci.InvalidatePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// InvalidateUserRelated 失效用户相关缓存
func (ci *CacheInvalidator) InvalidateUserRelated(ctx context.Context, userID string) error {
	patterns := []string{
		"user:info:" + userID,
		"post:list:*", // 用户变更可能影响文章列表（作者信息）
	}

	for _, pattern := range patterns {
		if err := ci.InvalidatePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}
