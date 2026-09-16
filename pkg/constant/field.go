package constant

// 文本字段的长度上限，单位均为字符数（rune）
const (
	// 用户名、密码
	UsernameMinLen = 2
	UsernameMaxLen = 64
	PasswordMinLen = 6
	PasswordMaxLen = 64

	// 视频标题、描述，与 migration/video.sql 的列宽一致
	TitleMaxLen       = 255
	DescriptionMaxLen = 1024
)
