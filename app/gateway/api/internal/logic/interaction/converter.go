package interaction

import (
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/pkg/errno"
)

// toStatus 把业务错误转成响应里的 Status
func toStatus(e *errno.Errno) types.Status {
	return types.Status{Code: e.Code, Msg: e.Msg}
}

// convertComment 把 rpc 返回的评论转成网关的响应类型。
// 用 protobuf 的 Get 访问器，字段为 nil 时也能安全取值。
func convertComment(c *interaction.Comment) types.Comment {
	return types.Comment{
		Id:         c.GetId(),
		UserId:     c.GetUserId(),
		VideoId:    c.GetVideoId(),
		ParentId:   c.GetParentId(),
		LikeCount:  c.GetLikeCount(),
		ChildCount: c.GetChildCount(),
		Content:    c.GetContent(),
		CreatedAt:  c.GetCreatedAt(),
		UpdatedAt:  c.GetUpdatedAt(),
		DeletedAt:  c.GetDeletedAt(),
	}
}

// convertComments 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null
func convertComments(list []*interaction.Comment) []types.Comment {
	out := make([]types.Comment, 0, len(list))
	for _, c := range list {
		out = append(out, convertComment(c))
	}
	return out
}

// convertVideo 把 rpc 返回的视频转成网关的响应类型
func convertVideo(v *interaction.Video) types.Video {
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

// convertVideos 批量转换，保持返回非 nil slice
func convertVideos(list []*interaction.Video) []types.Video {
	out := make([]types.Video, 0, len(list))
	for _, v := range list {
		out = append(out, convertVideo(v))
	}
	return out
}
