package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrLikeCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrLikeCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrLikeCountLogic {
	return &IncrLikeCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增减视频点赞数，供互动服务的同步任务调用。
// 视频表归本服务管，互动服务不直接改，所以走这个接口。
func (l *IncrLikeCountLogic) IncrLikeCount(in *video.IncrLikeCountReq) (*video.IncrLikeCountResp, error) {
	if in.VideoId <= 0 || in.Delta == 0 {
		return &video.IncrLikeCountResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	if err := l.svcCtx.VideoModel.IncrLikeCount(l.ctx, uint64(in.VideoId), in.Delta); err != nil {
		l.Errorf("更新视频 %d 点赞数失败: %v", in.VideoId, err)
		return nil, err
	}

	return &video.IncrLikeCountResp{Base: toStatus(errno.OK)}, nil
}
