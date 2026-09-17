package logic

import (
	"database/sql"
	"time"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"
)

// timeLayout 接口约定的时间格式，与响应里输出的格式一致
const timeLayout = "2006-01-02 15:04:05"

// toStatus 将业务错误转换为 proto 的 Status 响应
func toStatus(e *errno.Errno) *interaction.Status {
	return &interaction.Status{Code: e.Code, Msg: e.Msg}
}

// convertComment 将 model.Comment 转换为 proto 的 Comment；
// 时间字段按接口文档约定输出字符串，空值（NULL）输出空串
func convertComment(c *model.Comment) *interaction.Comment {
	return &interaction.Comment{
		Id:         int64(c.Id),
		UserId:     int64(c.UserId),
		VideoId:    int64(c.VideoId),
		ParentId:   int64(c.ParentId),
		LikeCount:  int64(c.LikeCount),
		ChildCount: int64(c.ChildCount),
		Content:    c.Content,
		CreatedAt:  formatTime(c.CreatedAt),
		UpdatedAt:  formatTime(c.UpdatedAt),
		DeletedAt:  formatNullTime(c.DeletedAt),
	}
}

// convertComments 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null
func convertComments(list []*model.Comment) []*interaction.Comment {
	out := make([]*interaction.Comment, 0, len(list))
	for _, c := range list {
		out = append(out, convertComment(c))
	}
	return out
}

// convertVideo 把 video 服务返回的视频转成本服务的响应类型。
// 两个 proto 包各有一份 Video，字段一一对应，用 Get 访问器安全取值。
func convertVideo(v *video.Video) *interaction.Video {
	return &interaction.Video{
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

func formatTime(t time.Time) string {
	return t.Local().Format(timeLayout)
}

func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return formatTime(t.Time)
}

// resolveLikeTarget 校验并解析点赞/取消点赞的目标。
// 视频与评论必须指定且只能指定一个，两者共用一张点赞表，靠类型区分。
func resolveLikeTarget(videoId, commentId int64) (uint64, uint64, *errno.Errno) {
	switch {
	case videoId > 0 && commentId > 0:
		return 0, 0, errno.ErrInvalidParam.WithMsg("视频 id 和评论 id 只能指定一个")
	case videoId > 0:
		return constant.LikeTargetVideo, uint64(videoId), nil
	case commentId > 0:
		return constant.LikeTargetComment, uint64(commentId), nil
	default:
		return 0, 0, errno.ErrInvalidParam.WithMsg("视频 id 和评论 id 至少指定一个")
	}
}
