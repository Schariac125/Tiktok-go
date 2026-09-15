// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 注册 POST /user/register
func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.RegisterResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}
	return &types.RegisterResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
