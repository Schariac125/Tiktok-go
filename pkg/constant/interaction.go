package constant

import "time"

// 点赞的目标类型。视频与评论共用一张点赞表，靠这个字段区分。
const (
	LikeTargetVideo   = 1 // 视频
	LikeTargetComment = 2 // 评论
)

// 点赞的缓存与落库策略。
//
// 点赞是高频操作，直接写库扛不住，因此：接口只写 redis 队列就返回，
// 后台任务按 LikeSyncInterval 批量把队列刷进 user_like 表并更新计数。
// 这样计数会有不超过一个同步周期的滞后，队列积压时也不会差太久。
const (
	// LikeQueueKey 待落库的点赞变更队列，元素形如 "1:12:2:34"（动作:用户:目标类型:目标id）
	LikeQueueKey = "like:queue"
	// LikeSyncInterval 后台同步周期，即计数滞后的上限
	LikeSyncInterval = 3 * time.Second
	// LikeSyncBatch 单次最多同步多少条，避免一次取太多
	LikeSyncBatch = 200
)

// CommentMaxLen 评论内容长度上限，与 comment 表 content 列宽一致
const CommentMaxLen = 1024
