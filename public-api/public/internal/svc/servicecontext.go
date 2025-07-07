package svc

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/heimdall-api/common/cache"
	"github.com/heimdall-api/common/client"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/public-api/public/internal/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServiceContext struct {
	Config             config.Config
	MongoDB            *mongo.Database
	Redis              *redis.Client
	PostDAO            *dao.PostDAO
	UserDAO            *dao.UserDAO
	PageDAO            *dao.PageDAO
	TagDAO             *dao.TagDAO
	ContentCacheManager *cache.ContentCacheManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 解析MongoDB Host和Port
	mongoHost := "localhost"
	mongoPort := 27017
	if c.MongoDB.Host != "" {
		parts := strings.Split(c.MongoDB.Host, ":")
		if len(parts) >= 1 {
			mongoHost = parts[0]
		}
		if len(parts) >= 2 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				mongoPort = p
			}
		}
	}

	// 构建MongoDB配置
	mongoConfig := client.MongoConfig{
		Host:        mongoHost,
		Port:        mongoPort,
		Database:    c.MongoDB.Database,
		Username:    c.MongoDB.Username,
		Password:    c.MongoDB.Password,
		AuthSource:  c.MongoDB.AuthSource,
		MaxPoolSize: uint64(c.MongoDB.MaxPoolSize),
		MinPoolSize: uint64(c.MongoDB.MinPoolSize),
		Timeout:     c.MongoDB.ConnectTimeout,
	}

	// 初始化MongoDB连接
	mongoClient, err := client.NewMongoClient(mongoConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// 测试MongoDB连接
	err = mongoClient.Ping(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// 获取数据库实例
	database := mongoClient.GetDatabase()

	// 构建Redis配置
	redisConfig := client.RedisConfig{
		Host:         c.Redis.Host,
		Port:         c.Redis.Port,
		Password:     c.Redis.Password,
		Database:     c.Redis.DB,
		MaxRetries:   c.Redis.MaxRetries,
		PoolSize:     c.Redis.PoolSize,
		MinIdleConns: c.Redis.MinIdleConns,
		Timeout:      c.Redis.DialTimeout,
	}

	// 初始化Redis连接
	redisClient, err := client.NewRedisClient(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// 测试Redis连接
	err = redisClient.Ping(context.Background())
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	// 初始化DAO层
	postDAO := dao.NewPostDAO(database)
	userDAO := dao.NewUserDAO(database)
	pageDAO := dao.NewPageDAO(database)
	tagDAO := dao.NewTagDAO(database)

	// 初始化缓存管理器
	contentCacheManager := cache.NewContentCacheManager(redisClient.GetClient(), "public")

	return &ServiceContext{
		Config:             c,
		MongoDB:            database,
		Redis:              redisClient.GetClient(),
		PostDAO:            postDAO,
		UserDAO:            userDAO,
		PageDAO:            pageDAO,
		TagDAO:             tagDAO,
		ContentCacheManager: contentCacheManager,
	}
}
