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

type PopularVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PopularVideoLogic {
	return &PopularVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 热门排行榜 GET /video/popular
// 无需登录；榜单排序与 redis 缓存都在 rpc 侧，网关只做透传与类型转换
func (l *PopularVideoLogic) PopularVideo(req *types.PopularVideoReq) (*types.PopularVideoResp, error) {
	rpcResp, err := l.svcCtx.VideoRpc.PopularVideo(l.ctx, &video.PopularVideoReq{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.PopularVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.PopularVideoResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.VideoFeedData{
			Items: convertVideos(rpcResp.GetData().GetItems()),
		},
	}, nil
}
