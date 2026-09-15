package logic

import (
	"context"
	"errors"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/mfa"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMFALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindMFALogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMFALogic {
	return &BindMFALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 绑定 MFA POST /auth/mfa/bind
func (l *BindMFALogic) BindMFA(in *user.BindMFAReq) (*user.BindMFAResp, error) {
	if in.UserId <= 0 || in.Secret == "" || in.Code == "" {
		return &user.BindMFAResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	// 校验用户存在
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if errors.Is(err, model.ErrNotFound) {
		return &user.BindMFAResp{Base: toStatus(errno.ErrUserNotFound)}, nil
	}
	if err != nil {
		return nil, err
	}
	// 已绑定的不允许重复绑定（解绑接口待做，目前只能绑一次）
	if u.MfaEnabled == 1 {
		return &user.BindMFAResp{Base: toStatus(errno.ErrMfaBound)}, nil
	}

	// 校验 TOTP 验证码，通过后 secret 落库并启用
	if !mfa.Validate(in.Secret, in.Code) {
		return &user.BindMFAResp{Base: toStatus(errno.ErrMfaCodeWrong)}, nil
	}
	if err := l.svcCtx.UserModel.UpdateMFA(l.ctx, uint64(in.UserId), in.Secret, true); err != nil {
		return nil, err
	}

	return &user.BindMFAResp{Base: toStatus(errno.OK)}, nil
}
