package logic

import (
	"context"

	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicLogic {
	return &PublicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublicLogic) Public(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
