package utils

// 用户名/密码长度限制
const (
	UsernameMinLen = 2
	UsernameMaxLen = 64
	PasswordMinLen = 6
	PasswordMaxLen = 64
)

// ValidUsername 校验用户名长度是否合法（2-64 个字符）
func ValidUsername(username string) bool {
	n := len([]rune(username))
	return n >= UsernameMinLen && n <= UsernameMaxLen
}

// ValidPassword 校验密码长度是否合法（6-64 个字符）
func ValidPassword(password string) bool {
	n := len([]rune(password))
	return n >= PasswordMinLen && n <= PasswordMaxLen
}
