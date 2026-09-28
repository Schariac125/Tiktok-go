package svc

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"

	"tiktok-go/app/interaction/rpc/internal/config"
	"tiktok-go/app/interaction/rpc/internal/task"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/app/video/rpc/videoservice"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/mq"
	"tiktok-go/pkg/mq/rocketmq"
)

type ServiceContext struct {
	Config        config.Config
	CommentModel  model.CommentModel
	UserLikeModel model.UserLikeModel
	// Redis 作 model 的单条记录缓存
	Redis *redis.Redis
	// VideoRpc 组装点赞列表时读取视频信息
	VideoRpc videoservice.VideoService
	// LikePublisher 点赞接口投递变更用；未配置 MQ 时为 nil，投递会返回 ErrUnavailable
	LikePublisher mq.Publisher
	// likeSubscriber 点赞落库的消费端，随 ctx 取消而停止
	likeSubscriber mq.Subscriber
}

func NewServiceContext(ctx context.Context, c config.Config) *ServiceContext {
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
	// CacheRedis 是节点数组，取第一个；未配置时不建客户端，调用处会跳过缓存
	if len(c.CacheRedis) > 0 {
		svcCtx.Redis = redis.MustNewRedis(c.CacheRedis[0].RedisConf)
	}

	// 未配置 MQ 时不建客户端，点赞会因队列不可用而失败。
	// 配了 MQ 却没配 Redis 是配置错误：点赞集合是计数的权威来源，没有它整条
	// 链路算不出计数。这种情况直接启动失败，而不是等用户点赞时才发现。
	if len(c.RocketMQ.NameServer) > 0 {
		if svcCtx.Redis == nil {
			panic("配置了 RocketMQ 但未配置 CacheRedis：点赞计数依赖 Redis 的点赞集合")
		}
		svcCtx.startLikeQueue(ctx, c.RocketMQ.NameServer, userLikeModel, commentModel, videoRpc)
	}
	return svcCtx
}

// startLikeQueue 建起点赞队列的生产者、消费者，以及计数的物化任务。
func (s *ServiceContext) startLikeQueue(ctx context.Context, nameServer []string,
	likes model.UserLikeModel, comments model.CommentModel, videos videoservice.VideoService) {

	cfg := rocketmq.Config{
		NameServer:        nameServer,
		Group:             constant.LikeGroup,
		MaxReconsumeTimes: constant.LikeMaxReconsumeTimes,
	}

	pub, err := rocketmq.NewPublisher(cfg)
	if err != nil {
		panic(fmt.Errorf("初始化点赞队列生产者失败: %w", err))
	}
	sub, err := rocketmq.NewSubscriber(cfg)
	if err != nil {
		panic(fmt.Errorf("初始化点赞队列消费者失败: %w", err))
	}
	// 消费端只做两件事：落 user_like、维护 Redis 点赞集合。
	// 计数的写回交给了物化任务，所以要单独拿 comment/video 客户端。
	if err := sub.Subscribe(ctx, constant.LikeTopic, task.NewLikeHandler(s.Redis, likes)); err != nil {
		panic(fmt.Errorf("订阅点赞队列失败: %w", err))
	}

	task.NewMaterializer(s.Redis, likes, comments, videos).Start(ctx)

	s.LikePublisher = pub
	s.likeSubscriber = sub
}

// Close 释放队列资源，进程退出时调用
func (s *ServiceContext) Close() {
	if s.likeSubscriber != nil {
		_ = s.likeSubscriber.Close()
	}
	if s.LikePublisher != nil {
		_ = s.LikePublisher.Close()
	}
}
