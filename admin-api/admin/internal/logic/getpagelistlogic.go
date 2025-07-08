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

type GetPageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取页面列表
func NewGetPageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPageListLogic {
	return &GetPageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPageListLogic) GetPageList(req *types.PageListRequest) (resp *types.PageListResponse, err error) {
	// 1. 构建查询过滤条件
	filter := l.buildPageFilter(req)

	// 2. 查询页面列表
	pages, total, err := l.svcCtx.PageDAO.List(l.ctx, filter, req.Page, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("获取页面列表失败: %v", err)
	}

	// 3. 构建页面列表项
	pageItems, err := l.buildPageListItems(pages)
	if err != nil {
		return nil, err
	}

	// 4. 计算分页信息
	pagination := l.calculatePagination(req.Page, req.Limit, int(total))

	// 5. 构建响应
	return &types.PageListResponse{
		Code:    200,
		Message: "获取页面列表成功",
		Data: types.PageListData{
			List:       pageItems,
			Pagination: pagination,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// buildPageFilter 构建查询过滤条件
func (l *GetPageListLogic) buildPageFilter(req *types.PageListRequest) model.PageFilter {
	filter := model.PageFilter{
		Status:   req.Status,
		Template: req.Template,
		AuthorID: req.AuthorID,
		Keyword:  req.Keyword,
		Page:     req.Page,
		Limit:    req.Limit,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	return filter
}

// buildPageListItems 构建页面列表项
func (l *GetPageListLogic) buildPageListItems(pages []*model.Page) ([]types.PageListItem, error) {
	if len(pages) == 0 {
		return []types.PageListItem{}, nil
	}

	// 获取所有唯一的作者ID
	authorIDs := l.extractAuthorIDs(pages)

	// 批量获取作者信息
	authors, err := l.getAuthorsInfo(authorIDs)
	if err != nil {
		return nil, fmt.Errorf("获取作者信息失败: %v", err)
	}

	// 构建页面列表项
	items := make([]types.PageListItem, len(pages))
	for i, page := range pages {
		authorInfo, exists := authors[page.AuthorID.Hex()]
		if !exists {
			return nil, fmt.Errorf("作者信息不存在: %s", page.AuthorID.Hex())
		}

		items[i] = l.buildPageListItem(page, authorInfo)
	}

	return items, nil
}

// extractAuthorIDs 提取所有唯一的作者ID
func (l *GetPageListLogic) extractAuthorIDs(pages []*model.Page) []string {
	authorIDMap := make(map[string]bool)
	for _, page := range pages {
		authorIDMap[page.AuthorID.Hex()] = true
	}

	authorIDs := make([]string, 0, len(authorIDMap))
	for authorID := range authorIDMap {
		authorIDs = append(authorIDs, authorID)
	}

	return authorIDs
}

// getAuthorsInfo 批量获取作者信息
func (l *GetPageListLogic) getAuthorsInfo(authorIDs []string) (map[string]*model.AuthorInfo, error) {
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

// buildPageListItem 构建单个页面列表项
func (l *GetPageListLogic) buildPageListItem(page *model.Page, author *model.AuthorInfo) types.PageListItem {
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
	if page.PublishedAt != nil {
		publishedAt = page.PublishedAt.Format(time.RFC3339)
	}

	return types.PageListItem{
		ID:            page.ID.Hex(),
		Title:         page.Title,
		Slug:          page.Slug,
		Author:        authorInfo,
		Status:        page.Status,
		Template:      page.Template,
		FeaturedImage: page.FeaturedImage,
		PublishedAt:   publishedAt,
		CreatedAt:     page.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     page.UpdatedAt.Format(time.RFC3339),
	}
}

// calculatePagination 计算分页信息
func (l *GetPageListLogic) calculatePagination(page, limit, total int) types.PaginationInfo {
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
