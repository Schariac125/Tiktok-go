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

// 绑定 MFA POST /auth/mfa/bind
func (l *BindMFALogic) BindMFA(req *types.BindMFAReq) (resp *types.BindMFAResp, err error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.BindMFAResp{Base: types.Status{Code: errno.ErrUnauthorized.Code, Msg: errno.ErrUnauthorized.Msg}}, nil
	}

	rpcResp, err := l.svcCtx.UserRpc.BindMFA(l.ctx, &user.BindMFAReq{
		UserId: userId,
		Code:   req.Code,
		Secret: req.Secret,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.BindMFAResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}
	return &types.BindMFAResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
