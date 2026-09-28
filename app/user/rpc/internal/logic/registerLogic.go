package logic

import (
	"context"
	"errors"
	"strings"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册 POST /user/register
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	username := strings.TrimSpace(in.Username)
	password := strings.TrimSpace(in.Password)

	// 参数校验
	if !utils.ValidUsername(username) {
		return &user.RegisterResp{Base: toStatus(errno.ErrUsernameLen)}, nil
	}
	if !utils.ValidPassword(password) {
		return &user.RegisterResp{Base: toStatus(errno.ErrPasswordLen)}, nil
	}

	// 用户名查重（软删除的记录同样占用用户名，防止复活）
	if _, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username); err == nil {
		return &user.RegisterResp{Base: toStatus(errno.ErrUsernameExist)}, nil
	} else if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	// 密码加密后落库
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username: username,
		Password: hashed,
	}); err != nil {
		// 并发下撞唯一索引同样视为用户名已存在
		return &user.RegisterResp{Base: toStatus(errno.ErrUsernameExist)}, nil
	}

	return &user.RegisterResp{Base: toStatus(errno.OK)}, nil
}
