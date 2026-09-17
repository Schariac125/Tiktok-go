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

type PublishCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishCommentLogic {
	return &PublishCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 发表评论 POST /comment/publish
// 发布者身份取自登录态；给 comment_id 表示回复某条评论
func (l *PublishCommentLogic) PublishComment(req *types.PublishCommentReq) (*types.PublishCommentResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.PublishCommentResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.InteractionRpc.PublishComment(l.ctx, &interaction.PublishCommentReq{
		UserId:    userId,
		VideoId:   req.VideoId,
		CommentId: req.CommentId,
		Content:   req.Content,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.PublishCommentResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.PublishCommentResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
