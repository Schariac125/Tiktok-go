package social

import (
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
)

// toStatus 把业务错误转成响应里的 Status
func toStatus(e *errno.Errno) types.Status {
	return types.Status{Code: e.Code, Msg: e.Msg}
}

// convertUserBrief 把 rpc 返回的用户简要信息转成网关的响应类型
func convertUserBrief(u *social.UserBrief) types.UserBrief {
	return types.UserBrief{
		Id:       u.GetId(),
		Username: u.GetUsername(),
	}
}

// convertUserBriefs 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null
func convertUserBriefs(list []*social.UserBrief) []types.UserBrief {
	out := make([]types.UserBrief, 0, len(list))
	for _, u := range list {
		out = append(out, convertUserBrief(u))
	}
	return out
}
