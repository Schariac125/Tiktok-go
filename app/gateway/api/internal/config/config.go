package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserRpc zrpc.RpcClientConf // user rpc 客户端配置（etcd 服务发现）
	JWT     struct {
		AccessSecret string // 与 user.rpc 的签名密钥一致
		AccessExpire time.Duration
	}
}
