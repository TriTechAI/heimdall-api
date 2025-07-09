package svc

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/heimdall-api/admin-api/admin/internal/config"
	"github.com/heimdall-api/admin-api/admin/internal/middleware"
	"github.com/heimdall-api/common/cache"
	"github.com/heimdall-api/common/dao"
)

type ServiceContext struct {
	Config            config.Config
	MongoDB           *mongo.Database
	Redis             *redis.Client
	UserDAO           *dao.UserDAO
	LoginLogDAO       *dao.LoginLogDAO
	PostDAO           *dao.PostDAO
	PageDAO           *dao.PageDAO
	TagDAO            *dao.TagDAO
	CommentDAO        *dao.CommentDAO
	AdminCacheManager *cache.AdminCacheManager
	CacheInvalidator  *cache.CacheInvalidator // 缓存失效器

	// 中间件
	JWTBlacklistMiddleware *middleware.JWTBlacklistMiddleware
	IPRateLimitMiddleware  *middleware.IPRateLimitMiddleware
	AuditMiddleware        *middleware.AuditMiddleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化MongoDB客户端
	mongoClient := initMongoDB(c)
	mongoDB := mongoClient.Database(c.MongoDB.Database)

	// 初始化Redis客户端
	redisClient := initRedis(c)

	// 初始化DAO
	userDAO := dao.NewUserDAO(mongoDB)
	loginLogDAO := dao.NewLoginLogDAO(mongoDB)
	postDAO := dao.NewPostDAO(mongoDB)
	pageDAO := dao.NewPageDAO(mongoDB)
	tagDAO := dao.NewTagDAO(mongoDB)
	commentDAO := dao.NewCommentDAO(mongoDB)

	// TODO: 在T123后续版本中实现AdminCacheManager集成
	// 目前暂时不初始化，避免Redis版本冲突
	// adminCacheManager := cache.NewAdminCacheManager(redisClient, "admin")

	// 初始化缓存失效器
	cacheInvalidator := cache.NewCacheInvalidator(redisClient, "public")

	// 初始化中间件
	var jwtBlacklistMiddleware *middleware.JWTBlacklistMiddleware
	var ipRateLimitMiddleware *middleware.IPRateLimitMiddleware
	var auditMiddleware *middleware.AuditMiddleware

	// JWT黑名单中间件
	if c.Middleware.JWTBlacklist.Enabled {
		jwtBlacklistMiddleware = middleware.NewJWTBlacklistMiddleware(
			redisClient,
			c.Auth.AccessSecret,
			"heimdall-admin",
		)
	}

	// IP限流中间件
	if c.Middleware.RateLimit.Enabled {
		rateLimitConfig := middleware.RateLimitConfig{
			GeneralRPS:   c.Middleware.RateLimit.GeneralRPS,
			GeneralBurst: c.Middleware.RateLimit.GeneralBurst,
			LoginRPS:     c.Middleware.RateLimit.LoginRPS,
			LoginBurst:   c.Middleware.RateLimit.LoginBurst,
			CreateRPS:    c.Middleware.RateLimit.CreateRPS,
			CreateBurst:  c.Middleware.RateLimit.CreateBurst,
			Window:       time.Minute,     // 固定1分钟窗口
			LoginWindow:  5 * time.Minute, // 登录5分钟窗口
			CreateWindow: time.Minute,     // 创建操作1分钟窗口
		}
		ipRateLimitMiddleware = middleware.NewIPRateLimitMiddleware(redisClient, rateLimitConfig)
	}

	// 操作审计中间件
	if c.Middleware.Audit.Enabled {
		auditMiddleware = middleware.NewAuditMiddleware(redisClient)
	}

	return &ServiceContext{
		Config:                 c,
		MongoDB:                mongoDB,
		Redis:                  redisClient,
		UserDAO:                userDAO,
		LoginLogDAO:            loginLogDAO,
		PostDAO:                postDAO,
		PageDAO:                pageDAO,
		TagDAO:                 tagDAO,
		CommentDAO:             commentDAO,
		AdminCacheManager:      nil, // TODO: 待后续实现
		CacheInvalidator:       cacheInvalidator,
		JWTBlacklistMiddleware: jwtBlacklistMiddleware,
		IPRateLimitMiddleware:  ipRateLimitMiddleware,
		AuditMiddleware:        auditMiddleware,
	}
}

// initMongoDB 初始化MongoDB连接
func initMongoDB(c config.Config) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.MongoDB.ConnectTimeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(c.GetMongoDBURI())
	clientOptions.SetMaxPoolSize(uint64(c.MongoDB.MaxPoolSize))
	clientOptions.SetMinPoolSize(uint64(c.MongoDB.MinPoolSize))
	clientOptions.SetServerSelectionTimeout(time.Duration(c.MongoDB.ServerSelectionTimeout) * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Printf("Successfully connected to MongoDB: %s", c.MongoDB.Host)
	return client
}

// initRedis 初始化Redis连接
func initRedis(c config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Host,
		Password:     c.Redis.Password,
		DB:           c.Redis.DB,
		MaxRetries:   c.Redis.MaxRetries,
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdleConns,
		DialTimeout:  time.Duration(c.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(c.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.Redis.WriteTimeout) * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Successfully connected to Redis: %s", c.Redis.Host)
	return rdb
}
