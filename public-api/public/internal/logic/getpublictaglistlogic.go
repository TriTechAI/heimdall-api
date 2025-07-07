package logic

import (
	"context"
	"errors"
	"time"

	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"
	"github.com/heimdall-api/common/model"
	"github.com/heimdall-api/common/utils"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublicTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublicTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublicTagListLogic {
	return &GetPublicTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPublicTagListLogic) GetPublicTagList(req *types.PublicTagListRequest) (resp *types.PublicTagListResponse, err error) {
	// 验证请求参数
	if err = l.validateRequest(req); err != nil {
		return l.errorResponse(utils.StatusBadRequest, err.Error()), nil
	}

	// 尝试从缓存获取数据
	var cachedResult types.PublicTagListResponse
	err = l.svcCtx.ContentCacheManager.GetTagList(l.ctx, req.Page, req.Limit, req.SortBy, req.SortDesc, &cachedResult)
	if err == nil {
		return &cachedResult, nil
	}

	// 缓存未命中，从数据库查询
	// 构建查询过滤器
	filter := l.buildTagFilter(req)

	// 查询公开标签列表
	tags, total, err := l.svcCtx.TagDAO.GetPublishedList(l.ctx, filter, req.Page, req.Limit)
	if err != nil {
		l.Errorf("Failed to get public tag list: %v", err)
		return l.errorResponse(utils.StatusInternalServerError, "Failed to get tag list"), nil
	}

	// 构建响应数据
	tagInfos := l.buildTagInfos(tags)
	pagination := l.buildPagination(req.Page, req.Limit, total)

	// 构建响应
	response := &types.PublicTagListResponse{
		Code:      utils.StatusOK,
		Message:   "Success",
		Data:      types.PublicTagListData{List: tagInfos, Pagination: pagination},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 设置缓存（异步执行，失败不影响响应）
	go func() {
		if err := l.svcCtx.ContentCacheManager.SetTagList(context.Background(), req.Page, req.Limit, req.SortBy, req.SortDesc, response); err != nil {
			l.Errorf("Failed to set tag list cache: %v", err)
		}
	}()

	return response, nil
}

// validateRequest 验证请求参数
func (l *GetPublicTagListLogic) validateRequest(req *types.PublicTagListRequest) error {
	// Page和Limit的验证由go-zero框架处理
	
	// 验证排序字段
	if !l.isValidSortBy(req.SortBy) {
		return errors.New("invalid sort field")
	}

	return nil
}

// buildTagFilter 构建标签过滤器
func (l *GetPublicTagListLogic) buildTagFilter(req *types.PublicTagListRequest) dao.TagFilter {
	// 确定排序方向
	sortOrder := "desc"
	if !req.SortDesc {
		sortOrder = "asc"
	}

	return dao.TagFilter{
		Visibility: constants.TagVisibilityPublic,
		SortBy:     req.SortBy,
		SortOrder:  sortOrder,
	}
}

// buildTagInfos 构建标签信息列表
func (l *GetPublicTagListLogic) buildTagInfos(tags []*model.TagModel) []types.PublicTagInfo {
	if len(tags) == 0 {
		return []types.PublicTagInfo{}
	}

	tagInfos := make([]types.PublicTagInfo, 0, len(tags))
	for _, tag := range tags {
		tagInfo := types.PublicTagInfo{
			ID:          tag.ID.Hex(),
			Name:        tag.Name,
			Slug:        tag.Slug,
			Description: tag.Description,
			Color:       tag.Color,
			PostCount:   tag.PostCount,
			CreatedAt:   tag.CreatedAt.Format(time.RFC3339),
		}
		tagInfos = append(tagInfos, tagInfo)
	}

	return tagInfos
}

// buildPagination 构建分页信息
func (l *GetPublicTagListLogic) buildPagination(page, limit int, total int64) types.PaginationInfo {
	totalPages := int(total+int64(limit)-1) / limit
	
	return types.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// isValidSortBy 验证排序字段有效性
func (l *GetPublicTagListLogic) isValidSortBy(sortBy string) bool {
	validSortFields := []string{"name", "postCount", "createdAt"}
	for _, field := range validSortFields {
		if sortBy == field {
			return true
		}
	}
	return false
}

// errorResponse 构建错误响应
func (l *GetPublicTagListLogic) errorResponse(code int, message string) *types.PublicTagListResponse {
	return &types.PublicTagListResponse{
		Code:      code,
		Message:   message,
		Data:      types.PublicTagListData{List: []types.PublicTagInfo{}, Pagination: types.PaginationInfo{}},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
