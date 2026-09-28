package task

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/app/video/rpc/videoservice"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Materializer struct {
	redis    *redis.Redis
	comments model.CommentModel
	videos   videoservice.VideoService
	interval time.Duration
}

func NewMaterializer(r *redis.Redis, comments model.CommentModel, videos videoservice.VideoService) *Materializer {
	return &Materializer{
		redis:    r,
		comments: comments,
		videos:   videos,
		interval: constant.LikeMaterializeInterval,
	}
}

// Start 启动后台物化，随 ctx 取消而停止
func (m *Materializer) Start(ctx context.Context) {
	if m == nil || m.redis == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		logx.Infof("点赞计数物化任务已启动，周期 %s", m.interval)
		for {
			select {
			case <-ctx.Done():
				logx.Infof("点赞计数物化任务退出")
				return
			case <-ticker.C:
				if err := m.Once(ctx); err != nil {
					logx.Errorf("物化点赞计数失败: %v", err)
				}
			}
		}
	}()
}

// Once 处理一轮，供测试或手动触发使用。
// 两类目标分开处理，一类失败不影响另一类。
func (m *Materializer) Once(ctx context.Context) error {
	if m == nil || m.redis == nil {
		return nil
	}

	videoErr := m.flush(ctx, constant.LikeTargetVideo)
	commentErr := m.flush(ctx, constant.LikeTargetComment)

	if videoErr != nil {
		return videoErr
	}
	return commentErr
}

// flush 处理某一类目标的脏标记。
//
// 只有写回成功才移除标记，失败的原样留在脏集合里下轮重试——写的是绝对值，
// 重试不会算错。
func (m *Materializer) flush(ctx context.Context, targetType uint64) error {
	dirtyKey := constant.LikeDirtyKey(targetType)

	ids, err := m.redis.SmembersCtx(ctx, dirtyKey)
	if err != nil {
		return fmt.Errorf("读取脏标记 %s: %w", dirtyKey, err)
	}
	if len(ids) == 0 {
		return nil
	}
	// 一轮处理不完就留到下一轮，避免一次拉太多
	if len(ids) > constant.LikeMaterializeBatch {
		ids = ids[:constant.LikeMaterializeBatch]
	}

	done := make([]any, 0, len(ids))
	for _, s := range ids {
		targetId, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			// 脏标记本身坏了，移除掉，否则每轮都要重新解析一次
			logx.Errorf("移除损坏的点赞脏标记 %q: %v", s, err)
			done = append(done, s)
			continue
		}

		count, err := m.redis.ScardCtx(ctx, constant.LikeSetKey(targetType, targetId))
		if err != nil {
			logx.Errorf("读取点赞计数 %d/%d 失败: %v", targetType, targetId, err)
			continue
		}

		if err := m.writeBack(ctx, targetType, targetId, count); err != nil {
			logx.Errorf("写回点赞计数 %d/%d 失败: %v", targetType, targetId, err)
			continue
		}
		done = append(done, s)
	}

	if len(done) == 0 {
		return nil
	}
	if _, err := m.redis.SremCtx(ctx, dirtyKey, done...); err != nil {
		return fmt.Errorf("清理脏标记 %s: %w", dirtyKey, err)
	}
	return nil
}

// writeBack 把绝对值写到目标的 like_count 列。
// video 表归 video 服务管，走 rpc；comment 表是本服务的，直接改。
func (m *Materializer) writeBack(ctx context.Context, targetType, targetId uint64, count int64) error {
	switch targetType {
	case constant.LikeTargetVideo:
		resp, err := m.videos.SetLikeCount(ctx, &video.SetLikeCountReq{
			VideoId: int64(targetId),
			Count:   count,
		})
		if err != nil {
			return err
		}
		if resp.GetBase().GetCode() != errno.OK.Code {
			return fmt.Errorf("写回视频点赞数被拒: %s", resp.GetBase().GetMsg())
		}
		return nil
	case constant.LikeTargetComment:
		return m.comments.SetLikeCount(ctx, targetId, count)
	default:
		return fmt.Errorf("未知的点赞目标类型 %d", targetType)
	}
}
