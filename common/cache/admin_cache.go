package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdminCacheManager 管理员缓存管理器
type AdminCacheManager struct {
	*CacheManager
}

// AdminCacheTTL 管理员缓存TTL配置
var AdminCacheTTL = struct {
	UserList   time.Duration // 用户列表缓存 5分钟
	UserDetail time.Duration // 用户详情缓存 10分钟
	PostList   time.Duration // 文章列表缓存 3分钟
	PostDetail time.Duration // 文章详情缓存 5分钟
	TagList    time.Duration // 标签列表缓存 10分钟
	PageList   time.Duration // 页面列表缓存 5分钟
	LoginLogs  time.Duration // 登录日志缓存 5分钟
	Dashboard  time.Duration // 仪表盘数据缓存 15分钟
	Analytics  time.Duration // 分析数据缓存 30分钟
	SystemInfo time.Duration // 系统信息缓存 5分钟
}{
	UserList:   5 * time.Minute,
	UserDetail: 10 * time.Minute,
	PostList:   3 * time.Minute,
	PostDetail: 5 * time.Minute,
	TagList:    10 * time.Minute,
	PageList:   5 * time.Minute,
	LoginLogs:  5 * time.Minute,
	Dashboard:  15 * time.Minute,
	Analytics:  30 * time.Minute,
	SystemInfo: 5 * time.Minute,
}

// NewAdminCacheManager 创建管理员缓存管理器
func NewAdminCacheManager(client *redis.Client, prefix string) *AdminCacheManager {
	return &AdminCacheManager{
		CacheManager: NewCacheManager(client, prefix),
	}
}

// === 用户管理缓存 ===

// GetUserList 获取用户列表缓存
func (acm *AdminCacheManager) GetUserList(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := acm.buildUserListKey(page, limit, filters)
	return acm.Get(ctx, key, dest)
}

// SetUserList 设置用户列表缓存
func (acm *AdminCacheManager) SetUserList(ctx context.Context, page, limit int, filters map[string]interface{}, users interface{}) error {
	key := acm.buildUserListKey(page, limit, filters)
	return acm.Set(ctx, key, users, AdminCacheTTL.UserList)
}

// InvalidateUserList 失效用户列表缓存
func (acm *AdminCacheManager) InvalidateUserList(ctx context.Context) error {
	pattern := "admin:user:list:*"
	return acm.DeletePattern(ctx, pattern)
}

// buildUserListKey 构建用户列表缓存键
func (acm *AdminCacheManager) buildUserListKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("admin:user:list:p%d:l%d", page, limit)

	if role, ok := filters["role"].(string); ok && role != "" {
		key += fmt.Sprintf(":role:%s", role)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		key += fmt.Sprintf(":status:%s", status)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		key += fmt.Sprintf(":keyword:%s", keyword)
	}
	if sortBy, ok := filters["sortBy"].(string); ok && sortBy != "" {
		key += fmt.Sprintf(":sort:%s", sortBy)
	}
	if sortDesc, ok := filters["sortDesc"].(bool); ok && sortDesc {
		key += ":desc"
	}

	return key
}

// GetUserDetail 获取用户详情缓存
func (acm *AdminCacheManager) GetUserDetail(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("admin:user:detail:%s", userID)
	return acm.Get(ctx, key, dest)
}

// SetUserDetail 设置用户详情缓存
func (acm *AdminCacheManager) SetUserDetail(ctx context.Context, userID string, user interface{}) error {
	key := fmt.Sprintf("admin:user:detail:%s", userID)
	return acm.Set(ctx, key, user, AdminCacheTTL.UserDetail)
}

// InvalidateUserDetail 失效用户详情缓存
func (acm *AdminCacheManager) InvalidateUserDetail(ctx context.Context, userID string) error {
	key := fmt.Sprintf("admin:user:detail:%s", userID)
	return acm.Delete(ctx, key)
}

// === 文章管理缓存 ===

// GetPostList 获取文章列表缓存
func (acm *AdminCacheManager) GetPostList(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := acm.buildPostListKey(page, limit, filters)
	return acm.Get(ctx, key, dest)
}

// SetPostList 设置文章列表缓存
func (acm *AdminCacheManager) SetPostList(ctx context.Context, page, limit int, filters map[string]interface{}, posts interface{}) error {
	key := acm.buildPostListKey(page, limit, filters)
	return acm.Set(ctx, key, posts, AdminCacheTTL.PostList)
}

// InvalidatePostList 失效文章列表缓存
func (acm *AdminCacheManager) InvalidatePostList(ctx context.Context) error {
	pattern := "admin:post:list:*"
	return acm.DeletePattern(ctx, pattern)
}

// buildPostListKey 构建文章列表缓存键
func (acm *AdminCacheManager) buildPostListKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("admin:post:list:p%d:l%d", page, limit)

	if status, ok := filters["status"].(string); ok && status != "" {
		key += fmt.Sprintf(":status:%s", status)
	}
	if authorID, ok := filters["authorID"].(string); ok && authorID != "" {
		key += fmt.Sprintf(":author:%s", authorID)
	}
	if tag, ok := filters["tag"].(string); ok && tag != "" {
		key += fmt.Sprintf(":tag:%s", tag)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		key += fmt.Sprintf(":keyword:%s", keyword)
	}
	if sortBy, ok := filters["sortBy"].(string); ok && sortBy != "" {
		key += fmt.Sprintf(":sort:%s", sortBy)
	}
	if sortDesc, ok := filters["sortDesc"].(bool); ok && sortDesc {
		key += ":desc"
	}

	return key
}

// GetPostDetail 获取文章详情缓存
func (acm *AdminCacheManager) GetPostDetail(ctx context.Context, postID string, dest interface{}) error {
	key := fmt.Sprintf("admin:post:detail:%s", postID)
	return acm.Get(ctx, key, dest)
}

// SetPostDetail 设置文章详情缓存
func (acm *AdminCacheManager) SetPostDetail(ctx context.Context, postID string, post interface{}) error {
	key := fmt.Sprintf("admin:post:detail:%s", postID)
	return acm.Set(ctx, key, post, AdminCacheTTL.PostDetail)
}

// InvalidatePostDetail 失效文章详情缓存
func (acm *AdminCacheManager) InvalidatePostDetail(ctx context.Context, postID string) error {
	key := fmt.Sprintf("admin:post:detail:%s", postID)
	return acm.Delete(ctx, key)
}

// === 标签管理缓存 ===

// GetTagList 获取标签列表缓存
func (acm *AdminCacheManager) GetTagList(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := acm.buildTagListKey(page, limit, filters)
	return acm.Get(ctx, key, dest)
}

// SetTagList 设置标签列表缓存
func (acm *AdminCacheManager) SetTagList(ctx context.Context, page, limit int, filters map[string]interface{}, tags interface{}) error {
	key := acm.buildTagListKey(page, limit, filters)
	return acm.Set(ctx, key, tags, AdminCacheTTL.TagList)
}

// InvalidateTagList 失效标签列表缓存
func (acm *AdminCacheManager) InvalidateTagList(ctx context.Context) error {
	pattern := "admin:tag:list:*"
	return acm.DeletePattern(ctx, pattern)
}

// buildTagListKey 构建标签列表缓存键
func (acm *AdminCacheManager) buildTagListKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("admin:tag:list:p%d:l%d", page, limit)

	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		key += fmt.Sprintf(":keyword:%s", keyword)
	}
	if sortBy, ok := filters["sortBy"].(string); ok && sortBy != "" {
		key += fmt.Sprintf(":sort:%s", sortBy)
	}
	if sortDesc, ok := filters["sortDesc"].(bool); ok && sortDesc {
		key += ":desc"
	}

	return key
}

// === 页面管理缓存 ===

// GetPageList 获取页面列表缓存
func (acm *AdminCacheManager) GetPageList(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := acm.buildPageListKey(page, limit, filters)
	return acm.Get(ctx, key, dest)
}

// SetPageList 设置页面列表缓存
func (acm *AdminCacheManager) SetPageList(ctx context.Context, page, limit int, filters map[string]interface{}, pages interface{}) error {
	key := acm.buildPageListKey(page, limit, filters)
	return acm.Set(ctx, key, pages, AdminCacheTTL.PageList)
}

// InvalidatePageList 失效页面列表缓存
func (acm *AdminCacheManager) InvalidatePageList(ctx context.Context) error {
	pattern := "admin:page:list:*"
	return acm.DeletePattern(ctx, pattern)
}

// buildPageListKey 构建页面列表缓存键
func (acm *AdminCacheManager) buildPageListKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("admin:page:list:p%d:l%d", page, limit)

	if status, ok := filters["status"].(string); ok && status != "" {
		key += fmt.Sprintf(":status:%s", status)
	}
	if template, ok := filters["template"].(string); ok && template != "" {
		key += fmt.Sprintf(":template:%s", template)
	}
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		key += fmt.Sprintf(":keyword:%s", keyword)
	}
	if sortBy, ok := filters["sortBy"].(string); ok && sortBy != "" {
		key += fmt.Sprintf(":sort:%s", sortBy)
	}
	if sortDesc, ok := filters["sortDesc"].(bool); ok && sortDesc {
		key += ":desc"
	}

	return key
}

// === 登录日志缓存 ===

// GetLoginLogs 获取登录日志缓存
func (acm *AdminCacheManager) GetLoginLogs(ctx context.Context, page, limit int, filters map[string]interface{}, dest interface{}) error {
	key := acm.buildLoginLogsKey(page, limit, filters)
	return acm.Get(ctx, key, dest)
}

// SetLoginLogs 设置登录日志缓存
func (acm *AdminCacheManager) SetLoginLogs(ctx context.Context, page, limit int, filters map[string]interface{}, logs interface{}) error {
	key := acm.buildLoginLogsKey(page, limit, filters)
	return acm.Set(ctx, key, logs, AdminCacheTTL.LoginLogs)
}

// InvalidateLoginLogs 失效登录日志缓存
func (acm *AdminCacheManager) InvalidateLoginLogs(ctx context.Context) error {
	pattern := "admin:loginlog:*"
	return acm.DeletePattern(ctx, pattern)
}

// buildLoginLogsKey 构建登录日志缓存键
func (acm *AdminCacheManager) buildLoginLogsKey(page, limit int, filters map[string]interface{}) string {
	key := fmt.Sprintf("admin:loginlog:p%d:l%d", page, limit)

	if userID, ok := filters["userID"].(string); ok && userID != "" {
		key += fmt.Sprintf(":user:%s", userID)
	}
	if ip, ok := filters["ip"].(string); ok && ip != "" {
		key += fmt.Sprintf(":ip:%s", ip)
	}
	if success, ok := filters["success"].(bool); ok {
		if success {
			key += ":success"
		} else {
			key += ":failed"
		}
	}
	if startTime, ok := filters["startTime"].(string); ok && startTime != "" {
		key += fmt.Sprintf(":start:%s", startTime)
	}
	if endTime, ok := filters["endTime"].(string); ok && endTime != "" {
		key += fmt.Sprintf(":end:%s", endTime)
	}

	return key
}

// === 仪表盘统计缓存 ===

// GetDashboardStats 获取仪表盘统计缓存
func (acm *AdminCacheManager) GetDashboardStats(ctx context.Context, dest interface{}) error {
	key := "admin:dashboard:stats"
	return acm.Get(ctx, key, dest)
}

// SetDashboardStats 设置仪表盘统计缓存
func (acm *AdminCacheManager) SetDashboardStats(ctx context.Context, stats interface{}) error {
	key := "admin:dashboard:stats"
	return acm.Set(ctx, key, stats, AdminCacheTTL.Dashboard)
}

// InvalidateDashboardStats 失效仪表盘统计缓存
func (acm *AdminCacheManager) InvalidateDashboardStats(ctx context.Context) error {
	key := "admin:dashboard:stats"
	return acm.Delete(ctx, key)
}

// === 分析数据缓存 ===

// GetAnalyticsData 获取分析数据缓存
func (acm *AdminCacheManager) GetAnalyticsData(ctx context.Context, dataType, period string, dest interface{}) error {
	key := fmt.Sprintf("admin:analytics:%s:%s", dataType, period)
	return acm.Get(ctx, key, dest)
}

// SetAnalyticsData 设置分析数据缓存
func (acm *AdminCacheManager) SetAnalyticsData(ctx context.Context, dataType, period string, data interface{}) error {
	key := fmt.Sprintf("admin:analytics:%s:%s", dataType, period)
	return acm.Set(ctx, key, data, AdminCacheTTL.Analytics)
}

// InvalidateAnalyticsData 失效分析数据缓存
func (acm *AdminCacheManager) InvalidateAnalyticsData(ctx context.Context, dataType string) error {
	pattern := fmt.Sprintf("admin:analytics:%s:*", dataType)
	return acm.DeletePattern(ctx, pattern)
}

// === 系统信息缓存 ===

// GetSystemInfo 获取系统信息缓存
func (acm *AdminCacheManager) GetSystemInfo(ctx context.Context, dest interface{}) error {
	key := "admin:system:info"
	return acm.Get(ctx, key, dest)
}

// SetSystemInfo 设置系统信息缓存
func (acm *AdminCacheManager) SetSystemInfo(ctx context.Context, info interface{}) error {
	key := "admin:system:info"
	return acm.Set(ctx, key, info, AdminCacheTTL.SystemInfo)
}

// InvalidateSystemInfo 失效系统信息缓存
func (acm *AdminCacheManager) InvalidateSystemInfo(ctx context.Context) error {
	key := "admin:system:info"
	return acm.Delete(ctx, key)
}

// === 批量失效操作 ===

// InvalidateUserRelated 失效用户相关的所有缓存
func (acm *AdminCacheManager) InvalidateUserRelated(ctx context.Context, userID string) error {
	// 失效用户详情
	if err := acm.InvalidateUserDetail(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user detail: %w", err)
	}

	// 失效用户列表
	if err := acm.InvalidateUserList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate user list: %w", err)
	}

	// 失效登录日志
	if err := acm.InvalidateLoginLogs(ctx); err != nil {
		return fmt.Errorf("failed to invalidate login logs: %w", err)
	}

	// 失效仪表盘统计
	if err := acm.InvalidateDashboardStats(ctx); err != nil {
		return fmt.Errorf("failed to invalidate dashboard stats: %w", err)
	}

	return nil
}

// InvalidatePostRelated 失效文章相关的所有缓存
func (acm *AdminCacheManager) InvalidatePostRelated(ctx context.Context, postID string) error {
	// 失效文章详情
	if err := acm.InvalidatePostDetail(ctx, postID); err != nil {
		return fmt.Errorf("failed to invalidate post detail: %w", err)
	}

	// 失效文章列表
	if err := acm.InvalidatePostList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate post list: %w", err)
	}

	// 失效仪表盘统计
	if err := acm.InvalidateDashboardStats(ctx); err != nil {
		return fmt.Errorf("failed to invalidate dashboard stats: %w", err)
	}

	// 失效相关分析数据
	if err := acm.InvalidateAnalyticsData(ctx, "posts"); err != nil {
		return fmt.Errorf("failed to invalidate analytics data: %w", err)
	}

	return nil
}

// InvalidateTagRelated 失效标签相关的所有缓存
func (acm *AdminCacheManager) InvalidateTagRelated(ctx context.Context) error {
	// 失效标签列表
	if err := acm.InvalidateTagList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate tag list: %w", err)
	}

	// 失效文章列表（因为可能有标签过滤）
	if err := acm.InvalidatePostList(ctx); err != nil {
		return fmt.Errorf("failed to invalidate post list: %w", err)
	}

	return nil
}
