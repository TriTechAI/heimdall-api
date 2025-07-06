package constants

// ====================
// 分页限制常量
// ====================

const (
	// 默认分页参数
	DefaultPage  = 1
	DefaultLimit = 10

	// 分页限制
	MinPage      = 1
	MinLimit     = 1
	MaxLimit     = 100
	MaxLimitLogs = 100 // 日志查询的最大限制

	// 特殊查询限制
	MaxSearchLimit    = 50   // 搜索结果最大限制
	MaxRecentLimit    = 20   // 最近记录最大限制
	MaxPopularLimit   = 10   // 热门内容最大限制
	MaxRelatedLimit   = 5    // 相关内容最大限制
	MaxFailedLogLimit = 1000 // 失败日志查询最大限制
)

// ====================
// 内容长度限制常量
// ====================

const (
	// 用户相关长度限制
	MinUsernameLength    = 3
	MaxUsernameLength    = 32
	MinPasswordLength    = 6
	MaxPasswordLength    = 128
	MaxDisplayNameLength = 64
	MaxEmailLength       = 255
	MaxBioLength         = 500
	MaxLocationLength    = 100
	MaxWebsiteLength     = 255

	// 文章相关长度限制
	MinTitleLength        = 1
	MaxTitleLength        = 255
	MaxSlugLength         = 255
	MaxExcerptLength      = 500
	MaxMetaTitleLength    = 70
	MaxMetaDescLength     = 160
	MaxCanonicalURLLength = 255
	MaxContentLength      = 1000000 // 1MB

	// 评论相关长度限制
	MinCommentLength = 1
	MaxCommentLength = 2000
	MaxCommentDepth  = 5 // 评论嵌套最大深度

	// 标签相关长度限制
	MinTagNameLength = 1
	MaxTagNameLength = 50
	MaxTagSlugLength = 50
	MaxTagDescLength = 200
	MaxTagsPerPost   = 10

	// 媒体文件限制
	MaxFileSize       = 10 * 1024 * 1024 // 10MB
	MaxImageSize      = 5 * 1024 * 1024  // 5MB
	MaxFilenameLength = 255
)

// ====================
// 安全限制常量
// ====================

const (
	// 登录安全限制
	DefaultMaxLoginAttempts     = 5
	DefaultLoginLockoutDuration = 1800  // 30分钟，单位：秒
	MaxLoginLockoutDuration     = 86400 // 24小时，单位：秒

	// 密码安全限制
	MinPasswordComplexity = 1  // 最小复杂度要求
	MaxPasswordAge        = 90 // 密码最大使用天数

	// 会话限制
	MaxSessionsPerUser = 10 // 每个用户最大会话数

	// API限流限制
	DefaultRateLimit      = 100 // 每分钟请求数
	DefaultBurstLimit     = 10  // 突发请求数
	MaxRateLimitPerMinute = 1000
	MaxRateLimitPerHour   = 10000
	MaxRateLimitPerDay    = 100000

	// IP限制
	MaxIPFailAttempts = 20   // IP最大失败尝试次数
	IPBlockDuration   = 3600 // IP封禁时长（秒）
)

// ====================
// 数据库限制常量
// ====================

const (
	// MongoDB文档限制
	MaxDocumentSize = 16 * 1024 * 1024 // 16MB MongoDB文档大小限制

	// 批量操作限制
	MaxBatchSize       = 1000 // 批量操作最大数量
	MaxBulkWriteSize   = 100  // 批量写入最大数量
	MaxAggregateStages = 20   // 聚合管道最大阶段数

	// 索引限制
	MaxIndexesPerCollection = 64   // 每个集合最大索引数
	MaxIndexKeyLength       = 1024 // 索引键最大长度

	// 查询限制
	MaxQueryTimeout    = 30 // 查询超时时间（秒）
	MaxSortFields      = 10 // 排序字段最大数量
	MaxProjectionDepth = 5  // 投影深度限制
)

// ====================
// 缓存限制常量
// ====================

const (
	// Redis缓存限制
	MaxCacheKeyLength   = 250         // 缓存键最大长度
	MaxCacheValueSize   = 1024 * 1024 // 缓存值最大大小 1MB
	MaxCacheExpireTime  = 86400 * 30  // 最大过期时间 30天
	DefaultCacheTimeout = 300         // 默认缓存超时 5分钟

	// 缓存池限制
	MaxCacheConnections = 100 // 最大缓存连接数
	MaxIdleConnections  = 10  // 最大空闲连接数
	CacheConnectTimeout = 5   // 连接超时时间（秒）
)

// ====================
// 文件上传限制常量
// ====================

const (
	// 支持的文件类型
	MaxUploadSize = 10 * 1024 * 1024 // 10MB

	// 图片限制
	MaxImageWidth  = 2048 // 最大图片宽度
	MaxImageHeight = 2048 // 最大图片高度
	MaxImageCount  = 20   // 单次最大上传图片数

	// 文档限制
	MaxDocumentPages = 100 // 文档最大页数

	// 存储限制
	MaxStoragePerUser = 100 * 1024 * 1024       // 每用户最大存储 100MB
	MaxTotalStorage   = 10 * 1024 * 1024 * 1024 // 总存储限制 10GB
)

// ====================
// 业务逻辑限制常量
// ====================

const (
	// 内容发布限制
	MaxPostsPerDay    = 50  // 每日最大发布文章数
	MaxCommentsPerDay = 100 // 每日最大评论数
	MaxTagsPerDay     = 20  // 每日最大创建标签数

	// 用户行为限制
	MaxFollowsPerDay = 50  // 每日最大关注数
	MaxLikesPerDay   = 200 // 每日最大点赞数
	MaxSharesPerDay  = 100 // 每日最大分享数
	MaxReportsPerDay = 10  // 每日最大举报数

	// 搜索限制
	MinSearchKeywordLength = 2   // 搜索关键词最小长度
	MaxSearchKeywordLength = 100 // 搜索关键词最大长度
	MaxSearchResults       = 100 // 搜索结果最大数量
	SearchCacheTimeout     = 300 // 搜索缓存超时（秒）
)

// ====================
// 系统性能限制常量
// ====================

const (
	// 并发限制
	MaxConcurrentRequests = 1000 // 最大并发请求数
	MaxGoroutines         = 100  // 最大协程数
	MaxChannelBuffer      = 1000 // 通道缓冲区大小

	// 内存限制
	MaxMemoryUsage     = 512 * 1024 * 1024 // 最大内存使用 512MB
	MaxHeapSize        = 256 * 1024 * 1024 // 最大堆大小 256MB
	GCTargetPercentage = 100               // GC目标百分比

	// 网络限制
	MaxConnectionsPerIP = 100              // 每IP最大连接数
	MaxRequestSize      = 10 * 1024 * 1024 // 最大请求大小 10MB
	MaxResponseSize     = 50 * 1024 * 1024 // 最大响应大小 50MB
	NetworkTimeout      = 30               // 网络超时时间（秒）
)

// ====================
// 验证函数
// ====================

// ValidatePageLimit 验证分页参数
func ValidatePageLimit(page, limit int) (int, int) {
	if page < MinPage {
		page = DefaultPage
	}
	if limit < MinLimit {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return page, limit
}

// ValidateLogPageLimit 验证日志分页参数
func ValidateLogPageLimit(page, limit int) (int, int) {
	if page < MinPage {
		page = DefaultPage
	}
	if limit < MinLimit {
		limit = DefaultLimit
	}
	if limit > MaxLimitLogs {
		limit = MaxLimitLogs
	}
	return page, limit
}

// ValidateStringLength 验证字符串长度
func ValidateStringLength(str string, minLen, maxLen int) bool {
	length := len(str)
	return length >= minLen && length <= maxLen
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) bool {
	return ValidateStringLength(username, MinUsernameLength, MaxUsernameLength)
}

// ValidatePassword 验证密码长度
func ValidatePassword(password string) bool {
	return ValidateStringLength(password, MinPasswordLength, MaxPasswordLength)
}

// ValidateEmail 验证邮箱长度
func ValidateEmail(email string) bool {
	return ValidateStringLength(email, 1, MaxEmailLength)
}

// ValidateTitle 验证标题长度
func ValidateTitle(title string) bool {
	return ValidateStringLength(title, MinTitleLength, MaxTitleLength)
}

// ValidateContent 验证内容长度
func ValidateContent(content string) bool {
	return len(content) <= MaxContentLength
}
