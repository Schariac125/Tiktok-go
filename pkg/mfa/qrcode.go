package mfa

import (
	"encoding/base64"

	qrcode "github.com/skip2/go-qrcode"
)

// QrcodeBase64 生成 otpauth URL 的二维码 PNG，返回 base64 编码字符串
// （接口文档约定：base64 编码后的 qrcode）
func QrcodeBase64(otpauthURL string) (string, error) {
	png, err := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(png), nil
}
