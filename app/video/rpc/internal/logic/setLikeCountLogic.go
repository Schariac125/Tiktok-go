package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetLikeCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetLikeCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetLikeCountLogic {
	return &SetLikeCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置视频点赞数（绝对值），供互动服务的物化任务把 Redis 计数写回数据库。
//
// 这是「物化」而不是「业务写入」：权威计数在 Redis 的点赞集合里，
// 本接口只负责把快照落到 video 表的 like_count 列。因此写的是绝对值，
// 重复调用结果一致，可以安全重试。
func (l *SetLikeCountLogic) SetLikeCount(in *video.SetLikeCountReq) (*video.SetLikeCountResp, error) {
	if in.VideoId <= 0 || in.Count < 0 {
		return &video.SetLikeCountResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	if err := l.svcCtx.VideoModel.SetLikeCount(l.ctx, uint64(in.VideoId), in.Count); err != nil {
		l.Errorf("写入视频 %d 点赞数副本失败: %v", in.VideoId, err)
		return nil, err
	}

	return &video.SetLikeCountResp{Base: toStatus(errno.OK)}, nil
}
