// Package task 组装互动模块自己的队列消费逻辑。
//
// 队列的投递与消费由 pkg/mq 提供（当前实现是 RocketMQ），这里只负责点赞特有的
// 部分：队列项的格式、顺序键的取法、幂等落库、以及按目标类型更新计数。
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
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/mq"

	"github.com/zeromicro/go-zero/core/logx"
)

// ErrUnavailable 队列不可用（未配置 MQ），调用方据此决定是报错还是降级
var ErrUnavailable = errors.New("队列不可用")

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

// NewLikeHandler 构造点赞变更的消费逻辑。
// 点赞是高频操作，接口只投递消息就返回，由这个 handler 异步刷进数据库并更新计数。
func NewLikeHandler(likes model.UserLikeModel, comments model.CommentModel, videos videoservice.VideoService) mq.Handler {
	h := &likeHandler{likes: likes, comments: comments, videos: videos}
	return h.handle
}

// PushLike 投递一次点赞
func PushLike(ctx context.Context, p mq.Publisher, userId, targetType, targetId uint64) error {
	return push(ctx, p, actionLike, userId, targetType, targetId)
}

// PushUnlike 投递一次取消点赞
func PushUnlike(ctx context.Context, p mq.Publisher, userId, targetType, targetId uint64) error {
	return push(ctx, p, actionUnlike, userId, targetType, targetId)
}

func push(ctx context.Context, p mq.Publisher, action int, userId, targetType, targetId uint64) error {
	if p == nil {
		return ErrUnavailable
	}
	return p.Publish(ctx, mq.Message{
		Topic: constant.LikeTopic,
		Key:   likeKey(userId, targetType, targetId),
		Body:  []byte(encodeItem(action, userId, targetType, targetId)),
	})
}

// likeKey 顺序键，取「行的身份」。
//
// 只有同一用户对同一目标的连续操作才会互相冲突（先点赞再取消，顺序反了最终
// 状态就错），不同用户操作的是不同的行。所以键取三者组合而非目标 id——
// 用目标 id 会让所有用户对同一目标的操作用挤进同一个队列，白白牺牲并行度。
func likeKey(userId, targetType, targetId uint64) string {
	return fmt.Sprintf("%d:%d:%d", userId, targetType, targetId)
}

// handle 落库一条变更，并按幂等结果调整计数。
//
// 可能被重复调用：MQ 是 at-least-once，消费失败会重投。因此依赖
// InsertIgnore / DeleteOne 的影响行数判断是否真的变更过，而不是假设只调用一次。
func (h *likeHandler) handle(ctx context.Context, body []byte) error {
	action, userId, targetType, targetId, err := decodeItem(string(body))
	if err != nil {
		// 解析不了说明是坏数据，重投多少次都不会成功，只会白白占着重试名额
		// 并阻塞本队列。记一笔后确认掉，不再重试。
		logx.Errorf("丢弃无法解析的点赞队列消息 %q: %v", body, err)
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

// encodeItem 把一条变更编码成消息体 "动作:用户:目标类型:目标id"
func encodeItem(action int, userId, targetType, targetId uint64) string {
	return fmt.Sprintf("%d:%d:%d:%d", action, userId, targetType, targetId)
}

// decodeItem 解析消息体
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
