package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	VideoRpc   zrpc.RpcClientConf // 点赞列表要读视频信息

	// RocketMQ 点赞落库队列。NameServer 留空时不建客户端，
	// 点赞会因队列不可用而失败——与未配 CacheRedis 时的行为一致。
	RocketMQ struct {
		NameServer []string
	}
}
