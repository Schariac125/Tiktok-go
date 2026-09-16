package logic

import (
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
)

// toStatus 将业务错误转换为 proto 的 Status 响应
func toStatus(e *errno.Errno) *video.Status {
	return &video.Status{Code: e.Code, Msg: e.Msg}
}
