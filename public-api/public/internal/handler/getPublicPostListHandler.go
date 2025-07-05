package handler

import (
	"net/http"

	"github.com/heimdall-api/public-api/public/internal/logic"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取公开文章列表
func GetPublicPostListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicPostListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPublicPostListLogic(r.Context(), svcCtx)
		resp, err := l.GetPublicPostList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
