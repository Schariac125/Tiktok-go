package logic

import (
	"context"
	"strings"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

// 字段长度限制，与 migration/video.sql 的列宽一致
const (
	titleMaxLen       = 255
	descriptionMaxLen = 1024
)

type PublishVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 视频投稿 POST /video/publish
// 视频与封面已由网关上传到 OSS，这里只做落库前的校验和写入
func (l *PublishVideoLogic) PublishVideo(in *video.PublishVideoReq) (*video.PublishVideoResp, error) {
	if in.UserId <= 0 {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}
	if strings.TrimSpace(in.VideoUrl) == "" {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频地址不能为空"))}, nil
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频标题不能为空"))}, nil
	}
	if len([]rune(title)) > titleMaxLen {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频标题过长"))}, nil
	}
	description := strings.TrimSpace(in.Description)
	if len([]rune(description)) > descriptionMaxLen {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频描述过长"))}, nil
	}

	// 投稿成功后观看量、点赞量、评论数均为 0
	if _, err := l.svcCtx.VideoModel.Insert(l.ctx, &model.Video{
		UserId:      uint64(in.UserId),
		VideoUrl:    in.VideoUrl,
		CoverUrl:    in.CoverUrl,
		Title:       title,
		Description: description,
	}); err != nil {
		l.Errorf("保存视频记录失败: %v", err)
		return nil, err
	}

	return &video.PublishVideoResp{Base: toStatus(errno.OK)}, nil
}
