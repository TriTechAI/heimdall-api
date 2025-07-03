package dao

import (
	"context"
	"errors"
	"time"

	"github.com/heimdall-api/common/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LoginLogDAO 登录日志数据访问层
type LoginLogDAO struct {
	collection *mongo.Collection
}

// NewLoginLogDAO 创建登录日志DAO实例
func NewLoginLogDAO(database *mongo.Database) *LoginLogDAO {
	return &LoginLogDAO{
		collection: database.Collection("loginLogs"),
	}
}

// Create 创建登录日志
func (d *LoginLogDAO) Create(ctx context.Context, log *model.LoginLog) error {
	if log == nil {
		return errors.New("log cannot be nil")
	}

	// 验证登录日志数据
	if err := log.ValidateForCreate(); err != nil {
		return err
	}

	// 准备插入数据
	log.PrepareForInsert()

	_, err := d.collection.InsertOne(ctx, log)
	if err != nil {
		return err
	}

	return nil
}

// List 获取登录日志列表
func (d *LoginLogDAO) List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]*model.LoginLog, int64, error) {
	// 参数验证
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// 构建查询条件
	query := d.buildQueryFilter(filter)

	// 获取总数
	total, err := d.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 计算跳过的文档数
	skip := (page - 1) * limit

	// 构建排序条件
	sort := d.buildSortCondition(filter)

	// 构建查询选项
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(sort)

	cursor, err := d.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []*model.LoginLog
	for cursor.Next(ctx) {
		var log model.LoginLog
		if err := cursor.Decode(&log); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserID 根据用户ID获取登录日志
func (d *LoginLogDAO) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*model.LoginLog, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("userID cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, errors.New("invalid userID format")
	}

	filter := map[string]interface{}{
		"userId": objectID,
	}

	return d.List(ctx, filter, page, limit)
}

// GetByIPAddress 根据IP地址获取登录日志
func (d *LoginLogDAO) GetByIPAddress(ctx context.Context, ipAddress string, page, limit int) ([]*model.LoginLog, int64, error) {
	if ipAddress == "" {
		return nil, 0, errors.New("ipAddress cannot be empty")
	}

	filter := map[string]interface{}{
		"ipAddress": ipAddress,
	}

	return d.List(ctx, filter, page, limit)
}

// GetRecentFailedLogins 获取最近的失败登录记录
func (d *LoginLogDAO) GetRecentFailedLogins(ctx context.Context, since time.Time, limit int) ([]*model.LoginLog, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query := bson.M{
		"status":  "failed",
		"loginAt": bson.M{"$gte": since},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{bson.E{Key: "loginAt", Value: -1}})

	cursor, err := d.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*model.LoginLog
	for cursor.Next(ctx) {
		var log model.LoginLog
		if err := cursor.Decode(&log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// CreateIndexes 创建登录日志集合的索引
func (d *LoginLogDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				bson.E{Key: "userId", Value: 1},
				bson.E{Key: "loginAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				bson.E{Key: "ipAddress", Value: 1},
				bson.E{Key: "loginAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				bson.E{Key: "status", Value: 1},
				bson.E{Key: "loginAt", Value: -1},
			},
		},
		{
			Keys: bson.D{bson.E{Key: "loginAt", Value: -1}},
		},
		{
			Keys: bson.D{bson.E{Key: "username", Value: 1}},
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// buildQueryFilter 构建查询过滤条件
func (d *LoginLogDAO) buildQueryFilter(filter map[string]interface{}) bson.M {
	query := bson.M{}

	if filter == nil {
		return query
	}

	for key, value := range filter {
		switch key {
		case "userId":
			if value != nil {
				if objectID, ok := value.(primitive.ObjectID); ok {
					query["userId"] = objectID
				} else if strID, ok := value.(string); ok && strID != "" {
					if objectID, err := primitive.ObjectIDFromHex(strID); err == nil {
						query["userId"] = objectID
					}
				}
			}
		case "username":
			if value != nil && value != "" {
				query["username"] = bson.M{"$regex": value, "$options": "i"}
			}
		case "status":
			if value != nil && value != "" {
				query["status"] = value
			}
		case "ipAddress":
			if value != nil && value != "" {
				query["ipAddress"] = value
			}
		case "startTime":
			if startTime, ok := value.(time.Time); ok {
				if query["loginAt"] == nil {
					query["loginAt"] = bson.M{}
				}
				query["loginAt"].(bson.M)["$gte"] = startTime
			}
		case "endTime":
			if endTime, ok := value.(time.Time); ok {
				if query["loginAt"] == nil {
					query["loginAt"] = bson.M{}
				}
				query["loginAt"].(bson.M)["$lte"] = endTime
			}
		case "country":
			if value != nil && value != "" {
				query["country"] = value
			}
		case "deviceType":
			if value != nil && value != "" {
				query["deviceType"] = value
			}
		case "browser":
			if value != nil && value != "" {
				query["browser"] = value
			}
		}
	}

	return query
}

// buildSortCondition 构建排序条件
func (d *LoginLogDAO) buildSortCondition(filter map[string]interface{}) bson.D {
	// 默认按登录时间降序排序
	sort := bson.D{bson.E{Key: "loginAt", Value: -1}}

	if filter == nil {
		return sort
	}

	if sortBy, exists := filter["sortBy"]; exists && sortBy != "" {
		sortDesc := true // 默认降序
		if desc, ok := filter["sortDesc"]; ok {
			sortDesc = desc.(bool)
		}

		sortOrder := -1
		if !sortDesc {
			sortOrder = 1
		}

		switch sortBy {
		case "loginAt":
			sort = bson.D{bson.E{Key: "loginAt", Value: sortOrder}}
		case "username":
			sort = bson.D{bson.E{Key: "username", Value: sortOrder}}
		case "ipAddress":
			sort = bson.D{bson.E{Key: "ipAddress", Value: sortOrder}}
		case "status":
			sort = bson.D{bson.E{Key: "status", Value: sortOrder}}
		default:
			sort = bson.D{bson.E{Key: "loginAt", Value: -1}}
		}
	}

	return sort
}
