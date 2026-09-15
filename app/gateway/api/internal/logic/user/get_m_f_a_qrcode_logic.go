// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

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

// 获取绑定 MFA 时所需的二维码 GET /auth/mfa/qrcode
func (l *GetMFAQrcodeLogic) GetMFAQrcode() (resp *types.GetMFAQrcodeResp, err error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.GetMFAQrcodeResp{Base: types.Status{Code: errno.ErrUnauthorized.Code, Msg: errno.ErrUnauthorized.Msg}}, nil
	}

	rpcResp, err := l.svcCtx.UserRpc.GetMFAQrcode(l.ctx, &user.GetMFAQrcodeReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.GetMFAQrcodeResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}

	return &types.GetMFAQrcodeResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.MFAQrcode{
			Secret: rpcResp.Secret,
			Qrcode: rpcResp.Qrcode,
		},
	}, nil
}
