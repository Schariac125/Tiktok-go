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

type FeedVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedVideoLogic {
	return &FeedVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 首页视频流 GET /video/feed
// 无需登录；时间游标的解析与条数控制都在 rpc 侧，网关只做透传与类型转换
func (l *FeedVideoLogic) FeedVideo(req *types.FeedVideoReq) (*types.FeedVideoResp, error) {
	rpcResp, err := l.svcCtx.VideoRpc.FeedVideo(l.ctx, &video.FeedVideoReq{
		LastestTime: req.LastestTime,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.FeedVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &types.FeedVideoResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.VideoFeedData{
			Items: convertVideos(rpcResp.GetData().GetItems()),
		},
	}, nil
}
