package logic

import (
	"context"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVideoLogic {
	return &ListVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看用户视频 GET /video/list
// 只返回未软删除的视频，按投稿时间倒序分页；用户不存在时返回空列表而非报错，
// 因为 video 服务不掌握用户数据，为此去调 user 服务不划算。
func (l *ListVideoLogic) ListVideo(in *video.ListVideoReq) (*video.ListVideoResp, error) {
	if in.UserId <= 0 {
		return &video.ListVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("用户 id 不能为空"))}, nil
	}

	userId := uint64(in.UserId)
	total, err := l.svcCtx.VideoModel.CountByUserId(l.ctx, userId)
	if err != nil {
		return nil, err
	}

	// 没有投稿时不必再查一次列表
	if total == 0 {
		return &video.ListVideoResp{
			Base: toStatus(errno.OK),
			Data: &video.VideoPageData{Items: convertVideos(l.ctx, l.svcCtx.Redis, nil)},
		}, nil
	}

	pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
	offset := (pageNum - 1) * pageSize

	// 页码超出范围时返回空列表，不报错
	list, err := l.svcCtx.VideoModel.FindByUserId(l.ctx, userId, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &video.ListVideoResp{
		Base: toStatus(errno.OK),
		Data: &video.VideoPageData{
			Items: convertVideos(l.ctx, l.svcCtx.Redis, list),
			Total: total,
		},
	}, nil
}
