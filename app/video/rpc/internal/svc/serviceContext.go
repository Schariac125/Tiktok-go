package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/app/video/rpc/internal/config"
	"tiktok-go/app/video/rpc/model"
)

type ServiceContext struct {
	Config     config.Config
	VideoModel model.VideoModel
	// Redis 供热门榜单这类「整份缓存」使用；
	// model 自带的单条记录缓存由 c.CacheRedis 单独处理，两者互不影响
	Redis *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	svcCtx := &ServiceContext{
		Config:     c,
		VideoModel: model.NewVideoModel(conn, c.CacheRedis),
	}
	// CacheRedis 是节点数组，取第一个；未配置时不建客户端，调用处会跳过缓存
	if len(c.CacheRedis) > 0 {
		svcCtx.Redis = redis.MustNewRedis(c.CacheRedis[0].RedisConf)
	}
	return svcCtx
}
