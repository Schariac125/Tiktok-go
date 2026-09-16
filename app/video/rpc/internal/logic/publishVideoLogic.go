package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 视频投稿 POST /video/publish
func (l *PublishVideoLogic) PublishVideo(in *video.PublishVideoReq) (*video.PublishVideoResp, error) {
	// todo: add your logic here and delete this line

	return &video.PublishVideoResp{}, nil
}
