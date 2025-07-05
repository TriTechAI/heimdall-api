package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicPostDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据slug获取公开文章详情
func NewGetPublicPostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicPostDetailLogic {
	return &GetPublicPostDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicPostDetailLogic) GetPublicPostDetail(req *types.PublicPostDetailRequest) (resp *types.PublicPostDetailResponse, err error) {
	// 1. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 获取文章
	post, err := l.getPostBySlug(req.Slug)
	if err != nil {
		return nil, fmt.Errorf("获取文章失败: %w", err)
	}

	// 3. 验证文章可见性
	if err := l.validatePostVisibility(post); err != nil {
		return nil, err
	}

	// 4. 获取作者信息
	author, err := l.getAuthorInfo(post.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %w", err)
	}

	// 5. 更新浏览计数（异步，不影响响应）
	l.updateViewCount(post.ID.Hex())

	// 6. 构建响应数据
	postDetail := l.buildPostDetail(post, author)

	// 7. 构建响应
	return l.buildResponse(postDetail), nil
}

// validateRequest 验证请求参数
func (l *GetPublicPostDetailLogic) validateRequest(req *types.PublicPostDetailRequest) error {
	if strings.TrimSpace(req.Slug) == "" {
		return fmt.Errorf("slug不能为空")
	}
	return nil
}

// getPostBySlug 根据slug获取文章
func (l *GetPublicPostDetailLogic) getPostBySlug(slug string) (*model.Post, error) {
	post, err := l.svcCtx.PostDAO.GetBySlug(l.ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %w", err)
	}
	return post, nil
}

// validatePostVisibility 验证文章可见性
func (l *GetPublicPostDetailLogic) validatePostVisibility(post *model.Post) error {
	// 检查文章状态
	if post.Status != constants.PostStatusPublished {
		return fmt.Errorf("文章未发布")
	}

	// 检查文章可见性
	if post.Visibility != constants.PostVisibilityPublic {
		return fmt.Errorf("文章不可见")
	}

	return nil
}

// getAuthorInfo 获取作者信息
func (l *GetPublicPostDetailLogic) getAuthorInfo(authorID string) (*model.User, error) {
	return l.svcCtx.UserDAO.GetByID(l.ctx, authorID)
}

// updateViewCount 更新浏览计数
func (l *GetPublicPostDetailLogic) updateViewCount(postID string) {
	// 同步更新浏览计数，确保测试可以验证
	if err := l.svcCtx.PostDAO.IncrementViewCount(l.ctx, postID); err != nil {
		l.Logger.Errorf("更新浏览计数失败: %v", err)
	}
}

// buildPostDetail 构建文章详情
func (l *GetPublicPostDetailLogic) buildPostDetail(post *model.Post, author *model.User) types.PublicPostDetailData {
	// 构建标签信息
	tags := l.buildTags(post.Tags)

	// 构建作者信息
	authorInfo := l.buildAuthorInfo(author)

	// 格式化时间
	publishedAt := ""
	if post.PublishedAt != nil {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	// 构建canonical URL
	canonicalURL := l.buildCanonicalURL(post.Slug)

	return types.PublicPostDetailData{
		Title:           post.Title,
		Slug:            post.Slug,
		Excerpt:         post.Excerpt,
		HTML:            post.HTML,
		FeaturedImage:   post.FeaturedImage,
		Author:          authorInfo,
		Tags:            tags,
		MetaTitle:       post.MetaTitle,
		MetaDescription: post.MetaDescription,
		CanonicalURL:    canonicalURL,
		ReadingTime:     post.ReadingTime,
		WordCount:       post.WordCount,
		ViewCount:       int64(post.ViewCount),
		PublishedAt:     publishedAt,
		UpdatedAt:       post.UpdatedAt.Format(time.RFC3339),
	}
}

// buildTags 构建标签信息
func (l *GetPublicPostDetailLogic) buildTags(tags []model.Tag) []types.TagInfo {
	tagInfos := make([]types.TagInfo, 0, len(tags))
	for _, tag := range tags {
		tagInfos = append(tagInfos, types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}
	return tagInfos
}

// buildAuthorInfo 构建作者信息
func (l *GetPublicPostDetailLogic) buildAuthorInfo(author *model.User) types.PublicAuthorInfo {
	return types.PublicAuthorInfo{
		Username:     author.Username,
		DisplayName:  author.DisplayName,
		ProfileImage: author.ProfileImage,
		Bio:          author.Bio,
	}
}

// buildCanonicalURL 构建canonical URL
func (l *GetPublicPostDetailLogic) buildCanonicalURL(slug string) string {
	// 这里可以从配置中读取域名，暂时使用相对路径
	return fmt.Sprintf("/posts/%s", slug)
}

// buildResponse 构建响应
func (l *GetPublicPostDetailLogic) buildResponse(postDetail types.PublicPostDetailData) *types.PublicPostDetailResponse {
	return &types.PublicPostDetailResponse{
		Code:      200,
		Message:   "success",
		Data:      postDetail,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
