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

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 登录 POST /user/login
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
		Code:     req.Code,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.LoginResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}

	resp = &types.LoginResp{
		Base:         types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		AccessToken:  rpcResp.AccessToken,
		RefreshToken: rpcResp.RefreshToken,
	}
	if rpcResp.Data != nil {
		resp.Data = types.User{
			Id:        rpcResp.Data.Id,
			Username:  rpcResp.Data.Username,
			AvatarUrl: rpcResp.Data.AvatarUrl,
			CreateAt:  rpcResp.Data.CreateAt,
			UpdatedAt: rpcResp.Data.UpdatedAt,
			DeleteAt:  rpcResp.Data.DeleteAt,
		}
	}
	return resp, nil
}
