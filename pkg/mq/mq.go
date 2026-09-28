// Package mq 定义与具体中间件无关的消息队列接口。
//
// 业务代码只依赖这里的类型，不感知底层是 RocketMQ 还是别的实现，换中间件时
// 业务侧不需要改动。
//
// 接口里刻意不出现「批次」「轮询周期」「队列键」这类某一种实现特有的概念：
// 一旦出现，别的实现就填不出来，业务代码反而要跟着实现走，解耦就是假的。
package mq

import "context"

// Message 一条待投递的消息
type Message struct {
	// Topic 主题
	Topic string
	// Key 顺序键：同一个 Key 的消息保证按发送顺序被消费。
	//
	// 取「行的身份」而不是目标 id——只有同一用户对同一目标的连续操作才会
	// 互相冲突，用目标 id 会让所有用户对同一目标的操作用挤进同一个队列，
	// 白白牺牲并行度。
	Key string
	// Body 消息体，编码方式由业务自己决定
	Body []byte
}

// Publisher 投递消息
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

// Handler 消费单条消息。
//
// 返回 nil 表示这条处理成功；返回 error 表示没成功，会按实现的重试策略重投，
// 超过上限后进入死信队列。语义是「单条重试」，而不是「整批回滚」。
//
// 因为存在重试，同一条消息可能被投递多次，Handler 必须幂等。
type Handler func(ctx context.Context, body []byte) error

// Subscriber 订阅并消费
type Subscriber interface {
	// Subscribe 订阅 topic 并开始后台消费。
	// 同一个消费组内的多个实例分摊消息，不同消费组各自收到全量。
	// ctx 取消时停止消费。
	Subscribe(ctx context.Context, topic string, h Handler) error
	Close() error
}
