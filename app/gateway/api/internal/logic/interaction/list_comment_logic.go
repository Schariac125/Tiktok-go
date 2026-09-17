// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package interaction

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentLogic {
	return &ListCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 查看评论列表 GET /comment/list
// 无需登录；给 video_id 取一级评论，给 comment_id 取该评论的回复
func (l *ListCommentLogic) ListComment(req *types.ListCommentReq) (*types.ListCommentResp, error) {
	rpcResp, err := l.svcCtx.InteractionRpc.ListComment(l.ctx, &interaction.ListCommentReq{
		VideoId:   req.VideoId,
		CommentId: req.CommentId,
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListCommentResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.ListCommentResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.CommentListData{
			Items: convertComments(rpcResp.GetItems()),
		},
	}, nil
}
