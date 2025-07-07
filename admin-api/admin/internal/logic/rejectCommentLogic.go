package logic

import (
	"context"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拒绝评论
func NewRejectCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectCommentLogic {
	return &RejectCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RejectCommentLogic) RejectComment(req *types.CommentRejectRequest) (resp *types.CommentRejectResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
