package middleware

import (
	"net/http"

	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 请求头名称，与接口文档约定一致
const (
	HeaderAccessToken  = "Access-Token"
	HeaderRefreshToken = "Refresh-Token"
)

type AuthMiddleware struct {
	signer *jwt.Signer
}

func NewAuthMiddleware(signer *jwt.Signer) *AuthMiddleware {
	return &AuthMiddleware{signer: signer}
}

// Handle 校验 Access-Token，通过后将用户身份写入请求 context 再放行
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.signer.ParseAccess(r.Header.Get(HeaderAccessToken))
		if err != nil {
			writeErrno(w, errno.ErrUnauthorized)
			return
		}

		// 用户身份注入 context，供后续 handler/logic 读取
		ctx := jwt.ContextWithClaims(r.Context(), claims)
		next(w, r.WithContext(ctx))
	}
}

// writeErrno 以统一 JSON 结构输出业务错误
func writeErrno(w http.ResponseWriter, e *errno.Errno) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	httpx.WriteJson(w, http.StatusOK, types.Status{Code: e.Code, Msg: e.Msg})
}
