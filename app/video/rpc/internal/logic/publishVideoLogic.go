package logic

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

// 投稿文件的大小限制
const (
	videoMaxSize = 100 << 20 // 视频 100MB
	coverMaxSize = 5 << 20   // 封面 5MB
)

// 字段长度限制，与 migration/video.sql 的列宽一致
const (
	titleMaxLen       = 255
	descriptionMaxLen = 1024
)

// 允许的视频格式
var allowedVideoExts = map[string]string{
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".webm": "video/webm",
	".flv":  "video/x-flv",
}

// 允许的封面格式
var allowedCoverExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

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
func (l *PublishVideoLogic) PublishVideo(in *video.PublishVideoReq) (*video.PublishVideoResp, error) {
	if in.UserId <= 0 {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	// 视频文件
	if len(in.Data) == 0 {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrEmptyFile.WithMsg("视频文件不能为空"))}, nil
	}
	if len(in.Data) > videoMaxSize {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrFileTooLarge.WithMsg("视频大小不能超过 100MB"))}, nil
	}
	videoExt := strings.ToLower(filepath.Ext(in.DataFilename))
	videoContentType, ok := allowedVideoExts[videoExt]
	if !ok {
		return &video.PublishVideoResp{Base: toStatus(errno.ErrFileType.WithMsg("不支持的视频格式"))}, nil
	}

	// 标题、描述
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

	// 封面（可选），先传小文件，校验不通过时不会留下已上传的视频
	var coverUrl string
	if len(in.Cover) > 0 {
		if len(in.Cover) > coverMaxSize {
			return &video.PublishVideoResp{Base: toStatus(errno.ErrFileTooLarge.WithMsg("封面大小不能超过 5MB"))}, nil
		}
		coverExt := strings.ToLower(filepath.Ext(in.CoverFilename))
		coverContentType, ok := allowedCoverExts[coverExt]
		if !ok {
			return &video.PublishVideoResp{Base: toStatus(errno.ErrFileType.WithMsg("不支持的封面格式"))}, nil
		}

		key := fmt.Sprintf("cover/%d/%d%s", in.UserId, time.Now().UnixMilli(), coverExt)
		url, err := l.svcCtx.OSS.Upload(key, bytes.NewReader(in.Cover), int64(len(in.Cover)), coverContentType)
		if err != nil {
			l.Errorf("上传封面到 oss 失败: %v", err)
			return &video.PublishVideoResp{Base: toStatus(errno.ErrUploadFailed)}, nil
		}
		coverUrl = url
	}

	// 视频上传到 OSS：video/<userId>/<毫秒时间戳><后缀>
	key := fmt.Sprintf("video/%d/%d%s", in.UserId, time.Now().UnixMilli(), videoExt)
	videoUrl, err := l.svcCtx.OSS.Upload(key, bytes.NewReader(in.Data), int64(len(in.Data)), videoContentType)
	if err != nil {
		l.Errorf("上传视频到 oss 失败: %v", err)
		return &video.PublishVideoResp{Base: toStatus(errno.ErrUploadFailed)}, nil
	}

	// 落库，投稿成功后观看量、点赞量、评论数均为 0
	if _, err := l.svcCtx.VideoModel.Insert(l.ctx, &model.Video{
		UserId:      uint64(in.UserId),
		VideoUrl:    videoUrl,
		CoverUrl:    coverUrl,
		Title:       title,
		Description: description,
	}); err != nil {
		l.Errorf("保存视频记录失败: %v", err)
		return nil, err
	}

	return &video.PublishVideoResp{Base: toStatus(errno.OK)}, nil
}
