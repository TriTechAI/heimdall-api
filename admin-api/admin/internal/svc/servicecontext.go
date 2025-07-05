package svc

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/heimdall-api/admin-api/admin/internal/config"
	"github.com/heimdall-api/common/dao"
)

type ServiceContext struct {
	Config      config.Config
	MongoDB     *mongo.Database
	Redis       *redis.Client
	UserDAO     *dao.UserDAO
	LoginLogDAO *dao.LoginLogDAO
	PostDAO     *dao.PostDAO
	PageDAO     *dao.PageDAO
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

	return &ServiceContext{
		Config:      c,
		MongoDB:     mongoDB,
		Redis:       redisClient,
		UserDAO:     userDAO,
		LoginLogDAO: loginLogDAO,
		PostDAO:     postDAO,
		PageDAO:     pageDAO,
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
