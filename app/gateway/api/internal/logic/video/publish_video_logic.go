// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
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
// 视频与封面由网关流式直传 OSS，rpc 只负责落库，文件内容不经 gRPC 传输。
// cover、coverHeader 可为 nil，表示本次投稿没有上传封面
func (l *PublishVideoLogic) PublishVideo(data multipart.File, dataHeader *multipart.FileHeader,
	cover multipart.File, coverHeader *multipart.FileHeader, title, description string,
) (*types.PublishVideoResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.PublishVideoResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	// 先校验轻量参数，避免用户白传一个大文件
	if strings.TrimSpace(title) == "" {
		return &types.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频标题不能为空"))}, nil
	}

	// 视频大小与格式，上传前先拦掉
	if dataHeader.Size > constant.VideoMaxSize {
		return &types.PublishVideoResp{Base: toStatus(errno.ErrFileTooLarge.WithMsg("视频大小不能超过 500MB"))}, nil
	}
	videoExt := strings.ToLower(filepath.Ext(dataHeader.Filename))
	videoContentType, ok := constant.AllowedVideoExts[videoExt]
	if !ok {
		return &types.PublishVideoResp{Base: toStatus(errno.ErrFileType.WithMsg("不支持的视频格式"))}, nil
	}

	// 封面校验（可选）
	var coverExt, coverContentType string
	if cover != nil && coverHeader != nil {
		if coverHeader.Size > constant.CoverMaxSize {
			return &types.PublishVideoResp{Base: toStatus(errno.ErrFileTooLarge.WithMsg("封面大小不能超过 5MB"))}, nil
		}
		coverExt = strings.ToLower(filepath.Ext(coverHeader.Filename))
		if coverContentType, ok = constant.AllowedImageExts[coverExt]; !ok {
			return &types.PublishVideoResp{Base: toStatus(errno.ErrFileType.WithMsg("不支持的封面格式"))}, nil
		}
	}

	// 先传小的封面，失败时不会留下已上传的视频
	var coverUrl string
	if coverContentType != "" {
		key := fmt.Sprintf("cover/%d/%d%s", userId, time.Now().UnixMilli(), coverExt)
		url, err := l.svcCtx.OSS.Upload(key, cover, coverHeader.Size, coverContentType)
		if err != nil {
			l.Errorf("上传封面到 oss 失败: %v", err)
			return &types.PublishVideoResp{Base: toStatus(errno.ErrUploadFailed)}, nil
		}
		coverUrl = url
	}

	// 视频直接以 multipart 的文件流上传，不读进内存
	key := fmt.Sprintf("video/%d/%d%s", userId, time.Now().UnixMilli(), videoExt)
	videoUrl, err := l.svcCtx.OSS.Upload(key, data, dataHeader.Size, videoContentType)
	if err != nil {
		l.Errorf("上传视频到 oss 失败: %v", err)
		return &types.PublishVideoResp{Base: toStatus(errno.ErrUploadFailed)}, nil
	}

	rpcResp, err := l.svcCtx.VideoRpc.PublishVideo(l.ctx, &video.PublishVideoReq{
		UserId:      userId,
		VideoUrl:    videoUrl,
		CoverUrl:    coverUrl,
		Title:       title,
		Description: description,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.PublishVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}
	return &types.PublishVideoResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
