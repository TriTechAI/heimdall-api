package client

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConfig MongoDB连接配置
type MongoConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username,optional"`
	Password    string `json:"password,optional"`
	AuthSource  string `json:"authSource,optional"`
	MaxPoolSize uint64 `json:"maxPoolSize,optional"`
	MinPoolSize uint64 `json:"minPoolSize,optional"`
	Timeout     int    `json:"timeout,optional"`
}

// MongoClient MongoDB客户端封装
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   MongoConfig
}

// NewMongoClient 创建新的MongoDB客户端
func NewMongoClient(config MongoConfig) (*MongoClient, error) {
	// 设置默认值
	if config.MaxPoolSize == 0 {
		config.MaxPoolSize = 10
	}
	if config.MinPoolSize == 0 {
		config.MinPoolSize = 0
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}
	if config.AuthSource == "" {
		config.AuthSource = "admin"
	}

	// 构建连接URI
	uri := buildMongoURI(config)

	// 配置客户端选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize).
		SetConnectTimeout(time.Duration(config.Timeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(config.Timeout) * time.Second)

	// 创建客户端
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &MongoClient{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// GetDatabase 获取数据库实例
func (mc *MongoClient) GetDatabase() *mongo.Database {
	return mc.database
}

// GetClient 获取客户端实例
func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}

// GetCollection 获取集合
func (mc *MongoClient) GetCollection(name string) *mongo.Collection {
	return mc.database.Collection(name)
}

// Ping 检查连接是否正常
func (mc *MongoClient) Ping(ctx context.Context) error {
	return mc.client.Ping(ctx, nil)
}

// Close 关闭连接
func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

// IsHealthy 检查MongoDB健康状态
func (mc *MongoClient) IsHealthy(ctx context.Context) bool {
	// 设置较短的超时时间用于健康检查
	healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// 执行ping命令
	err := mc.client.Ping(healthCtx, nil)
	if err != nil {
		return false
	}

	// 检查数据库状态
	result := mc.database.RunCommand(healthCtx, bson.D{bson.E{Key: "serverStatus", Value: 1}})
	if result.Err() != nil {
		return false
	}

	return true
}

// GetStats 获取数据库统计信息
func (mc *MongoClient) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取数据库统计
	result := mc.database.RunCommand(ctx, bson.D{bson.E{Key: "dbStats", Value: 1}})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var dbStats bson.M
	if err := result.Decode(&dbStats); err != nil {
		return nil, err
	}

	stats["database"] = dbStats

	// 获取服务器状态
	result = mc.database.RunCommand(ctx, bson.D{bson.E{Key: "serverStatus", Value: 1}})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var serverStats bson.M
	if err := result.Decode(&serverStats); err != nil {
		return nil, err
	}

	// 提取关键指标
	stats["connections"] = serverStats["connections"]
	stats["uptime"] = serverStats["uptime"]
	stats["version"] = serverStats["version"]

	return stats, nil
}

// buildMongoURI 构建MongoDB连接URI
func buildMongoURI(config MongoConfig) string {
	var uri string

	if config.Username != "" && config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.AuthSource)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s", config.Host, config.Port, config.Database)
	}

	return uri
}
