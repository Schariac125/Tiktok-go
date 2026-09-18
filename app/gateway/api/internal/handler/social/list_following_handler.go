// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package social

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tiktok-go/app/gateway/api/internal/logic/social"
	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
)

func ListFollowingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFollowingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := social.NewListFollowingLogic(r.Context(), svcCtx)
		resp, err := l.ListFollowing(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
