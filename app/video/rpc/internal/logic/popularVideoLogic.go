package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"

	"github.com/zeromicro/go-zero/core/logx"
)

type PopularVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PopularVideoLogic {
	return &PopularVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 热门排行榜 GET /video/popular
func (l *PopularVideoLogic) PopularVideo(in *video.PopularVideoReq) (*video.PopularVideoResp, error) {
	// todo: add your logic here and delete this line

	return &video.PopularVideoResp{}, nil
}
