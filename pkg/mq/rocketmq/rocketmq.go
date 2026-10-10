// Package rocketmq 是 mq 接口的 RocketMQ 实现。
//
// 顺序性靠两头配合，缺一不可：
//   - 生产端挂哈希队列选择器，按 Message.Key 选队列，同一 Key 的消息必定同队列；
//   - 消费端开顺序消费，一个队列单线程按序处理。
//
// 只做前者，消费端仍并发，同队列内照样乱序。
package rocketmq

import (
	"context"
	"fmt"
	"net"

	"tiktok-go/pkg/mq"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// Config 连接配置
type Config struct {
	// NameServer NameServer 地址列表，如 ["127.0.0.1:9876"]。
	// 主机名也可以（如容器里的 ["rocketmq-namesrv:9876"]），本包会先解析成 IP。
	NameServer []string
	// Group 生产者组 / 消费者组名
	Group string
	// MaxReconsumeTimes 单条消息的最大重试次数，超过后进入死信队列
	MaxReconsumeTimes int32
}

// resolveNameServer 把 NameServer 地址中的主机名解析成 IP。
//
// RocketMQ 的 Go 客户端要求 NameServer 地址里必须能匹配到 IPv4/IPv6 字面量
// （primitive.verifyIP 用正则去找），主机名直接判为 "IP addr error" 拒绝。
// 而容器编排里习惯写服务名，所以在这里先解析一次。
//
// 只在建连时解析，容器 IP 变化后需要重启进程——对无状态的 RPC 服务可以接受。
// 注意这个限制只作用于 NameServer：broker 地址是运行时由 NameServer 下发的，
// 不受校验，所以 docker-compose 里 broker 注册 host.docker.internal 仍然可行。
func resolveNameServer(addrs []string) ([]string, error) {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("NameServer 地址 %q 应为 host:port: %w", addr, err)
		}
		if net.ParseIP(host) != nil {
			out = append(out, addr)
			continue
		}
		ips, err := net.LookupHost(host)
		if err != nil {
			return nil, fmt.Errorf("解析 NameServer 主机名 %q: %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("NameServer 主机名 %q 未解析到任何地址", host)
		}
		out = append(out, net.JoinHostPort(ips[0], port))
	}
	return out, nil
}

// NewPublisher 创建生产者。
//
// 挂哈希队列选择器后，消息按 Key 落到固定队列；Key 为空时退化为随机，
// 顺序性不再有保证，所以业务侧必须给 Key。
func NewPublisher(c Config) (mq.Publisher, error) {
	nameServer, err := resolveNameServer(c.NameServer)
	if err != nil {
		return nil, err
	}

	p, err := rocketmq.NewProducer(
		producer.WithNameServer(nameServer),
		producer.WithGroupName(c.Group),
		producer.WithRetry(2),
		producer.WithQueueSelector(producer.NewHashQueueSelector()),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 rocketmq 生产者: %w", err)
	}
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("启动 rocketmq 生产者: %w", err)
	}
	return &publisher{p: p}, nil
}

type publisher struct {
	p rocketmq.Producer
}

func (x *publisher) Publish(ctx context.Context, msg mq.Message) error {
	m := &primitive.Message{
		Topic: msg.Topic,
		Body:  msg.Body,
	}
	// ShardingKey 决定进哪个队列，Keys 只是便于按 key 检索，两者都设
	m.WithShardingKey(msg.Key)
	m.WithKeys([]string{msg.Key})

	if _, err := x.p.SendSync(ctx, m); err != nil {
		return fmt.Errorf("投递消息到 %s: %w", msg.Topic, err)
	}
	return nil
}

func (x *publisher) Close() error {
	return x.p.Shutdown()
}

// NewSubscriber 创建顺序消费者。
func NewSubscriber(c Config) (mq.Subscriber, error) {
	nameServer, err := resolveNameServer(c.NameServer)
	if err != nil {
		return nil, err
	}

	s := &subscriber{
		nameServer:        nameServer,
		group:             c.Group,
		maxReconsumeTimes: c.MaxReconsumeTimes,
	}
	return s, nil
}

type subscriber struct {
	nameServer        []string
	group             string
	maxReconsumeTimes int32

	consumer rocketmq.PushConsumer
	started  bool
}

// Subscribe 订阅并开始后台消费，ctx 取消时停止。
func (s *subscriber) Subscribe(ctx context.Context, topic string, h mq.Handler) error {
	opts := []consumer.Option{
		consumer.WithNameServer(s.nameServer),
		consumer.WithGroupName(s.group),
		// 顺序消费：同一队列单线程按序处理，配合生产端的哈希选择器保证顺序。
		// 代价是队头阻塞——某条重试时同队列后面的消息要等，这是「顺序」本身的
		// 含义，换来的是失败只影响一个队列而非全部，且有重试上限兜底。
		consumer.WithConsumerOrder(true),
		consumer.WithConsumerModel(consumer.Clustering),
		// 客户端默认 ConsumeFromLastOffset，新消费组会从最新位点开始，
		// 把组建立之前已投递的消息全部跳过。落库队列不能有这种丢数据的窗口，
		// 所以从最早开始；Handler 幂等，重放只是白跑一遍，不会出错。
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	}
	if s.maxReconsumeTimes > 0 {
		opts = append(opts, consumer.WithMaxReconsumeTimes(s.maxReconsumeTimes))
	}

	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return fmt.Errorf("创建 rocketmq 消费者: %w", err)
	}
	s.consumer = c

	// 顺序消费下一次只投一条，但仍按切片处理，避免依赖批大小恰好为 1
	err = c.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt,
	) (consumer.ConsumeResult, error) {
		for _, m := range msgs {
			if err := h(ctx, m.Body); err != nil {
				// 返回 error 让 broker 稍后重投这条，重试超限后进死信队列
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		return fmt.Errorf("订阅 %s: %w", topic, err)
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("启动 rocketmq 消费者: %w", err)
	}
	s.started = true

	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	return nil
}

func (s *subscriber) Close() error {
	if !s.started {
		return nil
	}
	s.started = false
	// Shutdown 会把已消费进度同步回 broker，失败只记录不影响退出
	return s.consumer.Shutdown()
}
