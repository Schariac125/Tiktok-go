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

type ListLikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLikeLogic {
	return &ListLikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 查看某个用户点赞的视频列表 GET /like/list
// 无需登录；点赞关系由 interaction 服务提供，视频信息由它跨服务取回
func (l *ListLikeLogic) ListLike(req *types.ListLikeReq) (*types.ListLikeResp, error) {
	rpcResp, err := l.svcCtx.InteractionRpc.ListLike(l.ctx, &interaction.ListLikeReq{
		UserId:   req.UserId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListLikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.ListLikeResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.VideoPageData{
			Items: convertVideos(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
