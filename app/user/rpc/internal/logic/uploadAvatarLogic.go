package logic

import (
	"context"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"

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
	// todo: add your logic here and delete this line

	return &user.UploadAvatarResp{}, nil
}
