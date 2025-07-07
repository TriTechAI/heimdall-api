package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"
	"github.com/heimdall-api/common/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取评论列表
func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommentListLogic) GetCommentList(req *types.CommentListRequest) (resp *types.CommentListResponse, err error) {
	// 构建过滤条件
	filter, err := l.buildCommentFilter(req)
	if err != nil {
		l.Errorf("构建评论过滤条件失败: %v", err)
		return nil, errors.New("过滤条件无效")
	}

	// 验证分页参数
	if err := l.validatePaginationParams(req); err != nil {
		return nil, err
	}

	// 查询评论列表
	comments, total, err := l.svcCtx.CommentDAO.List(l.ctx, *filter, req.Page, req.Limit)
	if err != nil {
		l.Errorf("查询评论列表失败: %v", err)
		return nil, errors.New("查询评论列表失败")
	}

	// 转换为响应格式
	commentInfos := l.convertToCommentInfos(comments)

	// 构建分页信息
	pagination := l.buildPaginationInfo(req.Page, req.Limit, int(total))

	return &types.CommentListResponse{
		Code:       200,
		Message:    "获取评论列表成功",
		Data:       commentInfos,
		Pagination: pagination,
		Timestamp:  time.Now().Format(time.RFC3339),
	}, nil
}

// buildCommentFilter 构建评论过滤条件
func (l *GetCommentListLogic) buildCommentFilter(req *types.CommentListRequest) (*model.CommentFilter, error) {
	filter := &model.CommentFilter{
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	// 文章ID过滤
	if req.PostID != "" {
		if !primitive.IsValidObjectID(req.PostID) {
			return nil, fmt.Errorf("无效的文章ID格式")
		}
		filter.PostID = req.PostID
	}

	// 父评论ID过滤
	if req.ParentID != "" {
		if !primitive.IsValidObjectID(req.ParentID) {
			return nil, fmt.Errorf("无效的父评论ID格式")
		}
		filter.ParentID = req.ParentID
	}

	// 状态过滤
	if req.Status != "" {
		if !l.isValidStatus(req.Status) {
			return nil, fmt.Errorf("无效的评论状态")
		}
		filter.Status = req.Status
	}

	// 作者邮箱过滤
	if req.AuthorEmail != "" {
		filter.AuthorEmail = req.AuthorEmail
	}

	// 作者IP过滤
	if req.AuthorIP != "" {
		filter.AuthorIP = req.AuthorIP
	}

	// 关键词搜索
	if req.Keyword != "" {
		filter.Keyword = req.Keyword
	}

	// 时间范围过滤
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, fmt.Errorf("无效的开始时间格式")
		}
		filter.StartTime = startTime
	}

	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("无效的结束时间格式")
		}
		filter.EndTime = endTime
	}

	return filter, nil
}

// validatePaginationParams 验证分页参数
func (l *GetCommentListLogic) validatePaginationParams(req *types.CommentListRequest) error {
	if req.Page < 1 {
		return errors.New("页码不能小于1")
	}
	if req.Limit < 1 || req.Limit > 100 {
		return errors.New("每页记录数必须在1-100之间")
	}
	return nil
}

// isValidStatus 验证评论状态
func (l *GetCommentListLogic) isValidStatus(status string) bool {
	validStatuses := []string{
		constants.CommentStatusPending,
		constants.CommentStatusApproved,
		constants.CommentStatusRejected,
		constants.CommentStatusSpam,
	}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// convertToCommentInfos 转换评论为响应格式
func (l *GetCommentListLogic) convertToCommentInfos(comments []*model.Comment) []types.CommentInfo {
	commentInfos := make([]types.CommentInfo, 0, len(comments))
	for _, comment := range comments {
		commentInfo := types.CommentInfo{
			ID:            comment.ID.Hex(),
			PostID:        comment.PostID.Hex(),
			Content:       comment.Content,
			AuthorName:    comment.AuthorName,
			AuthorEmail:   comment.AuthorEmail,
			AuthorWebsite: comment.AuthorWebsite,
			AuthorIP:      comment.AuthorIP,
			UserAgent:     comment.UserAgent,
			Status:        comment.Status,
			Visibility:    comment.Visibility,
			Type:          comment.Type,
			Level:         comment.Level,
			ReplyCount:    comment.ReplyCount,
			LikeCount:     comment.LikeCount,
			CreatedAt:     comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     comment.UpdatedAt.Format(time.RFC3339),
		}

		// 父评论ID（可选）
		if !comment.ParentID.IsZero() {
			commentInfo.ParentID = comment.ParentID.Hex()
		}

		// 审核时间（可选）
		if !comment.ApprovedAt.IsZero() {
			commentInfo.ApprovedAt = comment.ApprovedAt.Format(time.RFC3339)
		}

		commentInfos = append(commentInfos, commentInfo)
	}
	return commentInfos
}

// buildPaginationInfo 构建分页信息
func (l *GetCommentListLogic) buildPaginationInfo(page, limit, total int) types.PaginationInfo {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
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
