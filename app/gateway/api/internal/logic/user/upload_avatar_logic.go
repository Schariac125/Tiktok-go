// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package user

import (
	"context"
	"io"
	"mime/multipart"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 上传头像 PUT /user/avatar/upload
func (l *UploadAvatarLogic) UploadAvatar(file multipart.File, header *multipart.FileHeader) (*types.UploadAvatarResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.UploadAvatarResp{Base: types.Status{Code: errno.ErrUnauthorized.Code, Msg: errno.ErrUnauthorized.Msg}}, nil
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.UserRpc.UploadAvatar(l.ctx, &user.UploadAvatarReq{
		UserId:   userId,
		Avatar:   content,
		Filename: header.Filename,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.UploadAvatarResp{Base: types.Status{Code: errno.ErrInternal.Code, Msg: errno.ErrInternal.Msg}}, nil
	}
	return &types.UploadAvatarResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
