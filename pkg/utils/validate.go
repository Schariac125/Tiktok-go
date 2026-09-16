package utils

import "tiktok-go/pkg/constant"

// ValidUsername 校验用户名长度是否合法
func ValidUsername(username string) bool {
	n := len([]rune(username))
	return n >= constant.UsernameMinLen && n <= constant.UsernameMaxLen
}

// ValidPassword 校验密码长度是否合法
func ValidPassword(password string) bool {
	n := len([]rune(password))
	return n >= constant.PasswordMinLen && n <= constant.PasswordMaxLen
}
