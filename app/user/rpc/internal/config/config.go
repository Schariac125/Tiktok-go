package config

import (
	"time"

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
	JWT        struct {
		AccessSecret  string        // 签名密钥
		AccessExpire  time.Duration // access token 有效期
		RefreshExpire time.Duration // refresh token 有效期
	}
	OSS storage.OSSConfig // 头像等对象存储
}
