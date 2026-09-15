package logic

import (
	"context"
	"fmt"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/mfa"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMFAQrcodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMFAQrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMFAQrcodeLogic {
	return &GetMFAQrcodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取绑定 MFA 时所需的二维码 GET /auth/mfa/qrcode
func (l *GetMFAQrcodeLogic) GetMFAQrcode(in *user.GetMFAQrcodeReq) (*user.GetMFAQrcodeResp, error) {
	if in.UserId <= 0 {
		return &user.GetMFAQrcodeResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	// 生成密钥与扫码二维码；secret 不落库，绑定验证通过时才提交
	secret := mfa.GenerateSecret()
	qrcode, err := mfa.QrcodeBase64(mfa.OtpauthURL(secret, l.account(in.UserId)))
	if err != nil {
		return nil, err
	}

	return &user.GetMFAQrcodeResp{
		Base:    toStatus(errno.OK),
		Secret:  secret,
		Qrcode:  qrcode,
	}, nil
}

// account 生成 otpauth 中的账号标识（用用户 id，避免泄露用户名）
func (l *GetMFAQrcodeLogic) account(userId int64) string {
	return "user-" + fmt.Sprint(userId)
}
