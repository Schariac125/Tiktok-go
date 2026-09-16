// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tiktok-go/app/gateway/api/internal/logic/user"
	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/pkg/constant"
)

// 上传头像，头像应该要采用云端存储 PUT /user/avatar/upload
// 头像源文件以 multipart/form-data 上传，字段名 avatar
func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 multipart 表单，超出阈值的部分落到临时文件；文件大小校验在 rpc 侧
		if err := r.ParseMultipartForm(constant.MultipartMaxMemory); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := user.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
