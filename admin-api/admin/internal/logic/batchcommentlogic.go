package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量操作评论
func NewBatchCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCommentLogic {
	return &BatchCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchCommentLogic) BatchComment(req *types.CommentBatchRequest) (resp *types.CommentBatchResponse, err error) {
	// 验证请求参数
	if err := l.validateBatchRequest(req); err != nil {
		return nil, err
	}

	// 验证评论ID格式
	if err := l.validateCommentIDs(req.IDs); err != nil {
		return nil, err
	}

	// 执行批量操作
	successful, failed, err := l.executeBatchOperation(req)
	if err != nil {
		l.Errorf("批量操作评论失败: %v", err)
		return nil, errors.New("批量操作失败")
	}

	l.Infof("批量操作评论完成: 成功 %d, 失败 %d", successful, failed)

	return &types.CommentBatchResponse{
		Code:       200,
		Message:    l.getBatchResultMessage(req.Action, successful, failed),
		Successful: successful,
		Failed:     failed,
		Timestamp:  time.Now().Format(time.RFC3339),
	}, nil
}

// validateBatchRequest 验证批量请求参数
func (l *BatchCommentLogic) validateBatchRequest(req *types.CommentBatchRequest) error {
	if len(req.IDs) == 0 {
		return errors.New("评论ID列表不能为空")
	}

	if len(req.IDs) > 100 {
		return errors.New("单次批量操作不能超过100条记录")
	}

	if !l.isValidBatchAction(req.Action) {
		return errors.New("无效的批量操作类型")
	}

	return nil
}

// validateCommentIDs 验证评论ID格式
func (l *BatchCommentLogic) validateCommentIDs(ids []string) error {
	for _, id := range ids {
		if id == "" {
			return errors.New("评论ID不能为空")
		}
		if !primitive.IsValidObjectID(id) {
			return fmt.Errorf("无效的评论ID格式: %s", id)
		}
	}
	return nil
}

// isValidBatchAction 验证批量操作类型
func (l *BatchCommentLogic) isValidBatchAction(action string) bool {
	validActions := []string{"approve", "reject", "spam", "delete"}
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// executeBatchOperation 执行批量操作
func (l *BatchCommentLogic) executeBatchOperation(req *types.CommentBatchRequest) (successful, failed int, err error) {
	switch req.Action {
	case "approve":
		return l.batchApproveComments(req.IDs)
	case "reject":
		return l.batchRejectComments(req.IDs)
	case "spam":
		return l.batchMarkSpamComments(req.IDs)
	case "delete":
		return l.batchDeleteComments(req.IDs)
	default:
		return 0, 0, fmt.Errorf("不支持的批量操作: %s", req.Action)
	}
}

// batchApproveComments 批量审核通过评论
func (l *BatchCommentLogic) batchApproveComments(ids []string) (successful, failed int, err error) {
	err = l.svcCtx.CommentDAO.BatchApprove(l.ctx, ids)
	if err != nil {
		return 0, len(ids), err
	}
	return len(ids), 0, nil
}

// batchRejectComments 批量拒绝评论
func (l *BatchCommentLogic) batchRejectComments(ids []string) (successful, failed int, err error) {
	err = l.svcCtx.CommentDAO.BatchReject(l.ctx, ids)
	if err != nil {
		return 0, len(ids), err
	}
	return len(ids), 0, nil
}

// batchMarkSpamComments 批量标记垃圾评论
func (l *BatchCommentLogic) batchMarkSpamComments(ids []string) (successful, failed int, err error) {
	err = l.svcCtx.CommentDAO.BatchMarkSpam(l.ctx, ids)
	if err != nil {
		return 0, len(ids), err
	}
	return len(ids), 0, nil
}

// batchDeleteComments 批量删除评论
func (l *BatchCommentLogic) batchDeleteComments(ids []string) (successful, failed int, err error) {
	successCount := 0
	failCount := 0

	for _, id := range ids {
		if err := l.svcCtx.CommentDAO.Delete(l.ctx, id); err != nil {
			l.Errorf("删除评论 %s 失败: %v", id, err)
			failCount++
		} else {
			successCount++
		}
	}

	return successCount, failCount, nil
}

// getBatchResultMessage 获取批量操作结果消息
func (l *BatchCommentLogic) getBatchResultMessage(action string, successful, failed int) string {
	var actionText string
	switch action {
	case "approve":
		actionText = "审核通过"
	case "reject":
		actionText = "拒绝"
	case "spam":
		actionText = "标记垃圾"
	case "delete":
		actionText = "删除"
	default:
		actionText = "操作"
	}

	if failed == 0 {
		return fmt.Sprintf("成功%s %d 条评论", actionText, successful)
	} else {
		return fmt.Sprintf("%s评论完成: 成功 %d 条, 失败 %d 条", actionText, successful, failed)
	}
}
