package logic

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/heimdall-api/common/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审核通过评论
func NewApproveCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveCommentLogic {
	return &ApproveCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveCommentLogic) ApproveComment(req *types.CommentApproveRequest) (resp *types.CommentApproveResponse, err error) {
	// 验证评论ID格式
	if err := l.validateCommentID(req.ID); err != nil {
		return nil, err
	}

	// 检查评论是否存在
	comment, err := l.svcCtx.CommentDAO.GetByID(l.ctx, req.ID)
	if err != nil {
		l.Errorf("获取评论失败: %v", err)
		return nil, errors.New("评论不存在")
	}

	// 检查评论状态是否允许审核
	if err := l.validateCommentStatus(comment.Status); err != nil {
		return nil, err
	}

	// 执行审核通过操作
	if err := l.svcCtx.CommentDAO.ApproveComment(l.ctx, req.ID); err != nil {
		l.Errorf("审核评论失败: %v", err)
		return nil, errors.New("审核评论失败")
	}

	l.Infof("评论 %s 审核通过", req.ID)

	return &types.CommentApproveResponse{
		Code:      200,
		Message:   "评论审核通过",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// validateCommentID 验证评论ID
func (l *ApproveCommentLogic) validateCommentID(commentID string) error {
	if commentID == "" {
		return errors.New("评论ID不能为空")
	}

	if !primitive.IsValidObjectID(commentID) {
		return errors.New("无效的评论ID格式")
	}

	return nil
}

// validateCommentStatus 验证评论状态
func (l *ApproveCommentLogic) validateCommentStatus(status string) error {
	if status == constants.CommentStatusApproved {
		return errors.New("评论已经审核通过")
	}

	if status == constants.CommentStatusDeleted {
		return errors.New("已删除的评论不能审核")
	}

	return nil
}
