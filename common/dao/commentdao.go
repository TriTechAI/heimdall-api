package dao

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
)

// CommentDAO 评论数据访问对象
type CommentDAO struct {
	collection *mongo.Collection
}

// NewCommentDAO 创建评论DAO
func NewCommentDAO(db *mongo.Database) *CommentDAO {
	return &CommentDAO{
		collection: db.Collection("comments"),
	}
}

// ===============================
// 基础CRUD操作
// ===============================

// Create 创建评论
func (dao *CommentDAO) Create(ctx context.Context, comment *model.Comment) error {
	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}

	// 准备插入数据
	comment.PrepareForInsert()

	// 验证数据
	if err := comment.ValidateForCreate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 插入数据
	_, err := dao.collection.InsertOne(ctx, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID 根据ID获取评论
func (dao *CommentDAO) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid comment ID: %w", err)
	}

	var comment model.Comment
	err = dao.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// Update 更新评论
func (dao *CommentDAO) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	// 添加更新时间
	updates["updatedAt"] = time.Now()

	// 构建更新文档
	updateDoc := bson.M{"$set": updates}

	result, err := dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// Delete 删除评论（软删除）
func (dao *CommentDAO) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	// 软删除：更新状态为rejected并添加删除标记
	updates := bson.M{
		"status":    constants.CommentStatusRejected,
		"updatedAt": time.Now(),
		"deletedAt": time.Now(),
	}

	result, err := dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// ===============================
// 查询方法
// ===============================

// List 获取评论列表
func (dao *CommentDAO) List(ctx context.Context, filter model.CommentFilter, page, limit int) ([]*model.Comment, int64, error) {
	// 构建查询条件
	query := dao.buildQuery(filter)

	// 计算总数
	total, err := dao.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	if total == 0 {
		return []*model.Comment{}, 0, nil
	}

	// 构建查询选项
	opts := dao.buildFindOptions(filter, page, limit)

	// 执行查询
	cursor, err := dao.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find comments: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("failed to decode comments: %w", err)
	}

	return comments, total, nil
}

// GetByPostID 获取指定文章的评论
func (dao *CommentDAO) GetByPostID(ctx context.Context, postID string, page, limit int) ([]*model.Comment, int64, error) {
	_, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid post ID: %w", err)
	}

	filter := model.CommentFilter{
		PostID: postID,
		Status: constants.CommentStatusApproved, // 只返回已审核通过的评论
	}

	return dao.List(ctx, filter, page, limit)
}

// GetRepliesByParentID 获取指定评论的回复
func (dao *CommentDAO) GetRepliesByParentID(ctx context.Context, parentID string, page, limit int) ([]*model.Comment, int64, error) {
	objectID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid parent comment ID: %w", err)
	}

	query := bson.M{
		"parentId": objectID,
		"status":   constants.CommentStatusApproved,
		"$or": []bson.M{
			{"deletedAt": bson.M{"$exists": false}},
			{"deletedAt": nil},
		},
	}

	// 计算总数
	total, err := dao.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count replies: %w", err)
	}

	if total == 0 {
		return []*model.Comment{}, 0, nil
	}

	// 构建查询选项
	opts := options.Find().
		SetSort(bson.M{"createdAt": 1}). // 回复按时间正序
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	// 执行查询
	cursor, err := dao.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find replies: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("failed to decode replies: %w", err)
	}

	return comments, total, nil
}

// GetPendingComments 获取待审核评论
func (dao *CommentDAO) GetPendingComments(ctx context.Context, page, limit int) ([]*model.Comment, int64, error) {
	filter := model.CommentFilter{
		Status: constants.CommentStatusPending,
	}

	return dao.List(ctx, filter, page, limit)
}

// GetByAuthorEmail 根据作者邮箱获取评论
func (dao *CommentDAO) GetByAuthorEmail(ctx context.Context, email string, page, limit int) ([]*model.Comment, int64, error) {
	filter := model.CommentFilter{
		AuthorEmail: email,
	}

	return dao.List(ctx, filter, page, limit)
}

// GetByAuthorIP 根据作者IP获取评论
func (dao *CommentDAO) GetByAuthorIP(ctx context.Context, ip string, page, limit int) ([]*model.Comment, int64, error) {
	filter := model.CommentFilter{
		AuthorIP: ip,
	}

	return dao.List(ctx, filter, page, limit)
}

// ===============================
// 统计方法
// ===============================

// GetCommentCount 获取评论总数
func (dao *CommentDAO) GetCommentCount(ctx context.Context, filter model.CommentFilter) (int64, error) {
	query := dao.buildQuery(filter)
	return dao.collection.CountDocuments(ctx, query)
}

// GetCommentCountByPostID 获取指定文章的评论数
func (dao *CommentDAO) GetCommentCountByPostID(ctx context.Context, postID string) (int64, error) {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return 0, fmt.Errorf("invalid post ID: %w", err)
	}

	query := bson.M{
		"postId": objectID,
		"status": constants.CommentStatusApproved,
		"$or": []bson.M{
			{"deletedAt": bson.M{"$exists": false}},
			{"deletedAt": nil},
		},
	}

	return dao.collection.CountDocuments(ctx, query)
}

// GetReplyCountByParentID 获取指定评论的回复数
func (dao *CommentDAO) GetReplyCountByParentID(ctx context.Context, parentID string) (int64, error) {
	objectID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		return 0, fmt.Errorf("invalid parent comment ID: %w", err)
	}

	query := bson.M{
		"parentId": objectID,
		"status":   constants.CommentStatusApproved,
		"$or": []bson.M{
			{"deletedAt": bson.M{"$exists": false}},
			{"deletedAt": nil},
		},
	}

	return dao.collection.CountDocuments(ctx, query)
}

// ===============================
// 业务方法
// ===============================

// ApproveComment 审核通过评论
func (dao *CommentDAO) ApproveComment(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	now := time.Now()
	updates := bson.M{
		"status":     constants.CommentStatusApproved,
		"approvedAt": now,
		"updatedAt":  now,
	}

	result, err := dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to approve comment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// RejectComment 拒绝评论
func (dao *CommentDAO) RejectComment(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	updates := bson.M{
		"status":    constants.CommentStatusRejected,
		"updatedAt": time.Now(),
		"$unset":    bson.M{"approvedAt": ""},
	}

	result, err := dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to reject comment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// MarkAsSpam 标记为垃圾评论
func (dao *CommentDAO) MarkAsSpam(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	updates := bson.M{
		"status":    constants.CommentStatusSpam,
		"updatedAt": time.Now(),
		"$unset":    bson.M{"approvedAt": ""},
	}

	result, err := dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to mark comment as spam: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

// IncrementReplyCount 增加回复数量
func (dao *CommentDAO) IncrementReplyCount(ctx context.Context, parentID string) error {
	objectID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		return fmt.Errorf("invalid parent comment ID: %w", err)
	}

	updates := bson.M{
		"$inc": bson.M{"replyCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err = dao.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updates)
	if err != nil {
		return fmt.Errorf("failed to increment reply count: %w", err)
	}

	return nil
}

// ===============================
// 批量操作
// ===============================

// BatchApprove 批量审核通过
func (dao *CommentDAO) BatchApprove(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no comment IDs provided")
	}

	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid comment ID %s: %w", id, err)
		}
		objectIDs[i] = objectID
	}

	now := time.Now()
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	updates := bson.M{
		"$set": bson.M{
			"status":     constants.CommentStatusApproved,
			"approvedAt": now,
			"updatedAt":  now,
		},
	}

	result, err := dao.collection.UpdateMany(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to batch approve comments: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no comments found")
	}

	return nil
}

// BatchReject 批量拒绝
func (dao *CommentDAO) BatchReject(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no comment IDs provided")
	}

	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid comment ID %s: %w", id, err)
		}
		objectIDs[i] = objectID
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	updates := bson.M{
		"$set": bson.M{
			"status":    constants.CommentStatusRejected,
			"updatedAt": time.Now(),
		},
		"$unset": bson.M{"approvedAt": ""},
	}

	result, err := dao.collection.UpdateMany(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to batch reject comments: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no comments found")
	}

	return nil
}

// BatchDelete 批量删除
func (dao *CommentDAO) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no comment IDs provided")
	}

	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid comment ID %s: %w", id, err)
		}
		objectIDs[i] = objectID
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	updates := bson.M{
		"$set": bson.M{
			"status":    constants.CommentStatusRejected,
			"updatedAt": time.Now(),
			"deletedAt": time.Now(),
		},
	}

	result, err := dao.collection.UpdateMany(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to batch delete comments: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no comments found")
	}

	return nil
}

// BatchMarkSpam 批量标记垃圾评论
func (dao *CommentDAO) BatchMarkSpam(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no comment IDs provided")
	}

	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid comment ID %s: %w", id, err)
		}
		objectIDs[i] = objectID
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	updates := bson.M{
		"$set": bson.M{
			"status":    constants.CommentStatusSpam,
			"updatedAt": time.Now(),
		},
	}

	result, err := dao.collection.UpdateMany(ctx, filter, updates)
	if err != nil {
		return fmt.Errorf("failed to batch mark spam comments: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no comments found")
	}

	return nil
}

// ===============================
// 索引管理
// ===============================

// CreateIndexes 创建索引
func (dao *CommentDAO) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// 文章ID索引
		{
			Keys: bson.D{{Key: "postId", Value: 1}},
		},
		// 父评论ID索引
		{
			Keys: bson.D{{Key: "parentId", Value: 1}},
		},
		// 状态索引
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		// 作者邮箱索引
		{
			Keys: bson.D{{Key: "authorEmail", Value: 1}},
		},
		// 作者IP索引
		{
			Keys: bson.D{{Key: "authorIP", Value: 1}},
		},
		// 创建时间索引
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
		// 复合索引：文章ID + 状态 + 创建时间
		{
			Keys: bson.D{
				{Key: "postId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
		// 复合索引：父评论ID + 状态 + 创建时间
		{
			Keys: bson.D{
				{Key: "parentId", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: 1},
			},
		},
	}

	_, err := dao.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// ===============================
// 私有辅助方法
// ===============================

// buildQuery 构建查询条件
func (dao *CommentDAO) buildQuery(filter model.CommentFilter) bson.M {
	query := bson.M{}

	// 基础查询条件：排除软删除的记录
	query["$or"] = []bson.M{
		{"deletedAt": bson.M{"$exists": false}},
		{"deletedAt": nil},
	}

	// 文章ID过滤
	if filter.PostID != "" {
		if objectID, err := primitive.ObjectIDFromHex(filter.PostID); err == nil {
			query["postId"] = objectID
		}
	}

	// 父评论ID过滤
	if filter.ParentID != "" {
		if objectID, err := primitive.ObjectIDFromHex(filter.ParentID); err == nil {
			query["parentId"] = objectID
		}
	}

	// 作者邮箱过滤
	if filter.AuthorEmail != "" {
		query["authorEmail"] = filter.AuthorEmail
	}

	// 作者IP过滤
	if filter.AuthorIP != "" {
		query["authorIP"] = filter.AuthorIP
	}

	// 状态过滤
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	// 可见性过滤
	if filter.Visibility != "" {
		query["visibility"] = filter.Visibility
	}

	// 类型过滤
	if filter.Type != "" {
		query["type"] = filter.Type
	}

	// 层级过滤
	if filter.Level > 0 {
		query["level"] = filter.Level
	}

	// 关键词搜索
	if filter.Keyword != "" {
		query["$or"] = []bson.M{
			{"content": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"authorName": bson.M{"$regex": filter.Keyword, "$options": "i"}},
		}
	}

	// 时间范围过滤
	if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
		timeQuery := bson.M{}
		if !filter.StartTime.IsZero() {
			timeQuery["$gte"] = filter.StartTime
		}
		if !filter.EndTime.IsZero() {
			timeQuery["$lte"] = filter.EndTime
		}
		query["createdAt"] = timeQuery
	}

	return query
}

// buildFindOptions 构建查询选项
func (dao *CommentDAO) buildFindOptions(filter model.CommentFilter, page, limit int) *options.FindOptions {
	opts := options.Find()

	// 分页
	if page > 0 && limit > 0 {
		opts.SetSkip(int64((page - 1) * limit))
		opts.SetLimit(int64(limit))
	}

	// 排序
	sort := dao.buildSort(filter)
	if len(sort) > 0 {
		opts.SetSort(sort)
	}

	return opts
}

// buildSort 构建排序条件
func (dao *CommentDAO) buildSort(filter model.CommentFilter) bson.M {
	sort := bson.M{}

	if filter.SortBy != "" {
		direction := 1
		if filter.SortDesc {
			direction = -1
		}

		switch filter.SortBy {
		case "createdAt", "updatedAt", "approvedAt":
			sort[filter.SortBy] = direction
		case "replyCount", "likeCount", "level":
			sort[filter.SortBy] = direction
		default:
			// 默认按创建时间倒序
			sort["createdAt"] = -1
		}
	} else {
		// 默认按创建时间倒序
		sort["createdAt"] = -1
	}

	return sort
}