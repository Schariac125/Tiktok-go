package logic

import (
	"context"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMFAQrcodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMFAQrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMFAQrcodeLogic {
	return &GetMFAQrcodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取绑定 MFA 时所需的二维码 GET /auth/mfa/qrcode
func (l *GetMFAQrcodeLogic) GetMFAQrcode(in *user.GetMFAQrcodeReq) (*user.GetMFAQrcodeResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetMFAQrcodeResp{}, nil
}
