// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedVideoLogic {
	return &FeedVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FeedVideoLogic) FeedVideo(req *types.FeedVideoReq) (resp *types.FeedVideoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
