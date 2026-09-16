package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 上传头像，头像应该要采用云端存储 PUT /user/avatar/upload
func (l *UploadAvatarLogic) UploadAvatar(in *user.UploadAvatarReq) (*user.UploadAvatarResp, error) {
	if in.UserId <= 0 {
		return &user.UploadAvatarResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}
	if len(in.Avatar) == 0 {
		return &user.UploadAvatarResp{Base: toStatus(errno.ErrEmptyFile)}, nil
	}
	if len(in.Avatar) > constant.AvatarMaxSize {
		return &user.UploadAvatarResp{Base: toStatus(errno.ErrFileTooLarge)}, nil
	}

	// 从文件名后缀判断类型（网关侧 multipart 传入）
	ext := strings.ToLower(filepath.Ext(in.Filename))
	contentType, ok := constant.AllowedImageExts[ext]
	if !ok {
		return &user.UploadAvatarResp{Base: toStatus(errno.ErrFileType)}, nil
	}

	// 上传到 OSS：avatar/<userId>/<毫秒时间戳><后缀>
	key := fmt.Sprintf("avatar/%d/%d%s", in.UserId, time.Now().UnixMilli(), ext)
	url, err := l.svcCtx.OSS.Upload(key, bytes.NewReader(in.Avatar), int64(len(in.Avatar)), contentType)
	if err != nil {
		l.Errorf("上传头像到 oss 失败: %v", err)
		return &user.UploadAvatarResp{Base: toStatus(errno.ErrUploadFailed)}, nil
	}

	// 更新用户头像字段
	if err := l.svcCtx.UserModel.UpdateAvatarUrl(l.ctx, uint64(in.UserId), url); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &user.UploadAvatarResp{Base: toStatus(errno.ErrUserNotFound)}, nil
		}
		return nil, err
	}

	return &user.UploadAvatarResp{Base: toStatus(errno.OK)}, nil
}
