// Package task 提供「redis 队列 + 定时批量落库」的通用骨架。
//
// 高频写操作（点赞、浏览计数之类）直接落库压力太大，常见做法是先把变更写队列、
// 攒一批再批量落库。这里把这套机制抽出来：调用方只要给出队列键和单条处理函数，
// 剩下的入队、批量取用、失败重试、周期调度都由 Worker 负责。
//
// 使用方式：
//
//	w := task.NewWorker(r, task.Options{
//	    Name:     "点赞",
//	    Key:      "like:queue",
//	    Interval: 3 * time.Second,
//	    Handler:  func(ctx context.Context, item string) error { ... },
//	})
//	w.Start(ctx)                              // 启动后台消费
//	w.Push(ctx, "1:10:1:8001")                // 生产端写队列
package task

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// ErrUnavailable 队列不可用（未配置 redis），调用方据此决定是报错还是降级
var ErrUnavailable = errors.New("队列不可用")

// Handler 处理队列中的单条记录。
// 返回 error 会中断本批、留在队列里下次重试；返回 nil 才继续消费下一条。
// 所以遇到不认识的脏数据应当返回 nil 跳过，否则会一直卡住整条队列。
type Handler func(ctx context.Context, item string) error

// Options 描述一个队列消费任务
type Options struct {
	// Name 日志里用的名字，如「点赞」
	Name string
	// Key redis 队列键
	Key string
	// Batch 单批最多处理多少条，默认 200
	Batch int
	// Interval 消费周期，默认 3 秒
	Interval time.Duration
	// Handler 单条处理函数
	Handler Handler
}

// Worker 周期性把队列内容交给 Handler 处理
type Worker struct {
	redis *redis.Redis
	opts  Options
}

// NewWorker 创建队列消费任务。redis 为 nil 时 Push 与 Start 都是空操作，
// 便于在没配缓存的部署里降级（调用方需要自己决定是否还能正常工作）。
func NewWorker(r *redis.Redis, opts Options) *Worker {
	if opts.Batch <= 0 {
		opts.Batch = 200
	}
	if opts.Interval <= 0 {
		opts.Interval = 3 * time.Second
	}
	if opts.Name == "" {
		opts.Name = opts.Key
	}
	return &Worker{redis: r, opts: opts}
}

// Push 把一条记录放进队列，由后台按周期落库
func (w *Worker) Push(ctx context.Context, item string) error {
	if w == nil || w.redis == nil {
		return ErrUnavailable
	}
	_, err := w.redis.RpushCtx(ctx, w.opts.Key, item)
	return err
}

// Start 启动后台消费，随 ctx 取消而停止
func (w *Worker) Start(ctx context.Context) {
	if w == nil || w.redis == nil || w.opts.Handler == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(w.opts.Interval)
		defer ticker.Stop()

		logx.Infof("%s队列消费任务已启动，周期 %s", w.opts.Name, w.opts.Interval)
		for {
			select {
			case <-ctx.Done():
				logx.Infof("%s队列消费任务退出", w.opts.Name)
				return
			case <-ticker.C:
				if err := w.SyncOnce(ctx); err != nil {
					logx.Errorf("消费%s队列失败: %v", w.opts.Name, err)
				}
			}
		}
	}()
}

// SyncOnce 立刻消费一批，供测试或手动触发使用。
// 整批处理成功后才从队列移除；中途失败就整批留着下次重试，
// 因此 Handler 必须保证幂等——同一条记录可能被处理多次。
func (w *Worker) SyncOnce(ctx context.Context) error {
	if w == nil || w.redis == nil || w.opts.Handler == nil {
		return nil
	}

	items, err := w.redis.LrangeCtx(ctx, w.opts.Key, 0, w.opts.Batch-1)
	if err != nil || len(items) == 0 {
		return err
	}

	for _, item := range items {
		if err := w.opts.Handler(ctx, item); err != nil {
			return err
		}
	}

	return w.redis.LtrimCtx(ctx, w.opts.Key, int64(len(items)), -1)
}
