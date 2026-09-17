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
}
