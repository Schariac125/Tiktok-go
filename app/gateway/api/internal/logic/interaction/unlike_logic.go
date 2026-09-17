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

type UnlikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 取消点赞 DELETE /like/delete
func (l *UnlikeLogic) Unlike(req *types.UnlikeReq) (*types.UnlikeResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.UnlikeResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.InteractionRpc.Unlike(l.ctx, &interaction.UnlikeReq{
		UserId:    userId,
		VideoId:   req.VideoId,
		CommentId: req.CommentId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.UnlikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.UnlikeResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
