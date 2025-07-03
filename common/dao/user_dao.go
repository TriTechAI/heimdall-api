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

// UserDAO 用户数据访问层
type UserDAO struct {
	collection *mongo.Collection
}

// NewUserDAO 创建用户DAO实例
func NewUserDAO(database *mongo.Database) *UserDAO {
	return &UserDAO{
		collection: database.Collection("users"),
	}
}

// Create 创建用户
func (d *UserDAO) Create(ctx context.Context, user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// 准备插入数据
	user.PrepareForInsert()

	_, err := d.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username or email already exists")
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取用户
func (d *UserDAO) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var user model.User
	err = d.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (d *UserDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	var user model.User
	err := d.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (d *UserDAO) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	var user model.User
	err := d.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// Update 更新用户信息
func (d *UserDAO) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if updates == nil || len(updates) == 0 {
		return errors.New("updates cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	// 添加更新时间
	updates["updatedAt"] = time.Now()

	// 构建更新文档
	updateDoc := bson.M{"$set": updates}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username or email already exists")
		}
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete 删除用户（软删除）
func (d *UserDAO) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	// 软删除：更新状态为inactive
	updates := map[string]interface{}{
		"status":    "inactive",
		"updatedAt": time.Now(),
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// List 获取用户列表
func (d *UserDAO) List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]*model.User, int64, error) {
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
	query := bson.M{}
	if filter != nil {
		for key, value := range filter {
			switch key {
			case "role":
				if value != nil && value != "" {
					query["role"] = value
				}
			case "status":
				if value != nil && value != "" {
					query["status"] = value
				}
			case "keyword":
				if value != nil && value != "" {
					keyword := value.(string)
					query["$or"] = []bson.M{
						{"username": bson.M{"$regex": keyword, "$options": "i"}},
						{"email": bson.M{"$regex": keyword, "$options": "i"}},
						{"displayName": bson.M{"$regex": keyword, "$options": "i"}},
					}
				}
			}
		}
	}

	// 获取总数
	total, err := d.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 计算跳过的文档数
	skip := (page - 1) * limit

	// 构建排序条件
	sort := bson.D{bson.E{Key: "createdAt", Value: -1}} // 默认按创建时间降序
	if filter != nil {
		if sortBy, exists := filter["sortBy"]; exists && sortBy != "" {
			sortDesc := false
			if desc, ok := filter["sortDesc"]; ok {
				sortDesc = desc.(bool)
			}

			sortOrder := 1
			if sortDesc {
				sortOrder = -1
			}

			switch sortBy {
			case "username":
				sort = bson.D{bson.E{Key: "username", Value: sortOrder}}
			case "createdAt":
				sort = bson.D{bson.E{Key: "createdAt", Value: sortOrder}}
			case "lastLoginAt":
				sort = bson.D{bson.E{Key: "lastLoginAt", Value: sortOrder}}
			default:
				sort = bson.D{bson.E{Key: "createdAt", Value: -1}}
			}
		}
	}

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

	var users []*model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLoginInfo 更新登录信息
func (d *UserDAO) UpdateLoginInfo(ctx context.Context, id string, ipAddress string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"lastLoginAt":    now,
		"lastLoginIP":    ipAddress,
		"loginFailCount": 0, // 重置失败次数
		"updatedAt":      now,
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// IncrementLoginFailCount 增加登录失败次数
func (d *UserDAO) IncrementLoginFailCount(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updates := bson.M{
		"$inc": bson.M{"loginFailCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updates)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// LockUser 锁定用户
func (d *UserDAO) LockUser(ctx context.Context, id string, lockedUntil time.Time) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updates := map[string]interface{}{
		"status":      "locked",
		"lockedUntil": lockedUntil,
		"updatedAt":   time.Now(),
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UnlockUser 解锁用户
func (d *UserDAO) UnlockUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updates := map[string]interface{}{
		"status":         "active",
		"loginFailCount": 0,
		"updatedAt":      time.Now(),
	}

	// 移除锁定时间字段
	unsetFields := bson.M{
		"$set":   updates,
		"$unset": bson.M{"lockedUntil": ""},
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, unsetFields)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetLockedUsers 获取被锁定的用户列表
func (d *UserDAO) GetLockedUsers(ctx context.Context) ([]*model.User, error) {
	query := bson.M{
		"status": "locked",
		"lockedUntil": bson.M{
			"$lte": time.Now(),
		},
	}

	cursor, err := d.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CreateIndexes 创建用户集合的索引
func (d *UserDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{bson.E{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				bson.E{Key: "role", Value: 1},
				bson.E{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{bson.E{Key: "lockedUntil", Value: 1}},
		},
		{
			Keys: bson.D{bson.E{Key: "createdAt", Value: -1}},
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
