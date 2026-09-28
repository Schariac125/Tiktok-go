package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetVideoLogic {
	return &BatchGetVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按 id 批量查视频，供互动服务组装点赞列表。
// 不保证返回顺序与入参一致，调用方按 id 自己对齐。
// 已删除的视频不会返回，调用方需要容忍数量少于入参。
func (l *BatchGetVideoLogic) BatchGetVideo(in *video.BatchGetVideoReq) (*video.BatchGetVideoResp, error) {
	ids := make([]uint64, 0, len(in.Ids))
	for _, id := range in.Ids {
		if id > 0 {
			ids = append(ids, uint64(id))
		}
	}
	if len(ids) == 0 {
		return &video.BatchGetVideoResp{
			Base:  toStatus(errno.OK),
			Items: convertVideos(l.ctx, l.svcCtx.Redis, nil),
		}, nil
	}

	list, err := l.svcCtx.VideoModel.FindByIds(l.ctx, ids)
	if err != nil {
		return nil, err
	}

	return &video.BatchGetVideoResp{
		Base:  toStatus(errno.OK),
		Items: convertVideos(l.ctx, l.svcCtx.Redis, list),
	}, nil
}
