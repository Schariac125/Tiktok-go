// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"
	"io"
	"mime/multipart"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 视频投稿 POST /video/publish
// cover、coverHeader 可为 nil，表示本次投稿没有上传封面
func (l *PublishVideoLogic) PublishVideo(data multipart.File, dataHeader *multipart.FileHeader,
	cover multipart.File, coverHeader *multipart.FileHeader, title, description string,
) (*types.PublishVideoResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.PublishVideoResp{Base: types.Status{Code: errno.ErrUnauthorized.Code, Msg: errno.ErrUnauthorized.Msg}}, nil
	}

	content, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	// 封面可选，未上传时传空
	var coverContent []byte
	var coverFilename string
	if cover != nil && coverHeader != nil {
		if coverContent, err = io.ReadAll(cover); err != nil {
			return nil, err
		}
		coverFilename = coverHeader.Filename
	}

	rpcResp, err := l.svcCtx.VideoRpc.PublishVideo(l.ctx, &video.PublishVideoReq{
		UserId:        userId,
		Data:          content,
		DataFilename:  dataHeader.Filename,
		Cover:         coverContent,
		CoverFilename: coverFilename,
		Title:         title,
		Description:   description,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.PublishVideoResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}
	return &types.PublishVideoResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
