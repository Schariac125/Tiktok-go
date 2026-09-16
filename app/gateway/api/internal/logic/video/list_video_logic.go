// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideoLogic {
	return &ListVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVideoLogic) ListVideo(req *types.ListVideoReq) (resp *types.ListVideoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
