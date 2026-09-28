package logic

import (
	"context"
	"database/sql"
	"time"

	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
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

// convertVideos 批量转换，保持返回非 nil slice 以便序列化成 [] 而不是 null。
// 转换后统一用 Redis 里的计数覆盖 like_count，见 fillLikeCounts。
func convertVideos(ctx context.Context, r *redis.Redis, list []*model.Video) []*video.Video {
	out := make([]*video.Video, 0, len(list))
	for _, v := range list {
		out = append(out, convertVideo(v))
	}
	fillLikeCounts(ctx, r, out)
	return out
}

// fillLikeCounts 用 Redis 点赞集合的大小覆盖数据库里的 like_count。
//
// 点赞数的权威来源是 Redis 的点赞集合（interaction 服务维护），video 表的
// like_count 只是定期物化出来的副本。读的时候以 Redis 为准，带来两个好处：
//
//   - 计数实时，不用等物化周期；
//   - 点赞不再写 video 表，视频行的缓存不会被点赞反复击穿——视频流是最热的
//     读路径，之前每次点赞都要让缓存失效。
//
// 读不到就保留数据库里的值：读路径不能因为计数拿不到就整个失败。
func fillLikeCounts(ctx context.Context, r *redis.Redis, items []*video.Video) {
	if r == nil || len(items) == 0 {
		return
	}

	counts := make([]*redis.IntCmd, len(items))
	err := r.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		for i, v := range items {
			key := constant.LikeSetKey(constant.LikeTargetVideo, uint64(v.Id))
			counts[i] = p.SCard(ctx, key)
		}
		return nil
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("批量读取点赞计数失败，回退到数据库值: %v", err)
		return
	}

	for i, cmd := range counts {
		n, err := cmd.Result()
		if err != nil {
			continue // 保留数据库里的值
		}
		items[i].LikeCount = n
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
