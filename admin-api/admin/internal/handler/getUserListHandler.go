package handler

import (
	"net/http"

	"github.com/heimdall-api/admin-api/admin/internal/logic"
	"github.com/heimdall-api/admin-api/admin/internal/svc"
	"github.com/heimdall-api/admin-api/admin/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取用户列表
func GetUserListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetUserListLogic(r.Context(), svcCtx)
		resp, err := l.GetUserList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
