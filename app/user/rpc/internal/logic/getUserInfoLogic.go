package logic

import (
	"context"
	"errors"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户信息，只允许用户获取自己的 GET /user/info
func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	// 参数校验：user_id 由网关从登录态注入，必填
	if in.UserId <= 0 {
		return &user.GetUserInfoResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if errors.Is(err, model.ErrNotFound) {
		return &user.GetUserInfoResp{Base: toStatus(errno.ErrUserNotFound)}, nil
	}
	if err != nil {
		return nil, err
	}

	return &user.GetUserInfoResp{
		Base: toStatus(errno.OK),
		Data: convertUser(u),
	}, nil
}
