// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tiktok-go/app/gateway/api/internal/logic/video"
	"tiktok-go/app/gateway/api/internal/svc"
)

func PublishVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := video.NewPublishVideoLogic(r.Context(), svcCtx)
		resp, err := l.PublishVideo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
