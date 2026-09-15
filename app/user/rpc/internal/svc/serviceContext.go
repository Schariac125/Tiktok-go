package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/app/user/rpc/internal/config"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/pkg/jwt"
	"tiktok-go/pkg/storage"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	Signer    *jwt.Signer
	OSS       storage.Object
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.CacheRedis),
		Signer:    jwt.NewSigner(c.JWT.AccessSecret, c.JWT.AccessExpire, c.JWT.RefreshExpire),
		// 惰性初始化：OSS 凭据缺失不影响服务启动，上传时才报错
		OSS: storage.NewLazyOSS(c.OSS),
	}
}
