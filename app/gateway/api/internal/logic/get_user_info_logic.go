// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 用户信息，只允许用户获取自己的 GET /user/info
func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoReq) (resp *types.GetUserInfoResp, err error) {
	// 只允许获取自己的信息：用户 id 以登录态（token）为准，忽略请求体的 user_id
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.GetUserInfoResp{Base: types.Status{Code: errno.ErrUnauthorized.Code, Msg: errno.ErrUnauthorized.Msg}}, nil
	}

	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.GetUserInfoResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}

	resp = &types.GetUserInfoResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
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
