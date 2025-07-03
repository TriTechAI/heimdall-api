package config

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

// Config public-api服务配置
type Config struct {
	rest.RestConf

	// 服务配置
	Service ServiceConfig `json:",optional"`

	// 数据库配置
	MongoDB MongoDBConfig `json:",optional"`

	// Redis配置
	Redis RedisConfig `json:",optional"`

	// CORS配置
	CORS CORSConfig `json:",optional"`

	// API限流配置
	RateLimit RateLimitConfig `json:",optional"`

	// 监控配置
	Monitoring MonitoringConfig `json:",optional"`

	// 业务配置
	Business BusinessConfig `json:",optional"`

	// 缓存配置
	Cache CacheConfig `json:",optional"`

	// SEO配置
	SEO SEOConfig `json:",optional"`

	// 安全配置
	Security SecurityConfig `json:",optional"`

	// 统计配置
	Analytics AnalyticsConfig `json:",optional"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Version     string `json:",default=1.0.0"`
	Environment string `json:",default=development"`
	LogLevel    string `json:",default=info"`
}

// MongoDBConfig MongoDB数据库配置
type MongoDBConfig struct {
	Host                   string `json:",default=localhost:27017"`
	Database               string `json:",default=heimdall_dev"`
	Username               string `json:",optional"`
	Password               string `json:",optional"`
	AuthSource             string `json:",default=admin"`
	MaxPoolSize            int    `json:",default=20"`
	MinPoolSize            int    `json:",default=2"`
	ConnectTimeout         int    `json:",default=10"`
	ServerSelectionTimeout int    `json:",default=10"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `json:",default=localhost:6379"`
	Password     string `json:",optional"`
	DB           int    `json:",default=1"`
	MaxRetries   int    `json:",default=3"`
	PoolSize     int    `json:",default=20"`
	MinIdleConns int    `json:",default=5"`
	DialTimeout  int    `json:",default=5"`
	ReadTimeout  int    `json:",default=5"`
	WriteTimeout int    `json:",default=5"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `json:",optional"`
	AllowMethods     []string `json:",optional"`
	AllowHeaders     []string `json:",optional"`
	ExposeHeaders    []string `json:",optional"`
	AllowCredentials bool     `json:",default=false"`
	MaxAge           int      `json:",default=86400"`
}

// RateLimitConfig API限流配置
type RateLimitConfig struct {
	Global GlobalRateLimit `json:",optional"`
	PerIP  PerIPRateLimit  `json:",optional"`
	Search SearchRateLimit `json:",optional"`
}

// GlobalRateLimit 全局限流
type GlobalRateLimit struct {
	Requests int `json:",default=1000"`
	Burst    int `json:",default=50"`
}

// PerIPRateLimit IP级别限流
type PerIPRateLimit struct {
	Requests int `json:",default=100"`
	Burst    int `json:",default=10"`
}

// SearchRateLimit 搜索API限流
type SearchRateLimit struct {
	Requests int `json:",default=20"`
	Burst    int `json:",default=5"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	EnableMetrics   bool   `json:",default=true"`
	MetricsPort     int    `json:",default=9091"`
	HealthPath      string `json:",default=/health"`
	EnableTracing   bool   `json:",default=false"`
	TracingEndpoint string `json:",optional"`
}

// BusinessConfig 业务配置
type BusinessConfig struct {
	DefaultPageSize        int  `json:",default=10"`
	MaxPageSize            int  `json:",default=50"`
	ExcerptLength          int  `json:",default=200"`
	RecentPostsLimit       int  `json:",default=5"`
	PopularPostsLimit      int  `json:",default=10"`
	SearchResultLimit      int  `json:",default=50"`
	SearchKeywordMinLength int  `json:",default=2"`
	CommentsPerPage        int  `json:",default=20"`
	MaxCommentLength       int  `json:",default=1000"`
	EnableCommentApproval  bool `json:",default=true"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	PostList     CacheItem `json:",optional"`
	PostDetail   CacheItem `json:",optional"`
	TagList      CacheItem `json:",optional"`
	SiteInfo     CacheItem `json:",optional"`
	SearchResult CacheItem `json:",optional"`
	PopularPosts CacheItem `json:",optional"`
}

// CacheItem 缓存项配置
type CacheItem struct {
	Prefix string `json:",optional"`
	TTL    int    `json:",default=3600"`
}

// SEOConfig SEO配置
type SEOConfig struct {
	SitemapPath      string `json:",default=/sitemap.xml"`
	SitemapCacheTime int    `json:",default=86400"`
	RSSPath          string `json:",default=/rss.xml"`
	RSSCacheTime     int    `json:",default=3600"`
	RSSItemLimit     int    `json:",default=20"`
	RobotsPath       string `json:",default=/robots.txt"`
	DefaultOGImage   string `json:",optional"`
	SiteName         string `json:",default=Heimdall Blog"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AntiBot         AntiBotConfig         `json:",optional"`
	ContentSecurity ContentSecurityConfig `json:",optional"`
}

// AntiBotConfig 防爬虫配置
type AntiBotConfig struct {
	EnableUACheck        bool `json:",default=true"`
	EnableRefererCheck   bool `json:",default=false"`
	MaxRequestsPerSecond int  `json:",default=10"`
	SuspiciousThreshold  int  `json:",default=5"`
}

// ContentSecurityConfig 内容安全配置
type ContentSecurityConfig struct {
	EnableXSSFilter bool `json:",default=true"`
	EnableSQLFilter bool `json:",default=true"`
	MaxURLLength    int  `json:",default=2048"`
}

// AnalyticsConfig 统计配置
type AnalyticsConfig struct {
	EnablePageViews              bool `json:",default=true"`
	PageViewCacheTime            int  `json:",default=300"`
	PopularContentUpdateInterval int  `json:",default=3600"`
	EnableUserBehavior           bool `json:",default=false"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证MongoDB配置
	if c.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database cannot be empty")
	}

	// 验证业务配置
	if c.Business.DefaultPageSize <= 0 || c.Business.DefaultPageSize > c.Business.MaxPageSize {
		return fmt.Errorf("default page size must be positive and <= max page size")
	}

	if c.Business.SearchKeywordMinLength < 1 {
		return fmt.Errorf("search keyword min length must be >= 1")
	}

	// 验证限流配置
	if c.RateLimit.Global.Requests <= 0 || c.RateLimit.Global.Burst <= 0 {
		return fmt.Errorf("rate limit requests and burst must be positive")
	}

	// 验证缓存配置
	if c.Cache.PostList.TTL < 0 || c.Cache.PostDetail.TTL < 0 {
		return fmt.Errorf("cache TTL must be non-negative")
	}

	return nil
}

// GetMongoDBURI 获取MongoDB连接URI
func (c *Config) GetMongoDBURI() string {
	if c.MongoDB.Username != "" && c.MongoDB.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
			c.MongoDB.Username, c.MongoDB.Password, c.MongoDB.Host, c.MongoDB.Database, c.MongoDB.AuthSource)
	}
	return fmt.Sprintf("mongodb://%s/%s", c.MongoDB.Host, c.MongoDB.Database)
}

// GetRedisAddr 获取Redis连接地址
func (c *Config) GetRedisAddr() string {
	return c.Redis.Host
}

// GetCacheTTL 获取缓存TTL时间
func (c *Config) GetCacheTTL(cacheType string) time.Duration {
	var ttl int
	switch cacheType {
	case "post_list":
		ttl = c.Cache.PostList.TTL
	case "post_detail":
		ttl = c.Cache.PostDetail.TTL
	case "tag_list":
		ttl = c.Cache.TagList.TTL
	case "site_info":
		ttl = c.Cache.SiteInfo.TTL
	case "search_result":
		ttl = c.Cache.SearchResult.TTL
	case "popular_posts":
		ttl = c.Cache.PopularPosts.TTL
	default:
		ttl = 3600 // 默认1小时
	}
	return time.Duration(ttl) * time.Second
}

// GetCacheKey 获取缓存键
func (c *Config) GetCacheKey(cacheType, key string) string {
	var prefix string
	switch cacheType {
	case "post_list":
		prefix = c.Cache.PostList.Prefix
	case "post_detail":
		prefix = c.Cache.PostDetail.Prefix
	case "tag_list":
		prefix = c.Cache.TagList.Prefix
	case "site_info":
		prefix = c.Cache.SiteInfo.Prefix
	case "search_result":
		prefix = c.Cache.SearchResult.Prefix
	case "popular_posts":
		prefix = c.Cache.PopularPosts.Prefix
	default:
		prefix = "cache:"
	}

	if prefix == "" {
		prefix = fmt.Sprintf("%s:", cacheType)
	}

	return prefix + key
}
