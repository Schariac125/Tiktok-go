// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interaction

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 点赞 POST /like/action
// 点赞者身份取自登录态；rpc 侧只写队列，计数会有数秒滞后
func (l *LikeLogic) Like(req *types.LikeReq) (*types.LikeResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.LikeResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.InteractionRpc.Like(l.ctx, &interaction.LikeReq{
		UserId:    userId,
		VideoId:   req.VideoId,
		CommentId: req.CommentId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.LikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.LikeResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
