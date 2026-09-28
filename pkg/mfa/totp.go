// Package mfa 提供 TOTP（基于时间的一次性密码）工具，
// 兼容 Google Authenticator 等验证器 App。
//
// 底层采用 github.com/pquerna/otp，参数取该库默认值，与 RFC 6238
// 及各验证器 App 的默认值一致：HMAC-SHA1、6 位、30 秒周期、160 位密钥。
package mfa

import (
	"github.com/pquerna/otp/totp"
)

// issuer 验证器 App 中展示的签发方
const issuer = "tiktok-go"

// NewKey 生成新的 TOTP 密钥及其 otpauth URL。
// 两者描述同一个密钥，故一并返回，避免调用方自行拼接导致二者不一致。
// secret 不落库，待绑定校验通过后才提交。
func NewKey(account string) (secret, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// Validate 校验验证码，容许前后各一个时间窗的时钟偏差。
// 注意参数顺序与 pquerna/otp 相反：本包以 (secret, code) 为约定。
func Validate(secret, code string) bool {
	return totp.Validate(code, secret)
}
