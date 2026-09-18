package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/pkg/storage"
)

type Config struct {
	rest.RestConf
	UserRpc        zrpc.RpcClientConf // user rpc 客户端配置（etcd 服务发现）
	VideoRpc       zrpc.RpcClientConf // video rpc 客户端配置（etcd 服务发现）
	InteractionRpc zrpc.RpcClientConf // interaction rpc 客户端配置（etcd 服务发现）
	SocialRpc      zrpc.RpcClientConf // social rpc 客户端配置（etcd 服务发现）
	OSS            storage.OSSConfig  // 视频、封面等对象存储
	JWT      struct {
		AccessSecret string // 与 user.rpc 的签名密钥一致
		AccessExpire time.Duration
	}
}
