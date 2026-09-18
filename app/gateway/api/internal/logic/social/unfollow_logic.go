// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package social

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 取关 DELETE /relation/unfollow
// 硬删除；没关注过也返回成功，重复调用不报错
func (l *UnfollowLogic) Unfollow(req *types.UnfollowReq) (*types.UnfollowResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.UnfollowResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.SocialRpc.Unfollow(l.ctx, &social.UnfollowReq{
		UserId:   userId,
		ToUserId: req.ToUserId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.UnfollowResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.UnfollowResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
