package dao

import (
	"context"
	"errors"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TagDAO 标签数据访问层
type TagDAO struct {
	collection *mongo.Collection
}

// NewTagDAO 创建标签DAO实例
func NewTagDAO(database *mongo.Database) *TagDAO {
	return &TagDAO{
		collection: database.Collection("tags"),
	}
}

// Create 创建标签
func (d *TagDAO) Create(ctx context.Context, tag *model.TagModel) error {
	if tag == nil {
		return errors.New("tag cannot be nil")
	}

	// 验证标签数据
	if err := tag.Validate(); err != nil {
		return err
	}

	// 准备插入数据
	tag.PrepareForCreation()

	_, err := d.collection.InsertOne(ctx, tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("tag slug already exists")
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取标签
func (d *TagDAO) GetByID(ctx context.Context, id string) (*model.TagModel, error) {
	if id == "" {
		return nil, errors.New("tag id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tag id format")
	}

	var tag model.TagModel
	err = d.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&tag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}

	return &tag, nil
}

// GetBySlug 根据Slug获取标签
func (d *TagDAO) GetBySlug(ctx context.Context, slug string) (*model.TagModel, error) {
	if slug == "" {
		return nil, errors.New("tag slug cannot be empty")
	}

	var tag model.TagModel
	err := d.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&tag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}

	return &tag, nil
}

// Update 更新标签
func (d *TagDAO) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("tag id cannot be empty")
	}
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid tag id format")
	}

	// 添加更新时间
	updates["updatedAt"] = time.Now()

	// 验证更新字段
	if err := d.validateUpdateFields(updates); err != nil {
		return err
	}

	result, err := d.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("tag slug already exists")
		}
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("tag not found")
	}

	return nil
}

// Delete 删除标签（软删除）
func (d *TagDAO) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("tag id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid tag id format")
	}

	// 软删除：设置删除时间
	updates := bson.M{
		"deletedAt": time.Now(),
		"updatedAt": time.Now(),
	}

	result, err := d.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID, "deletedAt": bson.M{"$exists": false}},
		bson.M{"$set": updates},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("tag not found or already deleted")
	}

	return nil
}

// List 获取标签列表
func (d *TagDAO) List(ctx context.Context, filter TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > constants.TagsPerPageMax {
		limit = constants.TagsPerPageDefault
	}

	// 构建查询条件
	query := d.buildQuery(filter)

	// 构建排序条件
	sortOptions := d.buildSort(filter.SortBy, filter.SortOrder)

	// 查询选项
	findOptions := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(sortOptions)

	// 执行查询
	cursor, err := d.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var tags []*model.TagModel
	if err := cursor.All(ctx, &tags); err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := d.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// GetPublishedList 获取公开可见的标签列表
func (d *TagDAO) GetPublishedList(ctx context.Context, filter TagFilter, page, limit int) ([]*model.TagModel, int64, error) {
	// 强制设置为公开可见
	filter.Visibility = constants.TagVisibilityPublic
	return d.List(ctx, filter, page, limit)
}

// GetPopularTags 获取热门标签
func (d *TagDAO) GetPopularTags(ctx context.Context, limit int) ([]*model.TagModel, error) {
	if limit <= 0 {
		limit = constants.PopularTagsCount
	}

	// 查询条件：只返回公开可见且有文章的标签
	query := bson.M{
		"visibility":      constants.TagVisibilityPublic,
		"postCount":       bson.M{"$gt": 0},
		"deletedAt":       bson.M{"$exists": false},
	}

	// 按文章数量降序排序
	sortOptions := bson.D{{"postCount", -1}, {"createdAt", -1}}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSort(sortOptions)

	cursor, err := d.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []*model.TagModel
	if err := cursor.All(ctx, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// IncrementPostCount 增加标签的文章数量
func (d *TagDAO) IncrementPostCount(ctx context.Context, tagID string, increment int) error {
	if tagID == "" {
		return errors.New("tag id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(tagID)
	if err != nil {
		return errors.New("invalid tag id format")
	}

	updates := bson.M{
		"$inc": bson.M{"postCount": increment},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := d.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		updates,
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("tag not found")
	}

	return nil
}

// UpdatePostCount 更新标签的文章数量
func (d *TagDAO) UpdatePostCount(ctx context.Context, tagID string, count int) error {
	if tagID == "" {
		return errors.New("tag id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(tagID)
	if err != nil {
		return errors.New("invalid tag id format")
	}

	updates := bson.M{
		"postCount": count,
		"updatedAt": time.Now(),
	}

	result, err := d.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("tag not found")
	}

	return nil
}

// CreateIndexes 创建索引
func (d *TagDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"slug", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"visibility", 1}, {"postCount", -1}},
		},
		{
			Keys: bson.D{{"name", "text"}, {"description", "text"}},
		},
		{
			Keys: bson.D{{"createdAt", -1}},
		},
		{
			Keys: bson.D{{"deletedAt", 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// TagFilter 标签过滤器
type TagFilter struct {
	Name       string // 标签名称过滤
	Visibility string // 可见性过滤
	SortBy     string // 排序字段
	SortOrder  string // 排序方向 (asc/desc)
}

// buildQuery 构建查询条件
func (d *TagDAO) buildQuery(filter TagFilter) bson.M {
	query := bson.M{
		"deletedAt": bson.M{"$exists": false}, // 排除已删除的标签
	}

	// 名称过滤
	if filter.Name != "" {
		query["name"] = bson.M{"$regex": filter.Name, "$options": "i"}
	}

	// 可见性过滤
	if filter.Visibility != "" && constants.IsValidTagVisibility(filter.Visibility) {
		query["visibility"] = filter.Visibility
	}

	return query
}

// buildSort 构建排序条件
func (d *TagDAO) buildSort(sortBy, sortOrder string) bson.D {
	// 默认排序
	defaultSort := bson.D{{"createdAt", -1}}

	// 验证排序字段
	if !constants.IsValidTagSortOrder(sortBy) {
		return defaultSort
	}

	// 确定排序方向
	direction := -1 // 默认降序
	if sortOrder == "asc" {
		direction = 1
	}

	// 构建排序条件
	switch sortBy {
	case constants.TagSortByName:
		return bson.D{{"name", direction}}
	case constants.TagSortBySlug:
		return bson.D{{"slug", direction}}
	case constants.TagSortByPostCount:
		return bson.D{{"postCount", direction}, {"createdAt", -1}}
	case constants.TagSortByCreatedAt:
		return bson.D{{"createdAt", direction}}
	case constants.TagSortByUpdatedAt:
		return bson.D{{"updatedAt", direction}}
	default:
		return defaultSort
	}
}

// validateUpdateFields 验证更新字段
func (d *TagDAO) validateUpdateFields(updates map[string]interface{}) error {
	// 不允许更新的字段
	forbiddenFields := []string{"_id", "createdAt", "deletedAt", "postCount"}
	
	for _, field := range forbiddenFields {
		if _, exists := updates[field]; exists {
			return errors.New("cannot update field: " + field)
		}
	}

	// 验证特定字段
	if name, exists := updates["name"]; exists {
		if nameStr, ok := name.(string); !ok || len(nameStr) == 0 {
			return errors.New("invalid name field")
		}
	}

	if slug, exists := updates["slug"]; exists {
		if slugStr, ok := slug.(string); !ok || len(slugStr) == 0 {
			return errors.New("invalid slug field")
		}
	}

	if visibility, exists := updates["visibility"]; exists {
		if visStr, ok := visibility.(string); !ok || !constants.IsValidTagVisibility(visStr) {
			return errors.New("invalid visibility field")
		}
	}

	return nil
}