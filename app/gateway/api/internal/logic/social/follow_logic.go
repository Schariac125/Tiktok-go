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

type FollowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 关注 POST /relation/follow
// 关注者身份取自登录态，重复关注按幂等处理
func (l *FollowLogic) Follow(req *types.FollowReq) (*types.FollowResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.FollowResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.SocialRpc.Follow(l.ctx, &social.FollowReq{
		UserId:   userId,
		ToUserId: req.ToUserId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.FollowResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.FollowResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
