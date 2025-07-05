package logic

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPostListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文章列表
func NewGetPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostListLogic {
	return &GetPostListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPostListLogic) GetPostList(req *types.PostListRequest) (resp *types.PostListResponse, err error) {
	// 1. 构建查询过滤条件
	filter := l.buildPostFilter(req)

	// 2. 查询文章列表
	posts, total, err := l.svcCtx.PostDAO.List(l.ctx, filter, req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("获取文章列表失败: %v", err)
	}

	// 3. 构建文章列表项
	postItems, err := l.buildPostListItems(posts)
	if err != nil {
		return nil, err
	}

	// 4. 计算分页信息
	pagination := l.calculatePagination(req.Page, req.Limit, int(total))

	// 5. 构建响应
	return &types.PostListResponse{
		Code:    200,
		Message: "获取文章列表成功",
		Data: types.PostListData{
			List:       postItems,
			Pagination: pagination,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPostFilter 构建查询过滤条件
func (l *GetPostListLogic) buildPostFilter(req *types.PostListRequest) model.PostFilter {
	filter := model.PostFilter{
		Status:     req.Status,
		Type:       req.Type,
		Visibility: req.Visibility,
		AuthorID:   req.AuthorID,
		Tag:        req.Tag,
		Keyword:    req.Keyword,
		SortBy:     req.SortBy,
		SortDesc:   req.SortDesc,
	}

	return filter
}

// buildPostListItems 构建文章列表项
func (l *GetPostListLogic) buildPostListItems(posts []*model.Post) ([]types.PostListItem, error) {
	if len(posts) == 0 {
		return []types.PostListItem{}, nil
	}

	// 获取所有唯一的作者ID
	authorIDs := l.extractAuthorIDs(posts)

	// 批量获取作者信息
	authors, err := l.getAuthorsInfo(authorIDs)
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %v", err)
	}

	// 构建文章列表项
	items := make([]types.PostListItem, len(posts))
	for i, post := range posts {
		authorInfo, exists := authors[post.AuthorID.Hex()]
		if !exists {
			return nil, fmt.Errorf("作者信息不存在: %s", post.AuthorID.Hex())
		}

		items[i] = l.buildPostListItem(post, authorInfo)
	}

	return items, nil
}

// extractAuthorIDs 提取所有唯一的作者ID
func (l *GetPostListLogic) extractAuthorIDs(posts []*model.Post) []string {
	authorIDMap := make(map[string]bool)
	for _, post := range posts {
		authorIDMap[post.AuthorID.Hex()] = true
	}

	authorIDs := make([]string, 0, len(authorIDMap))
	for authorID := range authorIDMap {
		authorIDs = append(authorIDs, authorID)
	}

	return authorIDs
}

// getAuthorsInfo 批量获取作者信息
func (l *GetPostListLogic) getAuthorsInfo(authorIDs []string) (map[string]*model.AuthorInfo, error) {
	authors := make(map[string]*model.AuthorInfo)

	for _, authorID := range authorIDs {
		user, err := l.svcCtx.UserDAO.GetByID(l.ctx, authorID)
		if err != nil {
			return nil, err
		}
		authors[authorID] = user.ToAuthorInfo()
	}

	return authors, nil
}

// buildPostListItem 构建单个文章列表项
func (l *GetPostListLogic) buildPostListItem(post *model.Post, author *model.AuthorInfo) types.PostListItem {
	// 转换标签
	tags := make([]types.TagInfo, len(post.Tags))
	for i, tag := range post.Tags {
		tags[i] = types.TagInfo{
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}

	// 转换作者信息
	authorInfo := types.AuthorInfo{
		ID:           author.ID,
		Username:     author.Username,
		DisplayName:  author.DisplayName,
		ProfileImage: author.ProfileImage,
		Bio:          author.Bio,
	}

	// 格式化发布时间
	var publishedAt string
	if post.PublishedAt != nil {
		publishedAt = post.PublishedAt.Format(time.RFC3339)
	}

	return types.PostListItem{
		ID:            post.ID.Hex(),
		Title:         post.Title,
		Slug:          post.Slug,
		Excerpt:       post.Excerpt,
		FeaturedImage: post.FeaturedImage,
		Type:          post.Type,
		Status:        post.Status,
		Visibility:    post.Visibility,
		Author:        authorInfo,
		Tags:          tags,
		ReadingTime:   post.ReadingTime,
		ViewCount:     post.ViewCount,
		PublishedAt:   publishedAt,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
	}
}

// calculatePagination 计算分页信息
func (l *GetPostListLogic) calculatePagination(page, limit, total int) types.PaginationInfo {
	var totalPages int
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	} else {
		totalPages = 0
	}

	return types.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
