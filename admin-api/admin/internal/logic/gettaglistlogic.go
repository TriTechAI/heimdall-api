package logic

import (
	"context"
	"math"
	"time"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/dao"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagListLogic {
	return &GetTagListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagListLogic) GetTagList(req *types.TagListRequest) (resp *types.TagListResponse, err error) {
	// 输入验证
	if err := l.validateListRequest(req); err != nil {
		l.Errorf("标签列表请求验证失败: %v", err)
		return nil, err
	}

	// 设置默认值
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 || limit > constants.TagsPerPageMax {
		limit = constants.TagsPerPageDefault
	}

	// 构建过滤器
	filter := dao.TagFilter{
		Name:       req.Name,
		Visibility: req.Visibility,
		SortBy:     req.SortBy,
		SortOrder:  "",
	}

	if req.SortDesc {
		filter.SortOrder = "desc"
	} else {
		filter.SortOrder = "asc"
	}

	// 调用DAO获取数据
	tags, total, err := l.svcCtx.TagDAO.List(l.ctx, filter, page, limit)
	if err != nil {
		l.Errorf("获取标签列表失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	tagList := make([]types.TagDetailInfo, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, types.TagDetailInfo{
			ID:              tag.ID.Hex(),
			Name:            tag.Name,
			Slug:            tag.Slug,
			Description:     tag.Description,
			Color:           tag.Color,
			FeaturedImage:   tag.FeaturedImage,
			MetaTitle:       tag.MetaTitle,
			MetaDescription: tag.MetaDescription,
			PostCount:       tag.PostCount,
			Visibility:      tag.Visibility,
			CreatedAt:       tag.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       tag.UpdatedAt.Format(time.RFC3339),
		})
	}

	// 计算分页信息
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	// 构建响应
	resp = &types.TagListResponse{
		Code:      200,
		Message:   "获取标签列表成功",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: types.TagListData{
			List: tagList,
			Pagination: types.PaginationInfo{
				Page:       page,
				Limit:      limit,
				Total:      int(total),
				TotalPages: totalPages,
				HasNext:    hasNext,
				HasPrev:    hasPrev,
			},
		},
	}

	l.Infof("获取标签列表成功: 页码=%d, 每页=%d, 总数=%d", page, limit, total)
	return resp, nil
}

// validateListRequest 验证列表请求
func (l *GetTagListLogic) validateListRequest(req *types.TagListRequest) error {
	if req == nil {
		return nil // 空请求使用默认值
	}

	// 验证可见性参数
	if req.Visibility != "" && !constants.IsValidTagVisibility(req.Visibility) {
		return &ValidationError{
			Field:   "visibility",
			Message: "无效的可见性参数",
			Valid:   constants.GetAllTagVisibilities(),
		}
	}

	// 验证排序字段
	if req.SortBy != "" && !constants.IsValidTagSortOrder(req.SortBy) {
		return &ValidationError{
			Field:   "sortBy",
			Message: "无效的排序字段",
			Valid:   constants.GetAllTagSortOrders(),
		}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Valid   interface{} `json:"valid,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}
