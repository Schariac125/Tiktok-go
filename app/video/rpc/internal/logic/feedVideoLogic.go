package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedVideoLogic {
	return &FeedVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 首页视频流 GET /video/feed
func (l *FeedVideoLogic) FeedVideo(in *video.FeedVideoReq) (*video.FeedVideoResp, error) {
	// todo: add your logic here and delete this line

	return &video.FeedVideoResp{}, nil
}
