// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMFAQrcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMFAQrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMFAQrcodeLogic {
	return &GetMFAQrcodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMFAQrcodeLogic) GetMFAQrcode() (resp *types.GetMFAQrcodeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
