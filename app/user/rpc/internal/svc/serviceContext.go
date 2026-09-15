package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/app/user/rpc/internal/config"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/pkg/jwt"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	Signer    *jwt.Signer
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.CacheRedis),
		Signer:    jwt.NewSigner(c.JWT.AccessSecret, c.JWT.AccessExpire, c.JWT.RefreshExpire),
	}
}
