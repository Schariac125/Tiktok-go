// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"

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

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// todo: add your logic here and delete this line
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.RegisterResp{Base: types.Status{Code: 500, Msg: "user rpc 返回为空"}}, nil
	}
	return &types.RegisterResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
