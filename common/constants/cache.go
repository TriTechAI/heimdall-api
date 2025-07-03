package constants

import "time"

// ====================
// 缓存键前缀常量
// ====================

const (
	// 服务前缀
	CachePrefix = "heimdall"

	// 模块前缀
	CachePrefixUser     = "user"
	CachePrefixAuth     = "auth"
	CachePrefixPost     = "post"
	CachePrefixComment  = "comment"
	CachePrefixMedia    = "media"
	CachePrefixSetting  = "setting"
	CachePrefixTag      = "tag"
	CachePrefixSession  = "session"
	CachePrefixSecurity = "security"
	CachePrefixStats    = "stats"
)

// ====================
// 用户相关缓存键
// ====================

const (
	// 用户信息缓存
	CacheKeyUserByID       = "heimdall:user:id:%s"       // 按ID缓存用户信息
	CacheKeyUserByUsername = "heimdall:user:username:%s" // 按用户名缓存用户信息
	CacheKeyUserByEmail    = "heimdall:user:email:%s"    // 按邮箱缓存用户信息
	CacheKeyUserList       = "heimdall:user:list:%s"     // 用户列表缓存（按条件）
	CacheKeyUserCount      = "heimdall:user:count"       // 用户总数缓存
)

// ====================
// 认证相关缓存键
// ====================

const (
	// JWT相关
	CacheKeyJWTBlacklist = "heimdall:auth:jwt:blacklist:%s" // JWT黑名单
	CacheKeyRefreshToken = "heimdall:auth:refresh:%s"       // 刷新令牌
	CacheKeyTokenUser    = "heimdall:auth:token:user:%s"    // 令牌对应的用户信息

	// 会话相关
	CacheKeyUserSession  = "heimdall:session:user:%s:%s" // 用户会话: user_id:session_id
	CacheKeyUserSessions = "heimdall:session:user:%s"    // 用户所有会话列表

	// 登录失败相关
	CacheKeyLoginFailCount = "heimdall:security:login:fail:%s" // 登录失败计数（按用户名/邮箱）
	CacheKeyLoginIPFail    = "heimdall:security:login:ip:%s"   // IP登录失败计数
	CacheKeyUserLock       = "heimdall:security:lock:%s"       // 用户锁定状态
	CacheKeyIPBlock        = "heimdall:security:block:ip:%s"   // IP封禁
)

// ====================
// 文章相关缓存键
// ====================

const (
	// 文章信息缓存
	CacheKeyPostByID       = "heimdall:post:id:%s"        // 按ID缓存文章
	CacheKeyPostBySlug     = "heimdall:post:slug:%s"      // 按Slug缓存文章
	CacheKeyPostList       = "heimdall:post:list:%s"      // 文章列表缓存
	CacheKeyPublishedPosts = "heimdall:post:published:%s" // 已发布文章列表
	CacheKeyPostsByAuthor  = "heimdall:post:author:%s:%s" // 按作者的文章列表
	CacheKeyPostsByTag     = "heimdall:post:tag:%s:%s"    // 按标签的文章列表
	CacheKeyPostCount      = "heimdall:post:count:%s"     // 文章数量统计

	// 文章内容缓存
	CacheKeyPostHTML    = "heimdall:post:html:%s"    // 渲染后的HTML
	CacheKeyPostTOC     = "heimdall:post:toc:%s"     // 文章目录
	CacheKeyPostSummary = "heimdall:post:summary:%s" // 文章摘要

	// 热门文章
	CacheKeyPopularPosts = "heimdall:post:popular"    // 热门文章列表
	CacheKeyRecentPosts  = "heimdall:post:recent"     // 最新文章列表
	CacheKeyRelatedPosts = "heimdall:post:related:%s" // 相关文章列表
)

// ====================
// 评论相关缓存键
// ====================

const (
	// 评论信息缓存
	CacheKeyCommentByID    = "heimdall:comment:id:%s"      // 按ID缓存评论
	CacheKeyCommentsByPost = "heimdall:comment:post:%s:%s" // 按文章缓存评论列表
	CacheKeyCommentTree    = "heimdall:comment:tree:%s"    // 评论树结构
	CacheKeyCommentCount   = "heimdall:comment:count:%s"   // 评论数量统计

	// 评论状态缓存
	CacheKeyPendingComments = "heimdall:comment:pending" // 待审核评论
	CacheKeyRecentComments  = "heimdall:comment:recent"  // 最新评论
)

// ====================
// 标签相关缓存键
// ====================

const (
	// 标签信息缓存
	CacheKeyTagByID     = "heimdall:tag:id:%s"   // 按ID缓存标签
	CacheKeyTagBySlug   = "heimdall:tag:slug:%s" // 按Slug缓存标签
	CacheKeyTagList     = "heimdall:tag:list"    // 标签列表
	CacheKeyTagCloud    = "heimdall:tag:cloud"   // 标签云
	CacheKeyTagCount    = "heimdall:tag:count"   // 标签总数
	CacheKeyPopularTags = "heimdall:tag:popular" // 热门标签
)

// ====================
// 媒体文件相关缓存键
// ====================

const (
	// 媒体信息缓存
	CacheKeyMediaByID   = "heimdall:media:id:%s"   // 按ID缓存媒体信息
	CacheKeyMediaList   = "heimdall:media:list:%s" // 媒体列表
	CacheKeyMediaCount  = "heimdall:media:count"   // 媒体文件总数
	CacheKeyRecentMedia = "heimdall:media:recent"  // 最新媒体文件
)

// ====================
// 站点设置相关缓存键
// ====================

const (
	// 站点设置缓存
	CacheKeySetting         = "heimdall:setting:%s"       // 单个设置项
	CacheKeyAllSettings     = "heimdall:setting:all"      // 所有设置
	CacheKeySettingsByGroup = "heimdall:setting:group:%s" // 按分组的设置

	// 站点信息缓存
	CacheKeySiteInfo   = "heimdall:site:info"       // 站点基本信息
	CacheKeyNavigation = "heimdall:site:navigation" // 导航菜单
	CacheKeySitemap    = "heimdall:site:sitemap"    // 站点地图
	CacheKeyRSSFeed    = "heimdall:site:rss"        // RSS订阅
)

// ====================
// 统计数据相关缓存键
// ====================

const (
	// 访问统计
	CacheKeyPostViewCount = "heimdall:stats:post:view:%s" // 文章浏览量
	CacheKeyDailyViews    = "heimdall:stats:daily:%s"     // 每日浏览量
	CacheKeyMonthlyViews  = "heimdall:stats:monthly:%s"   // 每月浏览量

	// 用户活跃度
	CacheKeyActiveUsers  = "heimdall:stats:users:active" // 活跃用户
	CacheKeyUserActivity = "heimdall:stats:user:%s"      // 用户活动统计

	// 内容统计
	CacheKeyContentStats   = "heimdall:stats:content"   // 内容统计
	CacheKeyDashboardStats = "heimdall:stats:dashboard" // 仪表盘统计
)

// ====================
// 缓存TTL时间常量
// ====================

const (
	// 短期缓存 (5分钟以内)
	CacheTTLVeryShort = 1 * time.Minute // 1分钟
	CacheTTLShort     = 5 * time.Minute // 5分钟

	// 中期缓存 (1小时以内)
	CacheTTLMedium   = 15 * time.Minute // 15分钟
	CacheTTLHalfHour = 30 * time.Minute // 30分钟
	CacheTTLHour     = 1 * time.Hour    // 1小时

	// 长期缓存 (1天以内)
	CacheTTL6Hours  = 6 * time.Hour  // 6小时
	CacheTTL12Hours = 12 * time.Hour // 12小时
	CacheTTLDay     = 24 * time.Hour // 1天

	// 特殊缓存时间
	CacheTTLWeek  = 7 * 24 * time.Hour  // 1周
	CacheTTLMonth = 30 * 24 * time.Hour // 1月
)

// ====================
// 安全相关缓存TTL
// ====================

const (
	// 认证相关TTL
	CacheTTLJWTBlacklist = 2 * time.Hour      // JWT黑名单缓存时间
	CacheTTLSession      = 2 * time.Hour      // 会话缓存时间
	CacheTTLRefreshToken = 7 * 24 * time.Hour // 刷新令牌缓存时间

	// 安全防护TTL
	CacheTTLLoginFail = 15 * time.Minute // 登录失败计数缓存时间
	CacheTTLUserLock  = 24 * time.Hour   // 用户锁定状态缓存时间
	CacheTTLIPBlock   = 1 * time.Hour    // IP封禁缓存时间
	CacheTTLRateLimit = 1 * time.Minute  // 限流缓存时间
)

// ====================
// 内容相关缓存TTL
// ====================

const (
	// 动态内容TTL
	CacheTTLUserInfo    = CacheTTLMedium // 用户信息
	CacheTTLPostContent = CacheTTLHour   // 文章内容
	CacheTTLCommentList = CacheTTLMedium // 评论列表

	// 静态内容TTL
	CacheTTLPostList     = CacheTTLHour // 文章列表
	CacheTTLTagList      = CacheTTLDay  // 标签列表
	CacheTTLSiteSettings = CacheTTLDay  // 站点设置
	CacheTTLNavigation   = CacheTTLDay  // 导航菜单

	// 统计数据TTL
	CacheTTLViewStats    = CacheTTLMedium // 浏览统计
	CacheTTLDashboard    = CacheTTLMedium // 仪表盘数据
	CacheTTLContentStats = CacheTTLHour   // 内容统计
)

// ====================
// 缓存键生成函数
// ====================

// FormatCacheKey 格式化缓存键
func FormatCacheKey(template string, args ...interface{}) string {
	if len(args) == 0 {
		return template
	}
	// 这里简化处理，实际使用时建议使用 fmt.Sprintf
	return template
}

// GenerateUserCacheKey 生成用户相关缓存键
func GenerateUserCacheKey(keyType, identifier string) string {
	switch keyType {
	case "id":
		return CacheKeyUserByID
	case "username":
		return CacheKeyUserByUsername
	case "email":
		return CacheKeyUserByEmail
	default:
		return ""
	}
}

// GeneratePostCacheKey 生成文章相关缓存键
func GeneratePostCacheKey(keyType, identifier string) string {
	switch keyType {
	case "id":
		return CacheKeyPostByID
	case "slug":
		return CacheKeyPostBySlug
	case "html":
		return CacheKeyPostHTML
	default:
		return ""
	}
}

// GetCacheTTL 获取缓存TTL时间
func GetCacheTTL(cacheType string) time.Duration {
	switch cacheType {
	case "user_info":
		return CacheTTLUserInfo
	case "post_content":
		return CacheTTLPostContent
	case "post_list":
		return CacheTTLPostList
	case "comment_list":
		return CacheTTLCommentList
	case "tag_list":
		return CacheTTLTagList
	case "site_settings":
		return CacheTTLSiteSettings
	case "navigation":
		return CacheTTLNavigation
	case "session":
		return CacheTTLSession
	case "jwt_blacklist":
		return CacheTTLJWTBlacklist
	case "login_fail":
		return CacheTTLLoginFail
	case "user_lock":
		return CacheTTLUserLock
	case "rate_limit":
		return CacheTTLRateLimit
	default:
		return CacheTTLMedium // 默认15分钟
	}
}

// IsCacheKeyPattern 检查是否为有效的缓存键模式
func IsCacheKeyPattern(key string) bool {
	// 简单检查是否包含项目前缀
	return len(key) > len(CachePrefix) && key[:len(CachePrefix)] == CachePrefix
}
