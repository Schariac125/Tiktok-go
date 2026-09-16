package constant

// 列表分页的默认值与上限
const (
	DefaultPageNum  = 1
	DefaultPageSize = 10
	MaxPageSize     = 100 // 防止一次拉取过多数据
)

// FeedSize 首页视频流每次返回的条数。
// 该接口按时间游标拉取（传了 lastest_time 就只要更新的内容），不分页，
// 条数由服务端固定，避免客户端一次拉走整站视频。
const FeedSize = 20
