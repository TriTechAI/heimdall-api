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

// PostDAO 文章数据访问层
type PostDAO struct {
	collection *mongo.Collection
}

// NewPostDAO 创建文章DAO实例
func NewPostDAO(database *mongo.Database) *PostDAO {
	return &PostDAO{
		collection: database.Collection("posts"),
	}
}

// Create 创建文章
func (d *PostDAO) Create(ctx context.Context, post *model.Post) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}

	// 验证创建数据
	if err := post.ValidateForCreate(); err != nil {
		return err
	}

	// 准备插入数据
	post.PrepareForInsert()

	_, err := d.collection.InsertOne(ctx, post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("slug already exists")
		}
		return err
	}

	return nil
}

// GetByID 根据ID获取文章
func (d *PostDAO) GetByID(ctx context.Context, id string) (*model.Post, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var post model.Post
	err = d.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

// GetBySlug 根据slug获取文章
func (d *PostDAO) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	if slug == "" {
		return nil, errors.New("slug cannot be empty")
	}

	var post model.Post
	err := d.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

// Update 更新文章信息
func (d *PostDAO) Update(ctx context.Context, id string, updates map[string]interface{}) error {
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
		return errors.New("post not found")
	}

	return nil
}

// Delete 删除文章（软删除）
func (d *PostDAO) Delete(ctx context.Context, id string) error {
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
		return errors.New("post not found")
	}

	return nil
}

// List 获取文章列表
func (d *PostDAO) List(ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
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

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post model.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, 0, err
		}
		posts = append(posts, &post)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPublishedList 获取已发布的文章列表
func (d *PostDAO) GetPublishedList(ctx context.Context, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
	// 强制设置为已发布状态和公开可见
	filter.Status = "published"
	filter.Visibility = "public"

	return d.List(ctx, filter, page, limit)
}

// IncrementViewCount 增加文章浏览量
func (d *PostDAO) IncrementViewCount(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	updates := bson.M{
		"$inc": bson.M{"viewCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := d.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updates)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("post not found")
	}

	return nil
}

// GetByAuthor 根据作者ID获取文章列表
func (d *PostDAO) GetByAuthor(ctx context.Context, authorID string, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
	if authorID == "" {
		return nil, 0, errors.New("authorID cannot be empty")
	}

	// 设置作者过滤条件
	filter.AuthorID = authorID

	return d.List(ctx, filter, page, limit)
}

// GetByTag 根据标签获取文章列表
func (d *PostDAO) GetByTag(ctx context.Context, tagSlug string, filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
	if tagSlug == "" {
		return nil, 0, errors.New("tagSlug cannot be empty")
	}

	// 设置标签过滤条件
	filter.Tag = tagSlug

	return d.List(ctx, filter, page, limit)
}

// GetScheduledPosts 获取应该发布的定时文章
func (d *PostDAO) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
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

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post model.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// Publish 发布文章
func (d *PostDAO) Publish(ctx context.Context, id string) error {
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
		return errors.New("post not found")
	}

	return nil
}

// Unpublish 取消发布文章
func (d *PostDAO) Unpublish(ctx context.Context, id string) error {
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
		return errors.New("post not found")
	}

	return nil
}

// GetPopularPosts 获取热门文章
func (d *PostDAO) GetPopularPosts(ctx context.Context, limit int, days int) ([]*model.Post, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if days < 1 {
		days = 30
	}

	// 查询最近N天的已发布文章，按浏览量排序
	since := time.Now().AddDate(0, 0, -days)
	query := bson.M{
		"status":     "published",
		"visibility": "public",
		"publishedAt": bson.M{
			"$gte": since,
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{
			bson.E{Key: "viewCount", Value: -1},
			bson.E{Key: "publishedAt", Value: -1},
		})

	cursor, err := d.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post model.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetRecentPosts 获取最新文章
func (d *PostDAO) GetRecentPosts(ctx context.Context, limit int) ([]*model.Post, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := bson.M{
		"status":     "published",
		"visibility": "public",
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{bson.E{Key: "publishedAt", Value: -1}})

	cursor, err := d.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []*model.Post
	for cursor.Next(ctx) {
		var post model.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// CreateIndexes 创建文章集合的索引
func (d *PostDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				bson.E{Key: "status", Value: 1},
				bson.E{Key: "visibility", Value: 1},
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
				bson.E{Key: "tags.slug", Value: 1},
				bson.E{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{bson.E{Key: "type", Value: 1}},
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
			Keys: bson.D{bson.E{Key: "viewCount", Value: -1}},
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
				bson.E{Key: "excerpt", Value: "text"},
				bson.E{Key: "markdown", Value: "text"},
			},
		},
	}

	_, err := d.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// buildQuery 构建查询条件
func (d *PostDAO) buildQuery(filter model.PostFilter) bson.M {
	query := bson.M{}

	// 状态过滤
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	// 类型过滤
	if filter.Type != "" {
		query["type"] = filter.Type
	}

	// 可见性过滤
	if filter.Visibility != "" {
		query["visibility"] = filter.Visibility
	}

	// 作者过滤
	if filter.AuthorID != "" {
		authorID, err := primitive.ObjectIDFromHex(filter.AuthorID)
		if err == nil {
			query["authorId"] = authorID
		}
	}

	// 标签过滤
	if filter.Tag != "" {
		query["tags.slug"] = filter.Tag
	}

	// 关键词搜索
	if filter.Keyword != "" {
		query["$or"] = []bson.M{
			{"title": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"excerpt": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"markdown": bson.M{"$regex": filter.Keyword, "$options": "i"}},
		}
	}

	return query
}

// buildSort 构建排序条件
func (d *PostDAO) buildSort(filter model.PostFilter) bson.D {
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
	case "view_count":
		return bson.D{bson.E{Key: "viewCount", Value: sortOrder}}
	case "created_at":
		return bson.D{bson.E{Key: "createdAt", Value: sortOrder}}
	default:
		return bson.D{bson.E{Key: "createdAt", Value: -1}}
	}
}
