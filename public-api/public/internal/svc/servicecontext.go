package svc

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/heimdall-api/common/client"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/public-api/public/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServiceContext struct {
	Config  config.Config
	MongoDB *mongo.Database
	PostDAO *dao.PostDAO
	UserDAO *dao.UserDAO
	PageDAO *dao.PageDAO
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 解析Host和Port
	host := "localhost"
	port := 27017
	if c.MongoDB.Host != "" {
		parts := strings.Split(c.MongoDB.Host, ":")
		if len(parts) >= 1 {
			host = parts[0]
		}
		if len(parts) >= 2 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				port = p
			}
		}
	}

	// 构建MongoDB配置
	mongoConfig := client.MongoConfig{
		Host:        host,
		Port:        port,
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

	// 初始化DAO层
	postDAO := dao.NewPostDAO(database)
	userDAO := dao.NewUserDAO(database)
	pageDAO := dao.NewPageDAO(database)

	return &ServiceContext{
		Config:  c,
		MongoDB: database,
		PostDAO: postDAO,
		UserDAO: userDAO,
		PageDAO: pageDAO,
	}
}
