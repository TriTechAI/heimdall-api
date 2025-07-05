package config

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

// Config admin-api服务配置
type Config struct {
	rest.RestConf

	// go-zero JWT中间件需要的字段 - 按照官方文档标准
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 服务配置
	Service ServiceConfig `json:",optional"`

	// 数据库配置
	MongoDB MongoDBConfig `json:",optional"`

	// Redis配置
	Redis RedisConfig `json:",optional"`

	// JWT业务配置 (扩展配置，用于refresh token等)
	JWTBusiness JWTBusinessConfig `json:",optional"`

	// 安全配置
	Security SecurityConfig `json:",optional"`

	// CORS配置
	CORS CORSConfig `json:",optional"`

	// 文件存储配置
	Storage StorageConfig `json:",optional"`

	// 邮件配置
	Email EmailConfig `json:",optional"`

	// 监控配置
	Monitoring MonitoringConfig `json:",optional"`

	// 业务配置
	Business BusinessConfig `json:",optional"`

	// 缓存配置
	Cache CacheConfig `json:",optional"`
}

// JWTBusinessConfig JWT业务扩展配置
type JWTBusinessConfig struct {
	RefreshExpire int64 `json:",default=604800"` // 7天
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
	MaxPoolSize            int    `json:",default=10"`
	MinPoolSize            int    `json:",default=0"`
	ConnectTimeout         int    `json:",default=10"`
	ServerSelectionTimeout int    `json:",default=10"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `json:",default=localhost:6379"`
	Password     string `json:",optional"`
	DB           int    `json:",default=0"`
	MaxRetries   int    `json:",default=3"`
	PoolSize     int    `json:",default=10"`
	MinIdleConns int    `json:",default=1"`
	DialTimeout  int    `json:",default=5"`
	ReadTimeout  int    `json:",default=5"`
	WriteTimeout int    `json:",default=5"`
}

// 注意：AuthConfig 已移除，JWT配置现在直接在 Config.Auth 中定义
// RefreshExpire 现在在 JWTBusinessConfig 中定义

// SecurityConfig 安全配置
type SecurityConfig struct {
	BcryptCost           int             `json:",default=12"`
	MaxLoginAttempts     int             `json:",default=5"`
	LoginLockoutDuration int             `json:",default=1800"` // 30分钟
	RateLimit            RateLimitConfig `json:",optional"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Requests int `json:",default=100"` // 每分钟请求数
	Burst    int `json:",default=10"`  // 突发请求数
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `json:",optional"`
	AllowMethods     []string `json:",optional"`
	AllowHeaders     []string `json:",optional"`
	ExposeHeaders    []string `json:",optional"`
	AllowCredentials bool     `json:",default=true"`
	MaxAge           int      `json:",default=86400"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string       `json:",default=local"`
	Local LocalStorage `json:",optional"`
	COS   COSStorage   `json:",optional"`
}

// LocalStorage 本地存储配置
type LocalStorage struct {
	Path      string `json:",default=uploads"`
	URLPrefix string `json:",default=http://localhost:8080/uploads"`
}

// COSStorage 腾讯云COS配置
type COSStorage struct {
	SecretID  string `json:",optional"`
	SecretKey string `json:",optional"`
	Region    string `json:",default=ap-beijing"`
	Bucket    string `json:",optional"`
	URLPrefix string `json:",optional"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTP SMTPConfig `json:",optional"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host      string `json:",optional"`
	Port      int    `json:",default=587"`
	Username  string `json:",optional"`
	Password  string `json:",optional"`
	FromName  string `json:",default=Heimdall Blog"`
	FromEmail string `json:",optional"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	EnableMetrics   bool   `json:",default=true"`
	MetricsPort     int    `json:",default=9090"`
	HealthPath      string `json:",default=/health"`
	EnableTracing   bool   `json:",default=false"`
	TracingEndpoint string `json:",optional"`
}

// BusinessConfig 业务配置
type BusinessConfig struct {
	DefaultPageSize  int      `json:",default=10"`
	MaxPageSize      int      `json:",default=100"`
	MaxFileSize      int64    `json:",default=10485760"` // 10MB
	AllowedFileTypes []string `json:",optional"`
	MaxTitleLength   int      `json:",default=200"`
	MaxExcerptLength int      `json:",default=500"`
	MaxContentLength int      `json:",default=1048576"` // 1MB
}

// CacheConfig 缓存配置
type CacheConfig struct {
	JWTBlacklist  CacheItem `json:",optional"`
	UserSession   CacheItem `json:",optional"`
	LoginAttempts CacheItem `json:",optional"`
}

// CacheItem 缓存项配置
type CacheItem struct {
	Prefix string `json:",optional"`
	TTL    int    `json:",default=3600"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证MongoDB配置
	if c.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database cannot be empty")
	}

	// 验证JWT配置
	if c.Auth.AccessSecret == "" || c.Auth.AccessSecret == "your-secret-key" {
		return fmt.Errorf("auth access secret must be set and not use default value")
	}

	if c.Auth.AccessExpire <= 0 {
		return fmt.Errorf("auth access expire must be positive")
	}

	// 验证安全配置
	if c.Security.BcryptCost < 10 || c.Security.BcryptCost > 15 {
		return fmt.Errorf("bcrypt cost must be between 10 and 15")
	}

	// 验证业务配置
	if c.Business.DefaultPageSize <= 0 || c.Business.DefaultPageSize > c.Business.MaxPageSize {
		return fmt.Errorf("default page size must be positive and <= max page size")
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

// GetJWTExpireDuration 获取JWT过期时间
func (c *Config) GetJWTExpireDuration() time.Duration {
	return time.Duration(c.Auth.AccessExpire) * time.Second
}

// GetRefreshTokenExpireDuration 获取刷新令牌过期时间
func (c *Config) GetRefreshTokenExpireDuration() time.Duration {
	return time.Duration(c.JWTBusiness.RefreshExpire) * time.Second
}
