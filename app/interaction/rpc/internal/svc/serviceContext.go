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
	// 配置了却连不上则直接启动失败——带着坏掉的队列跑起来，故障会推迟到
	// 用户点赞时才暴露，且接口还返回成功。
	if len(c.RocketMQ.NameServer) > 0 {
		svcCtx.startLikeQueue(ctx, c.RocketMQ.NameServer, userLikeModel, commentModel, videoRpc)
	}
	return svcCtx
}

// startLikeQueue 建起点赞队列的生产者与消费者。
// 生产者与消费者是两个独立对象，各自持有到 broker 的连接。
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
	if err := sub.Subscribe(ctx, constant.LikeTopic,
		task.NewLikeHandler(likes, comments, videos)); err != nil {
		panic(fmt.Errorf("订阅点赞队列失败: %w", err))
	}

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
