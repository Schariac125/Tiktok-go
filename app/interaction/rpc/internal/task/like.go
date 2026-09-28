// Package task 组装互动模块自己的点赞队列逻辑。
//
// 队列的投递与消费由 pkg/mq 提供（当前实现是 RocketMQ），这里只负责点赞特有的
// 部分：消息格式、顺序键的取法、幂等落库，以及维护 Redis 里的点赞集合。
package task

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/mq"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
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
	redis *redis.Redis
	likes model.UserLikeModel
}

// NewLikeHandler 构造点赞变更的消费逻辑。
// 点赞是高频操作，接口只投递消息就返回，由这个 handler 异步落库并维护计数。
func NewLikeHandler(r *redis.Redis, likes model.UserLikeModel) mq.Handler {
	h := &likeHandler{redis: r, likes: likes}
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

// handle 处理一条点赞变更，做两件事：
//
//	1. 落 user_like 表 —— 唯一持久真相
//	2. 改 Redis 点赞集合 —— 计数的权威来源，计数 = SCARD
//
// 两步各自幂等，且不能互相短路：MQ 是 at-least-once，上一次可能正是卡在第 2 步
// 才重投的。若因为第 1 步「没真的改」就跳过第 2 步，那次失败就永远补不上了。
func (h *likeHandler) handle(ctx context.Context, body []byte) error {
	action, userId, targetType, targetId, err := decodeItem(string(body))
	if err != nil {
		// 解析不了说明是坏数据，重投多少次都不会成功，只会白白占着重试名额
		// 并阻塞本队列。记一笔后确认掉，不再重试。
		logx.Errorf("丢弃无法解析的点赞队列消息 %q: %v", body, err)
		return nil
	}

	// 1. 持久化。InsertIgnore / DeleteOne 靠唯一索引去重，重投不会写重复。
	if action == actionLike {
		_, err = h.likes.InsertIgnore(ctx, userId, targetType, targetId)
	} else {
		_, err = h.likes.DeleteOne(ctx, userId, targetType, targetId)
	}
	if err != nil {
		return err
	}

	// 2. 计数
	if err := EnsureSetLoaded(ctx, h.redis, h.likes, targetType, targetId); err != nil {
		return err
	}
	if err := h.updateCountSet(ctx, action, userId, targetType, targetId); err != nil {
		return err
	}

	// 3. 标记该目标的计数待物化
	//
	// 不区分「计数是否真的变了」：重复消息很少，多刷一次只是多一条 UPDATE；
	// 而漏刷会让 MySQL 副本一直停在旧值。这个标记失败也不重投——Redis 计数
	// 本身已经是准的，标记丢了只影响副本的新鲜度。
	if _, err := h.redis.SaddCtx(ctx, constant.LikeDirtyKey(targetType), targetId); err != nil {
		logx.Errorf("标记点赞计数待物化失败 target=%d/%d: %v", targetType, targetId, err)
	}
	return nil
}

// EnsureSetLoaded 保证点赞集合是完整的：key 不存在时从 user_like 重建。
//
// 集合是 user_like 的投影，两者在算法上等价，但 Redis 丢数据时集合会整个不见，
// 而 user_like 还在。此时 SCARD 返回 0 且**不报错**——那不代表真的没人赞过，
// 只是集合没了。所以凡是「读计数」或「在计数上做增量」的地方都必须先过这一步，
// 否则会把 0 当成真值用出去。
//
// 调用方若刚写过 user_like，重建出来的集合已经包含本次变更，
// 随后的 SADD/SREM 会返回 0，这是正常的。
func EnsureSetLoaded(ctx context.Context, r *redis.Redis,
	likes model.UserLikeModel, targetType, targetId uint64) error {

	key := constant.LikeSetKey(targetType, targetId)

	exists, err := r.ExistsCtx(ctx, key)
	if err != nil {
		return fmt.Errorf("检查点赞集合 %s: %w", key, err)
	}
	if exists {
		return nil
	}

	userIds, err := likes.FindUserIdsByTarget(ctx, targetType, targetId)
	if err != nil {
		return fmt.Errorf("重建点赞集合 %s: %w", key, err)
	}
	if len(userIds) == 0 {
		// 目标确实没人赞过，不是丢失，交给调用方去建集合
		return nil
	}

	vals := make([]any, 0, len(userIds))
	for _, id := range userIds {
		vals = append(vals, id)
	}
	if _, err := r.SaddCtx(ctx, key, vals...); err != nil {
		return fmt.Errorf("重建点赞集合 %s: %w", key, err)
	}

	logx.Infof("从 user_like 重建点赞集合 %s，%d 个成员", key, len(userIds))
	return nil
}

// updateCountSet 把变更写进目标的点赞集合。
// SADD/SREM 都幂等，重投时返回 0、集合状态不变，因此不会重复计数。
func (h *likeHandler) updateCountSet(ctx context.Context, action int, userId, targetType, targetId uint64) error {
	key := constant.LikeSetKey(targetType, targetId)

	var err error
	if action == actionLike {
		_, err = h.redis.SaddCtx(ctx, key, userId)
	} else {
		_, err = h.redis.SremCtx(ctx, key, userId)
	}
	if err != nil {
		return fmt.Errorf("更新点赞集合 %s: %w", key, err)
	}
	return nil
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
