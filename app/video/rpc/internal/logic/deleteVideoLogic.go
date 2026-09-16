package logic

import (
	"context"
	"errors"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除视频 DELETE /video/delete
// 软删除：只把 delete_at 置为当前时间，记录仍保留在表里，且只允许作者删自己的视频
func (l *DeleteVideoLogic) DeleteVideo(in *video.DeleteVideoReq) (*video.DeleteVideoResp, error) {
	if in.UserId <= 0 || in.Id <= 0 {
		return &video.DeleteVideoResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	// 先取出记录判断归属，软删除本身也带条件，并发下只会成功一次
	v, err := l.svcCtx.VideoModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &video.DeleteVideoResp{Base: toStatus(errno.ErrVideoNotFound)}, nil
		}
		return nil, err
	}
	// 已软删除的记录仍在表里，对外一律当作不存在
	if v.DeleteAt.Valid {
		return &video.DeleteVideoResp{Base: toStatus(errno.ErrVideoNotFound)}, nil
	}
	if v.UserId != uint64(in.UserId) {
		return &video.DeleteVideoResp{Base: toStatus(errno.ErrVideoForbidden)}, nil
	}

	if err := l.svcCtx.VideoModel.SoftDelete(l.ctx, uint64(in.Id)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 并发删除时后到的请求会走到这里
			return &video.DeleteVideoResp{Base: toStatus(errno.ErrVideoNotFound)}, nil
		}
		return nil, err
	}

	return &video.DeleteVideoResp{Base: toStatus(errno.OK)}, nil
}
