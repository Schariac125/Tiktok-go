package constant

import (
	"fmt"
	"time"
)

// 点赞的目标类型。视频与评论共用一张点赞表，靠这个字段区分。
const (
	LikeTargetVideo   = 1 // 视频
	LikeTargetComment = 2 // 评论
)

// 点赞计数的 Redis 结构：每个目标一个集合，成员是点赞用户 id，计数 = SCARD。
//
// 用集合而不是裸计数器，是因为集合操作天然幂等：MQ 重投时 SADD 返回 0，
// 状态不变，计数由集合大小推导，不可能和状态对不上。裸计数器则是「判重 + 自增」
// 两步——第一步判重成功后第二步失败，重投时会被判重短路，导致永久少计数。
//
// 代价是内存 = 点赞数（每个赞一个 member），当前规模可忽略。
const (
	likeSetPrefix   = "like:set:"   // like:set:<目标类型>:<目标id>
	likeDirtyPrefix = "like:dirty:" // like:dirty:<目标类型>
)

// LikeSetKey 点赞集合的 key。
//
// 这是跨服务契约：由 interaction-rpc 的消费端写入，video-rpc 组装视频响应时
// 读取。改动必须两边同步。
func LikeSetKey(targetType, targetId uint64) string {
	return fmt.Sprintf("%s%d:%d", likeSetPrefix, targetType, targetId)
}

// LikeDirtyKey 脏目标集合的 key，按目标类型分开。
// 物化任务从它里面取出「计数变过的目标」，刷完 MySQL 后移除。
func LikeDirtyKey(targetType uint64) string {
	return fmt.Sprintf("%s%d", likeDirtyPrefix, targetType)
}

// 点赞的落库策略。
//
// 点赞是高频操作，直接写库扛不住，因此：接口只投递消息就返回，
// 由消费端异步把变更刷进 user_like 表并更新 Redis 计数。
// 消费是推送式的，滞后在毫秒级，不再有固定同步周期。
const (
	// LikeTopic 点赞变更的消息主题
	LikeTopic = "like-topic"
	// LikeGroup 点赞落库的消费组名
	LikeGroup = "like-consumer"
	// LikeMaxReconsumeTimes 单条消息最大重试次数，超过后进死信队列
	LikeMaxReconsumeTimes = 16
	// LikeMaterializeInterval 把 Redis 计数物化回 MySQL 的周期。
	// MySQL 的 like_count 只是副本，读路径以 Redis 为准，所以这个周期
	// 只影响「冷备份有多旧」，不影响用户看到的计数。
	LikeMaterializeInterval = 10 * time.Second
	// LikeMaterializeBatch 单轮物化最多处理多少个目标
	LikeMaterializeBatch = 500
)

// CommentMaxLen 评论内容长度上限，与 comment 表 content 列宽一致
const CommentMaxLen = 1024
