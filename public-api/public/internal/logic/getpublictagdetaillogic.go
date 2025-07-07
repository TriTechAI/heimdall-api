package logic

import (
	"context"
	"errors"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicTagDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicTagDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicTagDetailLogic {
	return &GetPublicTagDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicTagDetailLogic) GetPublicTagDetail(req *types.PublicTagDetailRequest) (resp *types.PublicTagDetailResponse, err error) {
	// 验证请求参数
	if err = l.validateRequest(req); err != nil {
		return l.errorResponse(utils.StatusBadRequest, err.Error()), nil
	}

	// 尝试从缓存获取数据
	var cachedResult types.PublicTagDetailResponse
	err = l.svcCtx.ContentCacheManager.GetTagDetail(l.ctx, req.Slug, &cachedResult)
	if err == nil {
		return &cachedResult, nil
	}

	// 缓存未命中，从数据库查询
	// 根据slug获取标签详情
	tag, err := l.svcCtx.TagDAO.GetBySlug(l.ctx, req.Slug)
	if err != nil {
		l.Errorf("Failed to get tag by slug %s: %v", req.Slug, err)
		return l.errorResponse(utils.StatusNotFound, "Tag not found"), nil
	}

	// 检查标签是否为公开可见
	if !tag.IsPublic() {
		return l.errorResponse(utils.StatusNotFound, "Tag not found"), nil
	}

	// 获取该标签下的最新文章
	recentPosts, err := l.getRecentPostsByTag(tag.Slug)
	if err != nil {
		l.Errorf("Failed to get recent posts for tag %s: %v", tag.Slug, err)
		// 即使获取文章失败，也不影响标签详情的返回
		recentPosts = []types.PublicPostListItem{}
	}

	// 构建响应数据
	tagDetail := l.buildTagDetail(tag, recentPosts)

	// 构建响应
	response := &types.PublicTagDetailResponse{
		Code:      utils.StatusOK,
		Message:   "Success",
		Data:      tagDetail,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 设置缓存（异步执行，失败不影响响应）
	go func() {
		if err := l.svcCtx.ContentCacheManager.SetTagDetail(context.Background(), req.Slug, response); err != nil {
			l.Errorf("Failed to set tag detail cache: %v", err)
		}
	}()

	return response, nil
}

// validateRequest 验证请求参数
func (l *GetPublicTagDetailLogic) validateRequest(req *types.PublicTagDetailRequest) error {
	if req.Slug == "" {
		return errors.New("tag slug cannot be empty")
	}
	return nil
}

// getRecentPostsByTag 获取标签下的最新文章
func (l *GetPublicTagDetailLogic) getRecentPostsByTag(tagSlug string) ([]types.PublicPostListItem, error) {
	// 构建文章查询过滤器
	postFilter := model.PostFilter{
		Tag:      tagSlug,
		Status:   constants.PostStatusPublished,
		SortBy:   constants.PostSortByPublishedAt,
		SortDesc: true,
	}

	// 查询最新的5篇文章
	posts, _, err := l.svcCtx.PostDAO.GetPublishedList(l.ctx, postFilter, 1, 5)
	if err != nil {
		return nil, err
	}

	return l.buildPostListItems(posts), nil
}

// buildTagDetail 构建标签详情数据
func (l *GetPublicTagDetailLogic) buildTagDetail(tag *model.TagModel, recentPosts []types.PublicPostListItem) types.PublicTagDetailData {
	return types.PublicTagDetailData{
		ID:          tag.ID.Hex(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		PostCount:   tag.PostCount,
		CreatedAt:   tag.CreatedAt.Format(time.RFC3339),
		RecentPosts: recentPosts,
	}
}

// buildPostListItems 构建文章列表项
func (l *GetPublicTagDetailLogic) buildPostListItems(posts []*model.Post) []types.PublicPostListItem {
	if len(posts) == 0 {
		return []types.PublicPostListItem{}
	}

	items := make([]types.PublicPostListItem, 0, len(posts))
	for _, post := range posts {
		// 获取作者信息
		author := l.getAuthorInfo(post.AuthorID.Hex())

		// 构建标签信息
		tags := l.buildTagInfosFromPost(post.Tags)

		item := types.PublicPostListItem{
			Title:         post.Title,
			Slug:          post.Slug,
			Excerpt:       post.Excerpt,
			FeaturedImage: post.FeaturedImage,
			Author:        author,
			Tags:          tags,
			ReadingTime:   post.ReadingTime,
			ViewCount:     post.ViewCount,
			PublishedAt:   post.PublishedAt.Format(time.RFC3339),
			UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
		}
		items = append(items, item)
	}

	return items
}

// getAuthorInfo 获取作者信息
func (l *GetPublicTagDetailLogic) getAuthorInfo(authorID string) types.PublicAuthorInfo {
	// 获取作者详情
	user, err := l.svcCtx.UserDAO.GetByID(l.ctx, authorID)
	if err != nil {
		l.Errorf("Failed to get author info for ID %s: %v", authorID, err)
		return types.PublicAuthorInfo{
			Username:    "unknown",
			DisplayName: "Unknown Author",
		}
	}

	return types.PublicAuthorInfo{
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		ProfileImage: user.ProfileImage,
		Bio:          user.Bio,
	}
}

// buildTagInfosFromPost 从文章Tags构建标签信息
func (l *GetPublicTagDetailLogic) buildTagInfosFromPost(postTags []model.Tag) []types.TagInfo {
	if len(postTags) == 0 {
		return []types.TagInfo{}
	}

	tags := make([]types.TagInfo, 0, len(postTags))
	for _, tag := range postTags {
		tags = append(tags, types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	return tags
}

// buildTagInfos 构建标签信息 (保留原方法以兼容)
func (l *GetPublicTagDetailLogic) buildTagInfos(tagNames, tagSlugs []string) []types.TagInfo {
	if len(tagNames) == 0 || len(tagSlugs) == 0 {
		return []types.TagInfo{}
	}

	minLen := len(tagNames)
	if len(tagSlugs) < minLen {
		minLen = len(tagSlugs)
	}

	tags := make([]types.TagInfo, 0, minLen)
	for i := 0; i < minLen; i++ {
		tags = append(tags, types.TagInfo{
			Name: tagNames[i],
			Slug: tagSlugs[i],
		})
	}

	return tags
}

// errorResponse 构建错误响应
func (l *GetPublicTagDetailLogic) errorResponse(code int, message string) *types.PublicTagDetailResponse {
	return &types.PublicTagDetailResponse{
		Code:      code,
		Message:   message,
		Data:      types.PublicTagDetailData{},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
