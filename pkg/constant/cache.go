package constant

import "time"

// PopularRankCacheKey 热门榜单整份缓存的键。
// 沿用 goctl 生成的 model 缓存键（cache:video:id:<id>）同样的 cache: 前缀，
// 便于在 redis 里一眼区分缓存数据与业务数据。
const PopularRankCacheKey = "cache:video:popular"

// 热门排行榜的缓存策略。
//
// 榜单是「整份读、频繁读、可容忍短时陈旧」的数据，因此缓存整份 Top N 列表，
// 分页在内存里切片——而不是按页缓存（那会让 key 随页码×每页数量爆炸，命中率极低）。
const (
	// PopularRankSize 榜单缓存的条数，超出部分的分页不再从榜单出
	PopularRankSize = 100
	// PopularRankTTL 榜单缓存有效期，过期后重新查库。
	// 排行榜本就允许滞后，这里不做过期时的主动失效。
	PopularRankTTL = 5 * time.Minute

	// RedisOpTimeout 单次 redis 操作的超时。
	// redis 不可用时客户端会连着重试，耗时能超过 zrpc 的整体超时（默认 2s），
	// 那样接口会直接 504 而不是回退查库——所以每次操作都单独设个短超时。
	RedisOpTimeout = 200 * time.Millisecond
)
