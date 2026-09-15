// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMFALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindMFALogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMFALogic {
	return &BindMFALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindMFALogic) BindMFA(req *types.BindMFAReq) (resp *types.BindMFAResp, err error) {
	// todo: add your logic here and delete this line

	return
}
