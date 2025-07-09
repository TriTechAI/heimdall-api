package constants

import "time"

// ====================
// 通用常量定义
// ====================

// 时间相关常量
const (
	// 默认时间格式
	DefaultTimeFormat = time.RFC3339

	// 评论时间窗口（分钟）
	CommentTimeoutMinutes = 5

	// 缓存过期时间
	CacheExpireShort  = 5 * time.Minute
	CacheExpireMedium = 30 * time.Minute
	CacheExpireLong   = 2 * time.Hour

	// 登录尝试限制
	MaxLoginAttempts   = 5
	LoginLockDuration  = 30 * time.Minute
	LoginAttemptWindow = 15 * time.Minute
)

// 分页相关常量
const (
	// 各类型默认分页大小
	CommentsPerPageDefault = 20
	UsersPerPageDefault    = 20
)

// 文本长度限制
const (
	// 评论相关
	CommentMinLength       = 1
	CommentMaxLength       = 1000
	CommentAuthorMaxLength = 50
	CommentEmailMaxLength  = 100
	CommentUrlMaxLength    = 255

	// SEO相关
	MetaTitleMaxLength       = 70
	MetaDescriptionMaxLength = 160
	CanonicalUrlMaxLength    = 255
)

// 文件相关常量
const (
	// 文件大小限制
	MaxAvatarSize = 2 * 1024 * 1024 // 2MB

	// 支持的文件类型
	AllowedImageTypes = "jpg,jpeg,png,gif,webp"
	AllowedFileTypes  = "jpg,jpeg,png,gif,webp,pdf,doc,docx,xls,xlsx"
)

// API相关常量
const (
	// API版本
	APIVersion = "v1"

	// 请求超时
	DefaultRequestTimeout = 30 * time.Second
	UploadRequestTimeout  = 5 * time.Minute

	// 速率限制
	RateLimitPerMinute = 60
	RateLimitPerHour   = 600
)

// 数据库相关常量
const (
	// 连接池配置
	DBMaxOpenConns    = 100
	DBMaxIdleConns    = 10
	DBConnMaxLifetime = 30 * time.Minute
	DBConnMaxIdleTime = 5 * time.Minute

	// 查询超时
	DBQueryTimeout  = 10 * time.Second
	DBSlowQueryTime = 1 * time.Second
)

// Redis相关常量
const (
	// Redis键前缀
	RedisKeyPrefix = "heimdall:"

	// 会话相关
	SessionKeyPrefix = "session:"
	SessionExpire    = 24 * time.Hour

	// JWT黑名单
	JWTBlacklistPrefix = "jwt:blacklist:"
	JWTBlacklistExpire = 24 * time.Hour

	// 限流相关
	RateLimitPrefix = "ratelimit:"
	RateLimitExpire = 1 * time.Hour
)

// 系统相关常量
const (
	// 默认语言和时区
	DefaultLanguage = "zh-CN"
	DefaultTimezone = "Asia/Shanghai"

	// 系统标识
	SystemName    = "Heimdall"
	SystemVersion = "1.0.0"

	// 默认值
	DefaultDomain = "blog.heimdall.com"
	DefaultEmail  = "admin@heimdall.com"
)

// 内容发现常量
const (
	// 热门文章统计天数
	PopularPostsDays  = 30
	TrendingPostsDays = 7

	// 默认返回数量
	PopularPostsLimit  = 10
	RecentPostsLimit   = 10
	TrendingPostsLimit = 10
	RelatedPostsLimit  = 5

	// 最大返回数量
	MaxPopularPosts  = 20
	MaxRecentPosts   = 20
	MaxTrendingPosts = 20
	MaxRelatedPosts  = 10
)

// 阅读时间计算常量
const (
	// 平均阅读速度（字/分钟）
	AverageReadingSpeed = 200

	// 最小阅读时间（分钟）
	MinReadingTime = 1
)
