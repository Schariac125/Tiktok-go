package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/app/video/rpc/internal/config"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/pkg/storage"
)

type ServiceContext struct {
	Config     config.Config
	VideoModel model.VideoModel
	OSS        storage.Object
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:     c,
		VideoModel: model.NewVideoModel(conn, c.CacheRedis),
		// 惰性初始化：OSS 凭据缺失不影响服务启动，上传时才报错
		OSS: storage.NewLazyOSS(c.OSS),
	}
}
