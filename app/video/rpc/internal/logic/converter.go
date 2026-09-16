package logic

import (
	"database/sql"
	"time"

	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
)

// timeLayout 接口约定的时间格式。
// 响应里的时间按它输出，请求里的时间游标（如 lastest_time）也按它解析，两边必须一致。
const timeLayout = "2006-01-02 15:04:05"

// toStatus 将业务错误转换为 proto 的 Status 响应
func toStatus(e *errno.Errno) *video.Status {
	return &video.Status{Code: e.Code, Msg: e.Msg}
}

// convertVideo 将 model.Video 转换为 proto 的 Video；
// 时间字段按接口文档约定输出字符串，空值（NULL）输出空串
func convertVideo(v *model.Video) *video.Video {
	return &video.Video{
		Id:           int64(v.Id),
		UserId:       int64(v.UserId),
		VideoUrl:     v.VideoUrl,
		CoverUrl:     v.CoverUrl,
		Title:        v.Title,
		Description:  v.Description,
		VisitCount:   int64(v.VisitCount),
		LikeCount:    int64(v.LikeCount),
		CommentCount: int64(v.CommentCount),
		CreatedAt:    formatTime(v.CreatedAt),
		UpdatedAt:    formatTime(v.UpdatedAt),
		DeleteAt:     formatNullTime(v.DeleteAt),
	}
}

// convertVideos 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null
func convertVideos(list []*model.Video) []*video.Video {
	out := make([]*video.Video, 0, len(list))
	for _, v := range list {
		out = append(out, convertVideo(v))
	}
	return out
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
