// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/app/gateway/api/internal/config"
	"tiktok-go/app/gateway/api/internal/middleware"
	"tiktok-go/app/user/rpc/userservice"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	UserRpc        userservice.UserService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware().Handle,
		UserRpc:        userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
	}
}
