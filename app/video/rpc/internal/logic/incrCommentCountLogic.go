package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrCommentCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrCommentCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrCommentCountLogic {
	return &IncrCommentCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增减视频评论数，供互动服务在发表/删除评论（含级联删除子评论）时调用
func (l *IncrCommentCountLogic) IncrCommentCount(in *video.IncrCommentCountReq) (*video.IncrCommentCountResp, error) {
	if in.VideoId <= 0 || in.Delta == 0 {
		return &video.IncrCommentCountResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	if err := l.svcCtx.VideoModel.IncrCommentCount(l.ctx, uint64(in.VideoId), in.Delta); err != nil {
		l.Errorf("更新视频 %d 评论数失败: %v", in.VideoId, err)
		return nil, err
	}

	return &video.IncrCommentCountResp{Base: toStatus(errno.OK)}, nil
}
