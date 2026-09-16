package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/app/video/rpc/internal/config"
	"tiktok-go/app/video/rpc/model"
)

type ServiceContext struct {
	Config     config.Config
	VideoModel model.VideoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:     c,
		VideoModel: model.NewVideoModel(conn, c.CacheRedis),
	}
}
