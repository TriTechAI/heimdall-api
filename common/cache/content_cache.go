package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ContentCacheManager 内容缓存管理器
type ContentCacheManager struct {
	*CacheManager
}

// ContentCacheTTL 内容缓存TTL配置
var ContentCacheTTL = struct {
	PostDetail time.Duration // 文章详情缓存 30分钟
	PostList   time.Duration // 文章列表缓存 10分钟
	TagList    time.Duration // 标签列表缓存 1小时
	TagDetail  time.Duration // 标签详情缓存 1小时
	PageDetail time.Duration // 页面详情缓存 30分钟
	UserInfo   time.Duration // 用户信息缓存 15分钟
	Search     time.Duration // 搜索结果缓存 5分钟
}{
	PostDetail: 30 * time.Minute,
	PostList:   10 * time.Minute,
	TagList:    1 * time.Hour,
	TagDetail:  1 * time.Hour,
	PageDetail: 30 * time.Minute,
	UserInfo:   15 * time.Minute,
	Search:     5 * time.Minute,
}

// NewContentCacheManager 创建内容缓存管理器
func NewContentCacheManager(client *redis.Client, prefix string) *ContentCacheManager {
	return &ContentCacheManager{
		CacheManager: NewCacheManager(client, prefix),
	}
}

// === 文章缓存 ===

// GetPostDetail 获取文章详情缓存
func (ccm *ContentCacheManager) GetPostDetail(ctx context.Context, slug string, dest interface{}) error {
	key := fmt.Sprintf("post:detail:%s", slug)
	return ccm.Get(ctx, key, dest)
}

// SetPostDetail 设置文章详情缓存
func (ccm *ContentCacheManager) SetPostDetail(ctx context.Context, slug string, post interface{}) error {
	key := fmt.Sprintf("post:detail:%s", slug)
	return ccm.Set(ctx, key, post, ContentCacheTTL.PostDetail)
}

// InvalidatePostDetail 失效文章详情缓存
func (ccm *ContentCacheManager) InvalidatePostDetail(ctx context.Context, slug string) error {
	key := fmt.Sprintf("post:detail:%s", slug)
	return ccm.Delete(ctx, key)
}

// GetPostList 获取文章列表缓存
func (ccm *ContentCacheManager) GetPostList(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := ccm.buildPostListKey(page, limit, filters)
	return ccm.Get(ctx, key, dest)
}

// SetPostList 设置文章列表缓存
func (ccm *ContentCacheManager) SetPostList(ctx context.Context, page, limit int, filters map[string]interface{}, posts interface{}) error {
	key := ccm.buildPostListKey(page, limit, filters)
	return ccm.Set(ctx, key, posts, ContentCacheTTL.PostList)
}

// InvalidatePostList 失效文章列表缓存
func (ccm *ContentCacheManager) InvalidatePostList(ctx context.Context) error {
	pattern := "post:list:*"
	return ccm.DeletePattern(ctx, pattern)
}

// buildPostListKey 构建文章列表缓存键
func (ccm *ContentCacheManager) buildPostListKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("post:list:p%d:l%d", page, limit)
	
	// 添加过滤条件到键中
	if tag, ok := filters["tag"].(string); ok && tag != "" {
		key += fmt.Sprintf(":tag:%s", tag)
	}
	if author, ok := filters["author"].(string); ok && author != "" {
		key += fmt.Sprintf(":author:%s", author)
	}
	if sortBy, ok := filters["sortBy"].(string); ok && sortBy != "" {
		key += fmt.Sprintf(":sort:%s", sortBy)
	}
	if sortDesc, ok := filters["sortDesc"].(bool); ok && sortDesc {
		key += ":desc"
	}
	
	return key
}

// === 标签缓存 ===

// GetTagList 获取标签列表缓存
func (ccm *ContentCacheManager) GetTagList(ctx context.Context, page, limit int, sortBy string, sortDesc bool, dest interface{}) error {
	key := fmt.Sprintf("tag:list:p%d:l%d:sort:%s", page, limit, sortBy)
	if sortDesc {
		key += ":desc"
	}
	return ccm.Get(ctx, key, dest)
}

// SetTagList 设置标签列表缓存
func (ccm *ContentCacheManager) SetTagList(ctx context.Context, page, limit int, sortBy string, sortDesc bool, tags interface{}) error {
	key := fmt.Sprintf("tag:list:p%d:l%d:sort:%s", page, limit, sortBy)
	if sortDesc {
		key += ":desc"
	}
	return ccm.Set(ctx, key, tags, ContentCacheTTL.TagList)
}

// InvalidateTagList 失效标签列表缓存
func (ccm *ContentCacheManager) InvalidateTagList(ctx context.Context) error {
	pattern := "tag:list:*"
	return ccm.DeletePattern(ctx, pattern)
}

// GetTagDetail 获取标签详情缓存
func (ccm *ContentCacheManager) GetTagDetail(ctx context.Context, slug string, dest interface{}) error {
	key := fmt.Sprintf("tag:detail:%s", slug)
	return ccm.Get(ctx, key, dest)
}

// SetTagDetail 设置标签详情缓存
func (ccm *ContentCacheManager) SetTagDetail(ctx context.Context, slug string, tag interface{}) error {
	key := fmt.Sprintf("tag:detail:%s", slug)
	return ccm.Set(ctx, key, tag, ContentCacheTTL.TagDetail)
}

// InvalidateTagDetail 失效标签详情缓存
func (ccm *ContentCacheManager) InvalidateTagDetail(ctx context.Context, slug string) error {
	key := fmt.Sprintf("tag:detail:%s", slug)
	return ccm.Delete(ctx, key)
}

// === 页面缓存 ===

// GetPageDetail 获取页面详情缓存
func (ccm *ContentCacheManager) GetPageDetail(ctx context.Context, slug string, dest interface{}) error {
	key := fmt.Sprintf("page:detail:%s", slug)
	return ccm.Get(ctx, key, dest)
}

// SetPageDetail 设置页面详情缓存
func (ccm *ContentCacheManager) SetPageDetail(ctx context.Context, slug string, page interface{}) error {
	key := fmt.Sprintf("page:detail:%s", slug)
	return ccm.Set(ctx, key, page, ContentCacheTTL.PageDetail)
}

// InvalidatePageDetail 失效页面详情缓存
func (ccm *ContentCacheManager) InvalidatePageDetail(ctx context.Context, slug string) error {
	key := fmt.Sprintf("page:detail:%s", slug)
	return ccm.Delete(ctx, key)
}

// === 用户缓存 ===

// GetUserInfo 获取用户信息缓存
func (ccm *ContentCacheManager) GetUserInfo(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("user:info:%s", userID)
	return ccm.Get(ctx, key, dest)
}

// SetUserInfo 设置用户信息缓存
func (ccm *ContentCacheManager) SetUserInfo(ctx context.Context, userID string, user interface{}) error {
	key := fmt.Sprintf("user:info:%s", userID)
	return ccm.Set(ctx, key, user, ContentCacheTTL.UserInfo)
}

// InvalidateUserInfo 失效用户信息缓存
func (ccm *ContentCacheManager) InvalidateUserInfo(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:info:%s", userID)
	return ccm.Delete(ctx, key)
}

// === 搜索缓存 ===

// GetSearchResult 获取搜索结果缓存
func (ccm *ContentCacheManager) GetSearchResult(ctx context.Context, query string, page, limit int, dest interface{}) error {
	key := fmt.Sprintf("search:%s:p%d:l%d", query, page, limit)
	return ccm.Get(ctx, key, dest)
}

// SetSearchResult 设置搜索结果缓存
func (ccm *ContentCacheManager) SetSearchResult(ctx context.Context, query string, page, limit int, result interface{}) error {
	key := fmt.Sprintf("search:%s:p%d:l%d", query, page, limit)
	return ccm.Set(ctx, key, result, ContentCacheTTL.Search)
}

// InvalidateSearchResults 失效搜索结果缓存
func (ccm *ContentCacheManager) InvalidateSearchResults(ctx context.Context) error {
	pattern := "search:*"
	return ccm.DeletePattern(ctx, pattern)
}

// === 统计信息缓存 ===

// GetViewCount 获取浏览计数缓存
func (ccm *ContentCacheManager) GetViewCount(ctx context.Context, contentType, contentID string) (int64, error) {
	key := fmt.Sprintf("view:count:%s:%s", contentType, contentID)
	countStr, err := ccm.client.Get(ctx, ccm.buildKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse view count: %w", err)
	}
	
	return count, nil
}

// IncrementViewCount 增加浏览计数
func (ccm *ContentCacheManager) IncrementViewCount(ctx context.Context, contentType, contentID string) error {
	key := fmt.Sprintf("view:count:%s:%s", contentType, contentID)
	_, err := ccm.Increment(ctx, key, 1)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}
	
	// 设置过期时间（24小时）
	return ccm.Expire(ctx, key, 24*time.Hour)
}

// === 批量操作 ===

// InvalidatePostRelated 失效文章相关的所有缓存
func (ccm *ContentCacheManager) InvalidatePostRelated(ctx context.Context, slug string) error {
	// 失效文章详情
	if err := ccm.InvalidatePostDetail(ctx, slug); err != nil {
		return fmt.Errorf("failed to invalidate post detail: %w", err)
	}
	
	// 失效文章列表
	if err := ccm.InvalidatePostList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate post list: %w", err)
	}
	
	// 失效搜索结果
	if err := ccm.InvalidateSearchResults(ctx); err != nil {
		return fmt.Errorf("failed to invalidate search results: %w", err)
	}
	
	return nil
}

// InvalidateTagRelated 失效标签相关的所有缓存
func (ccm *ContentCacheManager) InvalidateTagRelated(ctx context.Context, slug string) error {
	// 失效标签详情
	if err := ccm.InvalidateTagDetail(ctx, slug); err != nil {
		return fmt.Errorf("failed to invalidate tag detail: %w", err)
	}
	
	// 失效标签列表
	if err := ccm.InvalidateTagList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate tag list: %w", err)
	}
	
	// 失效相关文章列表
	if err := ccm.InvalidatePostList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate post list: %w", err)
	}
	
	return nil
}

// WarmUpCache 缓存预热
func (ccm *ContentCacheManager) WarmUpCache(ctx context.Context, warmupFunc func(ctx context.Context) error) error {
	return warmupFunc(ctx)
}

// === 缓存统计 ===

// GetCacheHitRate 获取缓存命中率
func (ccm *ContentCacheManager) GetCacheHitRate(ctx context.Context) (float64, error) {
	stats, err := ccm.GetStats(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get cache stats: %w", err)
	}
	
	// 从连接池统计中获取命中率
	if poolStats, ok := stats["pool_stats"].(map[string]interface{}); ok {
		if hits, hitOk := poolStats["hits"].(uint32); hitOk {
			if misses, missOk := poolStats["misses"].(uint32); missOk {
				total := hits + misses
				if total > 0 {
					return float64(hits) / float64(total), nil
				}
			}
		}
	}
	
	return 0, nil
}