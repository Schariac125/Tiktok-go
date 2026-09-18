package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/app/social/rpc/internal/config"
	"tiktok-go/app/social/rpc/model"
	"tiktok-go/app/user/rpc/userservice"
)

type ServiceContext struct {
	Config        config.Config
	RelationModel model.UserRelationModel
	// UserRpc 列表只存了 user_id，用户名要跨服务取
	UserRpc userservice.UserService
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:        c,
		RelationModel: model.NewUserRelationModel(conn, c.CacheRedis),
		UserRpc:       userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
	}
}
