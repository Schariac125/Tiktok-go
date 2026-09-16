// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"mime/multipart"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tiktok-go/app/gateway/api/internal/logic/video"
	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/pkg/constant"
)

// 视频投稿 POST /video/publish
// 以 multipart/form-data 上传：字段 data 为视频原始文件（必填）、
// cover 为封面文件（可省略），title、description 为普通表单字段
func PublishVideoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(constant.MultipartMaxMemory); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 视频源文件（必填）
		dataFile, dataHeader, err := r.FormFile("data")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer dataFile.Close()

		// 封面文件（可选）
		var coverFile multipart.File
		var coverHeader *multipart.FileHeader
		if f, h, err := r.FormFile("cover"); err == nil {
			coverFile, coverHeader = f, h
			defer coverFile.Close()
		}

		l := video.NewPublishVideoLogic(r.Context(), svcCtx)
		resp, err := l.PublishVideo(dataFile, dataHeader, coverFile, coverHeader,
			r.FormValue("title"), r.FormValue("description"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
