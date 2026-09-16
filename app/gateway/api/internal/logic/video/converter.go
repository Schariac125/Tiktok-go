package video

import (
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/pkg/errno"
)

// toStatus 把业务错误转成响应里的 Status
func toStatus(e *errno.Errno) types.Status {
	return types.Status{Code: e.Code, Msg: e.Msg}
}
