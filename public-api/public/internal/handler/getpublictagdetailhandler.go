package handler

import (
	"net/http"

	"github.com/heimdall-api/public-api/public/internal/logic"
	"github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/public-api/public/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPublicTagDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicTagDetailRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPublicTagDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetPublicTagDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
