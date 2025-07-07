package logic

import (
	"context"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SpamCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记垃圾评论
func NewSpamCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SpamCommentLogic {
	return &SpamCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SpamCommentLogic) SpamComment(req *types.CommentSpamRequest) (resp *types.CommentSpamResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
