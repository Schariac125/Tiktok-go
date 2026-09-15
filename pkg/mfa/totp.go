// Package mfa 提供 TOTP（基于时间的一次性密码）工具，
// 兼容 Google Authenticator 等验证器 App。
package mfa

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	period  = 30 // 时间步长（秒），验证器 App 默认值
	skew    = 1  // 允许的前后时间窗偏移（±1 个周期，容忍时钟误差）
	digits  = 6  // 验证码位数
	issuer  = "tiktok-go"
	base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567" // 无填充标准字母表
)

// GenerateSecret 生成 160 位随机 TOTP 密钥（base32 编码，无填充）
func GenerateSecret() string {
	b := make([]byte, 20)
	for i := range b {
		b[i] = randomByte()
	}
	encoder := base32.NewEncoding(base32Alphabet).WithPadding(base32.NoPadding)
	return encoder.EncodeToString(b)
}

// TOTP 计算给定时刻的验证码
func TOTP(secret string, t time.Time) string {
	key, err := decodeSecret(secret)
	if err != nil {
		return ""
	}
	counter := uint64(t.Unix()) / period
	return hotp(key, counter)
}

// Validate 校验验证码（允许 ±1 个时间窗的时钟偏差）
func Validate(secret, code string) bool {
	if len(code) != digits {
		return false
	}
	key, err := decodeSecret(secret)
	if err != nil {
		return false
	}
	now := time.Now()
	counter := uint64(now.Unix()) / period
	for i := -skew; i <= skew; i++ {
		if hotp(key, counter+uint64(i)) == code {
			return true
		}
	}
	return false
}

// OtpauthURL 生成验证器 App 扫码所需的 otpauth:// URL
func OtpauthURL(secret, account string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s&period=%d&digits=%d",
		url.PathEscape(issuer), url.PathEscape(account), url.PathEscape(issuer), secret, period, digits)
}

// hotp 按 RFC 4226 计算 HMAC-based 一次性密码
func hotp(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff) % 1000000
	return fmt.Sprintf("%06d", code)
}

// decodeSecret base32 解码密钥（容忍小写与填充符）
func decodeSecret(secret string) ([]byte, error) {
	s := strings.ToUpper(strings.TrimRight(strings.TrimSpace(secret), "="))
	return base32.NewEncoding(base32Alphabet).WithPadding(base32.NoPadding).DecodeString(s)
}
