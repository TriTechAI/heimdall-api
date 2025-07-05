package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublishPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发布文章
func NewPublishPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishPostLogic {
	return &PublishPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublishPostLogic) PublishPost(req *types.PostPublishRequest) (resp *types.PostPublishResponse, err error) {
	// 1. 验证文章ID
	if err := l.validatePostID(req.ID); err != nil {
		return nil, err
	}

	// 2. 获取当前用户ID
	userID, err := l.getCurrentUserID()
	if err != nil {
		return nil, err
	}

	// 3. 获取文章信息
	post, err := l.svcCtx.PostDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("获取文章信息失败: %w", err)
	}

	// 4. 检查权限
	if err := l.checkPermission(post, userID); err != nil {
		return nil, err
	}

	// 5. 验证发布状态
	if err := l.validatePublishStatus(post); err != nil {
		return nil, err
	}

	// 6. 解析发布时间
	publishedAt, err := l.parsePublishedAt(req.PublishedAt)
	if err != nil {
		return nil, err
	}

	// 7. 执行发布操作
	if err := l.executePublish(req.ID, publishedAt); err != nil {
		return nil, err
	}

	// 8. 构建响应
	return l.buildPublishResponse(req.ID)
}

// validatePostID 验证文章ID格式
func (l *PublishPostLogic) validatePostID(id string) error {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return fmt.Errorf("无效的文章ID格式")
	}
	return nil
}

// getCurrentUserID 获取当前用户ID
func (l *PublishPostLogic) getCurrentUserID() (string, error) {
	userID, ok := l.ctx.Value("uid").(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("用户认证失败")
	}
	return userID, nil
}

// checkPermission 检查用户权限
func (l *PublishPostLogic) checkPermission(post *model.Post, userID string) error {
	// 验证用户是否存在
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否为文章作者
	if post.AuthorID.Hex() != userID {
		return fmt.Errorf("无权限发布此文章")
	}

	return nil
}

// validatePublishStatus 验证发布状态
func (l *PublishPostLogic) validatePublishStatus(post *model.Post) error {
	if post.Status == constants.PostStatusPublished {
		return fmt.Errorf("文章已经发布")
	}
	return nil
}

// parsePublishedAt 解析发布时间
func (l *PublishPostLogic) parsePublishedAt(publishedAtStr string) (*time.Time, error) {
	if publishedAtStr == "" {
		// 如果没有指定时间，使用当前时间
		now := time.Now()
		return &now, nil
	}

	// 解析RFC3339格式的时间
	publishedAt, err := time.Parse(time.RFC3339, publishedAtStr)
	if err != nil {
		return nil, fmt.Errorf("无效的发布时间格式: %w", err)
	}

	return &publishedAt, nil
}

// executePublish 执行发布操作
func (l *PublishPostLogic) executePublish(postID string, publishedAt *time.Time) error {
	// 1. 更新发布时间
	updates := map[string]interface{}{
		"publishedAt": publishedAt,
	}
	if err := l.svcCtx.PostDAO.Update(l.ctx, postID, updates); err != nil {
		return fmt.Errorf("更新发布时间失败: %w", err)
	}

	// 2. 执行发布操作
	if err := l.svcCtx.PostDAO.Publish(l.ctx, postID); err != nil {
		return fmt.Errorf("发布文章失败: %w", err)
	}

	return nil
}

// buildPublishResponse 构建发布响应
func (l *PublishPostLogic) buildPublishResponse(postID string) (*types.PostPublishResponse, error) {
	// 获取发布后的文章信息
	post, err := l.svcCtx.PostDAO.GetByID(l.ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("获取发布后文章信息失败: %w", err)
	}

	// 获取作者信息
	author, err := l.svcCtx.UserDAO.GetByID(l.ctx, post.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}

	// 构建文章详情数据
	data := l.buildPostDetailData(post, author)

	return &types.PostPublishResponse{
		Code:      200,
		Message:   "文章发布成功",
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPostDetailData 构建文章详情数据
func (l *PublishPostLogic) buildPostDetailData(post *model.Post, author *model.User) types.PostDetailData {
	// 转换标签
	tags := make([]types.TagInfo, len(post.Tags))
	for i, tag := range post.Tags {
		tags[i] = types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}

	// 构建作者信息
	var authorInfo types.AuthorInfo
	if author != nil {
		authorInfo = types.AuthorInfo{
			ID:           author.ID.Hex(),
			Username:     author.Username,
			DisplayName:  author.DisplayName,
			ProfileImage: author.ProfileImage,
			Bio:          author.Bio,
		}
	}

	// 格式化发布时间
	var publishedAt string
	if post.PublishedAt != nil {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	return types.PostDetailData{
		ID:              post.ID.Hex(),
		Title:           post.Title,
		Slug:            post.Slug,
		Excerpt:         post.Excerpt,
		Markdown:        post.Markdown,
		HTML:            post.HTML,
		FeaturedImage:   post.FeaturedImage,
		Type:            post.Type,
		Status:          post.Status,
		Visibility:      post.Visibility,
		Author:          authorInfo,
		Tags:            tags,
		MetaTitle:       post.MetaTitle,
		MetaDescription: post.MetaDescription,
		CanonicalURL:    post.CanonicalURL,
		ReadingTime:     post.ReadingTime,
		WordCount:       post.WordCount,
		ViewCount:       post.ViewCount,
		PublishedAt:     publishedAt,
		CreatedAt:       post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       post.UpdatedAt.Format(time.RFC3339),
	}
}
