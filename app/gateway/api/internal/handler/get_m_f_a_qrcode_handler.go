// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tiktok-go/app/gateway/api/internal/logic"
	"tiktok-go/app/gateway/api/internal/svc"
)

func getMFAQrcodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetMFAQrcodeLogic(r.Context(), svcCtx)
		resp, err := l.GetMFAQrcode()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
