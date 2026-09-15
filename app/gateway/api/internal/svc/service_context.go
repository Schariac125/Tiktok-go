package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/app/gateway/api/internal/config"
	"tiktok-go/app/gateway/api/internal/middleware"
	"tiktok-go/app/user/rpc/userservice"
	"tiktok-go/pkg/jwt"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	UserRpc        userservice.UserService
	Signer         *jwt.Signer
}

func NewServiceContext(c config.Config) *ServiceContext {
	signer := jwt.NewSigner(c.JWT.AccessSecret, c.JWT.AccessExpire, 0)
	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware(signer).Handle,
		UserRpc:        userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
		Signer:         signer,
	}
}
