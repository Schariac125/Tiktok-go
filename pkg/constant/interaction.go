package constant

// 点赞的目标类型。视频与评论共用一张点赞表，靠这个字段区分。
const (
	LikeTargetVideo   = 1 // 视频
	LikeTargetComment = 2 // 评论
)

// 点赞的落库策略。
//
// 点赞是高频操作，直接写库扛不住，因此：接口只投递消息就返回，
// 由消费端异步把变更刷进 user_like 表并更新计数。
// 消费是推送式的，滞后在毫秒级，不再有固定同步周期。
const (
	// LikeTopic 点赞变更的消息主题
	LikeTopic = "like-topic"
	// LikeGroup 点赞落库的消费组名
	LikeGroup = "like-consumer"
	// LikeMaxReconsumeTimes 单条消息最大重试次数，超过后进死信队列
	LikeMaxReconsumeTimes = 16
)

// CommentMaxLen 评论内容长度上限，与 comment 表 content 列宽一致
const CommentMaxLen = 1024
