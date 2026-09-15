package logic

import (
	"context"
	"errors"
	"strings"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/mfa"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录 POST /user/login
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	username := strings.TrimSpace(in.Username)
	password := strings.TrimSpace(in.Password)

	// 查找用户
	u, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, username)
	if errors.Is(err, model.ErrNotFound) {
		return &user.LoginResp{Base: toStatus(errno.ErrUserNotFound)}, nil
	}
	if err != nil {
		return nil, err
	}

	// 校验密码
	if !utils.CheckPassword(password, u.Password) {
		return &user.LoginResp{Base: toStatus(errno.ErrPasswordWrong)}, nil
	}

	// 已启用 MFA 的用户必须携带验证码并校验 TOTP
	if u.MfaEnabled == 1 {
		if in.Code == "" {
			return &user.LoginResp{Base: toStatus(errno.ErrMfaCodeNeeded)}, nil
		}
		if !u.MfaSecret.Valid || !mfa.Validate(u.MfaSecret.String, in.Code) {
			return &user.LoginResp{Base: toStatus(errno.ErrMfaCodeWrong)}, nil
		}
	}

	// 签发 token 对
	pair, err := l.svcCtx.Signer.SignPair(int64(u.Id), u.Username)
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Base:         toStatus(errno.OK),
		Data:         convertUser(u),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
