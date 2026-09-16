package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/pkg/storage"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	OSS        storage.OSSConfig // 视频、封面等对象存储
}
