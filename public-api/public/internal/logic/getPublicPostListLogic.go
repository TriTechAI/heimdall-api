package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicPostListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取公开文章列表
func NewGetPublicPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicPostListLogic {
	return &GetPublicPostListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicPostListLogic) GetPublicPostList(req *types.PublicPostListRequest) (resp *types.PublicPostListResponse, err error) {
	// 1. 验证请求参数
	if err := l.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. 构建查询过滤器
	filter, err := l.buildPostFilter(req)
	if err != nil {
		return nil, err
	}

	// 3. 查询文章列表
	posts, total, err := l.queryPosts(filter, req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("获取文章列表失败: %w", err)
	}

	// 4. 构建响应数据
	postItems, err := l.buildPostItems(posts)
	if err != nil {
		return nil, fmt.Errorf("构建文章列表失败: %w", err)
	}

	// 5. 构建分页信息
	pagination := l.buildPagination(req.Page, req.Limit, int(total))

	// 6. 构建响应
	return l.buildResponse(postItems, pagination), nil
}

// validateRequest 验证请求参数
func (l *GetPublicPostListLogic) validateRequest(req *types.PublicPostListRequest) error {
	if req.Page < 1 {
		return fmt.Errorf("页码必须大于0")
	}
	if req.Limit < 1 || req.Limit > 20 {
		return fmt.Errorf("每页记录数必须在1-20之间")
	}
	return nil
}

// buildPostFilter 构建文章查询过滤器
func (l *GetPublicPostListLogic) buildPostFilter(req *types.PublicPostListRequest) (model.PostFilter, error) {
	filter := model.PostFilter{
		Status:     constants.PostStatusPublished,
		Visibility: constants.PostVisibilityPublic,
		Keyword:    req.Keyword,
		Tag:        req.Tag,
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
	}

	// 如果指定了作者，需要先获取作者ID
	if req.Author != "" {
		authorID, err := l.getAuthorIDByUsername(req.Author)
		if err != nil {
			return filter, fmt.Errorf("作者不存在: %w", err)
		}
		filter.AuthorID = authorID
	}

	return filter, nil
}

// getAuthorIDByUsername 根据用户名获取作者ID
func (l *GetPublicPostListLogic) getAuthorIDByUsername(username string) (string, error) {
	user, err := l.svcCtx.UserDAO.GetByUsername(l.ctx, username)
	if err != nil {
		return "", err
	}
	return user.ID.Hex(), nil
}

// queryPosts 查询文章列表
func (l *GetPublicPostListLogic) queryPosts(filter model.PostFilter, page, limit int) ([]*model.Post, int64, error) {
	return l.svcCtx.PostDAO.GetPublishedList(l.ctx, filter, page, limit)
}

// buildPostItems 构建文章列表项
func (l *GetPublicPostListLogic) buildPostItems(posts []*model.Post) ([]types.PublicPostListItem, error) {
	if len(posts) == 0 {
		return []types.PublicPostListItem{}, nil
	}

	items := make([]types.PublicPostListItem, 0, len(posts))

	for _, post := range posts {
		// 获取作者信息
		author, err := l.getAuthorInfo(post.AuthorID.Hex())
		if err != nil {
			return nil, fmt.Errorf("获取作者信息失败: %w", err)
		}

		// 构建文章项
		item := l.buildPostItem(post, author)
		items = append(items, item)
	}

	return items, nil
}

// getAuthorInfo 获取作者信息
func (l *GetPublicPostListLogic) getAuthorInfo(authorID string) (*model.User, error) {
	return l.svcCtx.UserDAO.GetByID(l.ctx, authorID)
}

// buildPostItem 构建单个文章项
func (l *GetPublicPostListLogic) buildPostItem(post *model.Post, author *model.User) types.PublicPostListItem {
	// 构建标签信息
	tags := make([]types.TagInfo, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tags = append(tags, types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	// 构建作者信息
	authorInfo := types.PublicAuthorInfo{
		Username:     author.Username,
		DisplayName:  author.DisplayName,
		ProfileImage: author.ProfileImage,
		Bio:          author.Bio,
	}

	// 格式化时间
	publishedAt := ""
	if post.PublishedAt != nil {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	return types.PublicPostListItem{
		Title:         post.Title,
		Slug:          post.Slug,
		Excerpt:       post.Excerpt,
		FeaturedImage: post.FeaturedImage,
		Author:        authorInfo,
		Tags:          tags,
		ReadingTime:   post.ReadingTime,
		ViewCount:     int64(post.ViewCount),
		PublishedAt:   publishedAt,
		UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
	}
}

// buildPagination 构建分页信息
func (l *GetPublicPostListLogic) buildPagination(page, limit int, total int) types.PaginationInfo {
	totalPages := (total + limit - 1) / limit

	return types.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// buildResponse 构建响应
func (l *GetPublicPostListLogic) buildResponse(items []types.PublicPostListItem, pagination types.PaginationInfo) *types.PublicPostListResponse {
	return &types.PublicPostListResponse{
		Code:    200,
		Message: "success",
		Data: types.PublicPostListData{
			List:       items,
			Pagination: pagination,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
