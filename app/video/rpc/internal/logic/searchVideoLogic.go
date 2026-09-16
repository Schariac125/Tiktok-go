package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchVideoLogic {
	return &SearchVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 搜索视频 POST /video/search
func (l *SearchVideoLogic) SearchVideo(in *video.SearchVideoReq) (*video.SearchVideoResp, error) {
	// todo: add your logic here and delete this line

	return &video.SearchVideoResp{}, nil
}
