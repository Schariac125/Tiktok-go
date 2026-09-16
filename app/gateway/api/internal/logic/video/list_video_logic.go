// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideoLogic {
	return &ListVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 查看用户视频 GET /video/list
// 无需登录，分页参数由 rpc 侧归一化（网关原样透传）
func (l *ListVideoLogic) ListVideo(req *types.ListVideoReq) (*types.ListVideoResp, error) {
	rpcResp, err := l.svcCtx.VideoRpc.ListVideo(l.ctx, &video.ListVideoReq{
		UserId:   req.UserId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.ListVideoResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.VideoPageData{
			Items: convertVideos(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
