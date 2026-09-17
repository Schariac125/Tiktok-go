package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"tiktok-go/app/interaction/rpc/internal/config"
	"tiktok-go/app/interaction/rpc/internal/task"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/app/video/rpc/videoservice"
	gopkgtask "tiktok-go/pkg/task"
)

type ServiceContext struct {
	Config        config.Config
	CommentModel  model.CommentModel
	UserLikeModel model.UserLikeModel
	// Redis 既作 model 的单条缓存，也承载点赞的待落库队列
	Redis *redis.Redis
	// VideoRpc 组装点赞列表时读取视频信息
	VideoRpc videoservice.VideoService
	// LikeWorker 点赞队列的消费任务，接口只写队列，由它按周期刷进数据库
	LikeWorker *gopkgtask.Worker
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	commentModel := model.NewCommentModel(conn, c.CacheRedis)
	userLikeModel := model.NewUserLikeModel(conn, c.CacheRedis)
	videoRpc := videoservice.NewVideoService(zrpc.MustNewClient(c.VideoRpc))

	svcCtx := &ServiceContext{
		Config:        c,
		CommentModel:  commentModel,
		UserLikeModel: userLikeModel,
		VideoRpc:      videoRpc,
	}
	// CacheRedis 是节点数组，取第一个；未配置时不建客户端，点赞会因队列不可用而失败
	if len(c.CacheRedis) > 0 {
		r := redis.MustNewRedis(c.CacheRedis[0].RedisConf)
		svcCtx.Redis = r
		svcCtx.LikeWorker = task.NewLikeWorker(r, userLikeModel, commentModel, videoRpc)
	}
	return svcCtx
}
