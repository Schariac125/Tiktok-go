package video

import (
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
)

// toStatus 把业务错误转成响应里的 Status
func toStatus(e *errno.Errno) types.Status {
	return types.Status{Code: e.Code, Msg: e.Msg}
}

// convertVideo 把 rpc 返回的 Video 转成网关的响应类型。
// 用 protobuf 的 Get 访问器，rpcResp.Data 为 nil 时也能安全取值。
func convertVideo(v *video.Video) types.Video {
	return types.Video{
		Id:           v.GetId(),
		UserId:       v.GetUserId(),
		VideoUrl:     v.GetVideoUrl(),
		CoverUrl:     v.GetCoverUrl(),
		Title:        v.GetTitle(),
		Description:  v.GetDescription(),
		VisitCount:   v.GetVisitCount(),
		LikeCount:    v.GetLikeCount(),
		CommentCount: v.GetCommentCount(),
		CreatedAt:    v.GetCreatedAt(),
		UpdatedAt:    v.GetUpdatedAt(),
		DeleteAt:     v.GetDeleteAt(),
	}
}

// convertVideos 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null
func convertVideos(list []*video.Video) []types.Video {
	out := make([]types.Video, 0, len(list))
	for _, v := range list {
		out = append(out, convertVideo(v))
	}
	return out
}
