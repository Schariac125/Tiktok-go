package logic

import (
	"context"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMFALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindMFALogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMFALogic {
	return &BindMFALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定 MFA POST /auth/mfa/bind
func (l *BindMFALogic) BindMFA(in *user.BindMFAReq) (*user.BindMFAResp, error) {
	// todo: add your logic here and delete this line

	return &user.BindMFAResp{}, nil
}
