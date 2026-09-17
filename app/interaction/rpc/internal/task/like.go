// Package task 组装互动模块自己的队列消费任务。
//
// 队列的批量取用与周期调度由 pkg/task 提供，这里只负责点赞特有的部分：
// 队列项的格式、幂等落库、以及按目标类型更新计数。
package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/app/video/rpc/videoservice"
	"tiktok-go/pkg/constant"
	gopkgtask "tiktok-go/pkg/task"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 队列里的动作码
const (
	actionLike   = 1 // 点赞
	actionUnlike = 0 // 取消点赞
)

// likeHandler 处理点赞队列里的单条变更
type likeHandler struct {
	likes    model.UserLikeModel
	comments model.CommentModel
	videos   videoservice.VideoService
}

// NewLikeWorker 组装点赞队列的消费任务。
// 点赞是高频操作，接口只写队列就返回，由这个任务按周期刷进数据库并更新计数。
func NewLikeWorker(r *redis.Redis, likes model.UserLikeModel, comments model.CommentModel, videos videoservice.VideoService) *gopkgtask.Worker {
	h := &likeHandler{likes: likes, comments: comments, videos: videos}

	return gopkgtask.NewWorker(r, gopkgtask.Options{
		Name:     "点赞",
		Key:      constant.LikeQueueKey,
		Batch:    constant.LikeSyncBatch,
		Interval: constant.LikeSyncInterval,
		Handler:  h.handle,
	})
}

// PushLike 把一次点赞写进队列
func PushLike(ctx context.Context, w *gopkgtask.Worker, userId, targetType, targetId uint64) error {
	return w.Push(ctx, encodeItem(actionLike, userId, targetType, targetId))
}

// PushUnlike 把一次取消点赞写进队列
func PushUnlike(ctx context.Context, w *gopkgtask.Worker, userId, targetType, targetId uint64) error {
	return w.Push(ctx, encodeItem(actionUnlike, userId, targetType, targetId))
}

// handle 落库一条变更，并按幂等结果调整计数。
// 可能被重复调用（整批重试），因此依赖 InsertIgnore / DeleteOne 的影响行数来判断是否真的变更过。
func (h *likeHandler) handle(ctx context.Context, item string) error {
	action, userId, targetType, targetId, err := decodeItem(item)
	if err != nil {
		// 脏数据直接跳过，别卡住整条队列
		logx.Errorf("跳过无法解析的点赞队列项 %q: %v", item, err)
		return nil
	}

	var affected int64
	if action == actionLike {
		affected, err = h.likes.InsertIgnore(ctx, userId, targetType, targetId)
	} else {
		affected, err = h.likes.DeleteOne(ctx, userId, targetType, targetId)
	}
	if err != nil {
		return err
	}
	// 没有真正增删（重复点赞、重复取消）就不动计数
	if affected == 0 {
		return nil
	}

	delta := int64(1)
	if action == actionUnlike {
		delta = -1
	}
	return h.applyCountDelta(ctx, targetType, targetId, delta)
}

// applyCountDelta 调整目标上的点赞数。
// video 表归 video 服务管，这里走 rpc 更新；comment 表是本服务的，直接改。
func (h *likeHandler) applyCountDelta(ctx context.Context, targetType, targetId uint64, delta int64) error {
	switch targetType {
	case constant.LikeTargetVideo:
		resp, err := h.videos.IncrLikeCount(ctx, &video.IncrLikeCountReq{
			VideoId: int64(targetId),
			Delta:   delta,
		})
		if err != nil {
			return err
		}
		if resp.GetBase().GetCode() != errno.OK.Code {
			return fmt.Errorf("更新视频点赞数被拒: %s", resp.GetBase().GetMsg())
		}
		return nil
	case constant.LikeTargetComment:
		return h.comments.IncrLikeCount(ctx, targetId, delta)
	default:
		return fmt.Errorf("未知的点赞目标类型 %d", targetType)
	}
}

// encodeItem 把一条变更编码成队列项 "动作:用户:目标类型:目标id"
func encodeItem(action int, userId, targetType, targetId uint64) string {
	return fmt.Sprintf("%d:%d:%d:%d", action, userId, targetType, targetId)
}

// decodeItem 解析队列项
func decodeItem(item string) (action int, userId, targetType, targetId uint64, err error) {
	parts := strings.Split(item, ":")
	if len(parts) != 4 {
		return 0, 0, 0, 0, errors.New("字段数应为 4")
	}

	if action, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if userId, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
		return
	}
	if targetType, err = strconv.ParseUint(parts[2], 10, 64); err != nil {
		return
	}
	targetId, err = strconv.ParseUint(parts[3], 10, 64)
	return
}
