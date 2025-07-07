package logic

import (
	"context"

	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取评论详情
func NewGetCommentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentDetailLogic {
	return &GetCommentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommentDetailLogic) GetCommentDetail(req *types.CommentDetailRequest) (resp *types.CommentDetailResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
