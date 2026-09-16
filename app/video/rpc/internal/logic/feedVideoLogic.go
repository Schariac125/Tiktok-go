package logic

import (
	"context"
	"strings"
	"time"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedVideoLogic {
	return &FeedVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 首页视频流 GET /video/feed
// 按发布时间倒序返回最新的一批；带上 lastest_time 时只返回该时间之后的，
// 供客户端下拉刷新。游标格式与响应里输出的时间格式一致。
func (l *FeedVideoLogic) FeedVideo(in *video.FeedVideoReq) (*video.FeedVideoResp, error) {
	var since time.Time
	if s := strings.TrimSpace(in.LastestTime); s != "" {
		t, err := time.ParseInLocation(timeLayout, s, time.Local)
		if err != nil {
			return &video.FeedVideoResp{
				Base: toStatus(errno.ErrInvalidParam.WithMsg("lastest_time 格式应为 " + timeLayout)),
			}, nil
		}
		since = t
	}

	list, err := l.svcCtx.VideoModel.FeedLatest(l.ctx, since, constant.FeedSize)
	if err != nil {
		return nil, err
	}

	return &video.FeedVideoResp{
		Base: toStatus(errno.OK),
		Data: &video.VideoFeedData{Items: convertVideos(list)},
	}, nil
}
