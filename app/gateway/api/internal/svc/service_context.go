package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/app/gateway/api/internal/config"
	"tiktok-go/app/gateway/api/internal/middleware"
	"tiktok-go/app/interaction/rpc/interactionservice"
	"tiktok-go/app/user/rpc/userservice"
	"tiktok-go/app/video/rpc/videoservice"
	"tiktok-go/pkg/jwt"
	"tiktok-go/pkg/storage"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	UserRpc        userservice.UserService
	VideoRpc       videoservice.VideoService
	InteractionRpc interactionservice.InteractionService
	Signer         *jwt.Signer
	OSS            storage.Object
}

func NewServiceContext(c config.Config) *ServiceContext {
	signer := jwt.NewSigner(c.JWT.AccessSecret, c.JWT.AccessExpire, 0)
	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware(signer).Handle,
		UserRpc:        userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
		VideoRpc:       videoservice.NewVideoService(zrpc.MustNewClient(c.VideoRpc)),
		InteractionRpc: interactionservice.NewInteractionService(zrpc.MustNewClient(c.InteractionRpc)),
		Signer:         signer,
		// 惰性初始化：OSS 凭据缺失不影响服务启动，上传时才报错
		OSS: storage.NewLazyOSS(c.OSS),
	}
}
