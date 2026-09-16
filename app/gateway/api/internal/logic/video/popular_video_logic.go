// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PopularVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PopularVideoLogic {
	return &PopularVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PopularVideoLogic) PopularVideo(req *types.PopularVideoReq) (resp *types.PopularVideoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
