package mfa

import (
	"strings"
	"testing"
	"time"
)

// 时间窗时长
var periodDuration = time.Duration(period) * time.Second

// RFC 6238 语义：同一密钥同一时间窗生成相同验证码，不同窗不同
func TestTOTPConsistent(t *testing.T) {
	secret := GenerateSecret()
	if len(secret) == 0 {
		t.Fatal("secret 为空")
	}

	now := time.Now()
	code := TOTP(secret, now)
	if len(code) != digits {
		t.Fatalf("验证码长度: %s", code)
	}
	if TOTP(secret, now.Add(2*time.Second)) != code {
		t.Fatal("同一窗内的验证码应一致")
	}

	// 相邻时间窗的验证码大概率不同（1/1000000 碰撞概率，跨 3 个窗观察）
	changed := false
	for i := 1; i <= 3; i++ {
		if TOTP(secret, now.Add(periodDuration*time.Duration(i))) != code {
			changed = true
		}
	}
	if !changed {
		t.Fatal("不同时间窗的验证码应不同")
	}
}

// Validate 接受当前窗与相邻窗（skew=1），拒绝两窗之前
func TestValidate(t *testing.T) {
	secret := GenerateSecret()
	now := time.Now()

	// 以当前真实时间为基准生成，Validate 才能通过
	if !Validate(secret, TOTP(secret, now)) {
		t.Fatal("当前窗验证码应通过")
	}

	// 前后相邻窗（skew=1 容忍）
	prev := TOTP(secret, now.Add(-periodDuration))
	if !Validate(secret, prev) {
		t.Fatal("前一窗验证码应通过（skew=1）")
	}
	next := TOTP(secret, now.Add(periodDuration))
	if !Validate(secret, next) {
		t.Fatal("后一窗验证码应通过（skew=1）")
	}

	// 两窗之前：若碰巧等于有效码则跳过（概率 1/1000000）
	old := TOTP(secret, now.Add(-3*periodDuration))
	if old == prev || old == TOTP(secret, now) {
		t.Skip("碰撞，跳过")
	}
	if Validate(secret, old) {
		t.Fatal("两窗之前的验证码应被拒绝")
	}

	if Validate(secret, "12345") || Validate(secret, "abcdef") || Validate(secret, "") {
		t.Fatal("格式非法的验证码应被拒绝")
	}
}

// 小写的 secret 应被容忍
func TestDecodeTolerant(t *testing.T) {
	secret := GenerateSecret()
	code := TOTP(secret, time.Now())
	if !Validate(strings.ToLower(secret), code) {
		t.Fatal("小写 secret 应被容忍")
	}
}

// otpauth URL 格式
func TestOtpauthURL(t *testing.T) {
	url := OtpauthURL("SECRET", "user-1")
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Fatalf("url 前缀: %s", url)
	}
	if !strings.Contains(url, "secret=SECRET") {
		t.Fatalf("url 缺少 secret: %s", url)
	}
}
