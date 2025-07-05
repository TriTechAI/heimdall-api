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

// PageDAO 页面数据访问层
type PageDAO struct {
	collection *mongo.Collection
}

// NewPageDAO 创建页面DAO实例
func NewPageDAO(database *mongo.Database) *PageDAO {
	return &PageDAO{
		collection: database.Collection("pages"),
	}
}

// Create 创建页面
func (d *PageDAO) Create(ctx context.Context, page *model.Page) error {
	if page == nil {
		return errors.New("page cannot be nil")
	}

	// 验证创建数据
	if err := page.ValidateForCreate(); err != nil {
		return err
	}

	// 准备插入数据
	page.PrepareForInsert()

	_, err := d.collection.InsertOne(ctx, page)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("slug already exists")
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取页面
func (d *PageDAO) GetByID(ctx context.Context, id string) (*model.Page, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var page model.Page
	err = d.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&page)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &page, nil
}

// GetBySlug 根据slug获取页面
func (d *PageDAO) GetBySlug(ctx context.Context, slug string) (*model.Page, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	var page model.Page
	err := d.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&page)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &page, nil
}

// Update 更新页面信息
func (d *PageDAO) Update(ctx context.Context, id string, updates map[string]interface{}) error {
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
			return errors.New("slug already exists")
		}
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("page not found")
	}

	return nil
}

// Delete 删除页面（软删除）
func (d *PageDAO) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	// 软删除：更新状态为archived
	updates := map[string]interface{}{
		"status":    "archived",
		"updatedAt": time.Now(),
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("page not found")
	}

	return nil
}

// List 获取页面列表
func (d *PageDAO) List(ctx context.Context, filter model.PageFilter, page, limit int) ([]*model.Page, int64, error) {
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
	query := d.buildQuery(filter)

	// 获取总数
	total, err := d.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 计算跳过的文档数
	skip := (page - 1) * limit

	// 构建排序条件
	sort := d.buildSort(filter)

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

	var pages []*model.Page
	for cursor.Next(ctx) {
		var page model.Page
		if err := cursor.Decode(&page); err != nil {
			return nil, 0, err
		}
		pages = append(pages, &page)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return pages, total, nil
}

// GetPublishedList 获取已发布的页面列表
func (d *PageDAO) GetPublishedList(ctx context.Context, filter model.PageFilter, page, limit int) ([]*model.Page, int64, error) {
	// 强制设置为已发布状态
	filter.Status = "published"

	return d.List(ctx, filter, page, limit)
}

// GetByAuthor 根据作者ID获取页面列表
func (d *PageDAO) GetByAuthor(ctx context.Context, authorID string, filter model.PageFilter, page, limit int) ([]*model.Page, int64, error) {
	if authorID == "" {
		return nil, 0, errors.New("authorID cannot be empty")
	}

	// 设置作者过滤条件
	filter.AuthorID = authorID

	return d.List(ctx, filter, page, limit)
}

// GetByTemplate 根据模板获取页面列表
func (d *PageDAO) GetByTemplate(ctx context.Context, template string, filter model.PageFilter, page, limit int) ([]*model.Page, int64, error) {
	if template == "" {
		return nil, 0, errors.New("template cannot be empty")
	}

	// 设置模板过滤条件
	filter.Template = template

	return d.List(ctx, filter, page, limit)
}

// GetScheduledPages 获取应该发布的定时页面
func (d *PageDAO) GetScheduledPages(ctx context.Context) ([]*model.Page, error) {
	query := bson.M{
		"status": "scheduled",
		"publishedAt": bson.M{
			"$lte": time.Now(),
		},
	}

	cursor, err := d.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pages []*model.Page
	for cursor.Next(ctx) {
		var page model.Page
		if err := cursor.Decode(&page); err != nil {
			return nil, err
		}
		pages = append(pages, &page)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// Publish 发布页面
func (d *PageDAO) Publish(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      "published",
		"publishedAt": now,
		"updatedAt":   now,
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("page not found")
	}

	return nil
}

// Unpublish 取消发布页面
func (d *PageDAO) Unpublish(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updates := map[string]interface{}{
		"status":    "draft",
		"updatedAt": time.Now(),
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("page not found")
	}

	return nil
}

// GetRecentPages 获取最新页面
func (d *PageDAO) GetRecentPages(ctx context.Context, limit int) ([]*model.Page, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := bson.M{
		"status": "published",
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{bson.E{Key: "publishedAt", Value: -1}})

	cursor, err := d.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pages []*model.Page
	for cursor.Next(ctx) {
		var page model.Page
		if err := cursor.Decode(&page); err != nil {
			return nil, err
		}
		pages = append(pages, &page)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// CreateIndexes 创建页面集合的索引
func (d *PageDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				bson.E{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				bson.E{Key: "authorId", Value: 1},
				bson.E{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				bson.E{Key: "template", Value: 1},
				bson.E{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{bson.E{Key: "publishedAt", Value: -1}},
		},
		{
			Keys: bson.D{bson.E{Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{bson.E{Key: "updatedAt", Value: -1}},
		},
		{
			Keys: bson.D{
				bson.E{Key: "status", Value: 1},
				bson.E{Key: "publishedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				bson.E{Key: "title", Value: "text"},
				bson.E{Key: "content", Value: "text"},
			},
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// buildQuery 构建查询条件
func (d *PageDAO) buildQuery(filter model.PageFilter) bson.M {
	query := bson.M{}

	// 状态过滤
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	// 模板过滤
	if filter.Template != "" {
		query["template"] = filter.Template
	}

	// 作者过滤
	if filter.AuthorID != "" {
		authorID, err := primitive.ObjectIDFromHex(filter.AuthorID)
		if err == nil {
			query["authorId"] = authorID
		}
	}

	// 关键词搜索
	if filter.Keyword != "" {
		query["$or"] = []bson.M{
			{"title": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"content": bson.M{"$regex": filter.Keyword, "$options": "i"}},
		}
	}

	return query
}

// buildSort 构建排序条件
func (d *PageDAO) buildSort(filter model.PageFilter) bson.D {
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := 1
	if filter.SortDesc {
		sortOrder = -1
	}

	switch sortBy {
	case "title":
		return bson.D{bson.E{Key: "title", Value: sortOrder}}
	case "updated_at":
		return bson.D{bson.E{Key: "updatedAt", Value: sortOrder}}
	case "published_at":
		return bson.D{bson.E{Key: "publishedAt", Value: sortOrder}}
	case "created_at":
		return bson.D{bson.E{Key: "createdAt", Value: sortOrder}}
	default:
		return bson.D{bson.E{Key: "createdAt", Value: -1}}
	}
}
