package logic

import (
	"context"
	"errors"
	"strings"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByUsernameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByUsernameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByUsernameLogic {
	return &GetUserByUsernameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按用户名查用户，供其他服务把用户名解析成 user_id（如视频搜索按作者筛选）。
// 只提供查询，与登录不同，这里不做密码校验。
func (l *GetUserByUsernameLogic) GetUserByUsername(in *user.GetUserByUsernameReq) (*user.GetUserByUsernameResp, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" {
		return &user.GetUserByUsernameResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("用户名不能为空"))}, nil
	}

	u, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &user.GetUserByUsernameResp{Base: toStatus(errno.ErrUserNotFound)}, nil
		}
		return nil, err
	}

	return &user.GetUserByUsernameResp{
		Base: toStatus(errno.OK),
		Data: convertUser(u),
	}, nil
}
