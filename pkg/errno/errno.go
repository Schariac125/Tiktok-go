package errno

// Errno 业务错误
type Errno struct {
	Code int32  `json:"code"` // 业务码
	Msg  string `json:"msg"`  // 提示信息
}

// Error 实现 error 接口
func (e *Errno) Error() string {
	return e.Msg
}

// Is 让 errors.Is 按业务码判定。
// WithMsg 产生的副本、以及与业务码一并被包装过的错误，都能命中本判定。
func (e *Errno) Is(target error) bool {
	t, ok := target.(*Errno)
	return ok && t.Code == e.Code
}

// WithMsg 基于已有 Errno 替换提示信息（如附带具体字段名）
func (e *Errno) WithMsg(msg string) *Errno {
	return &Errno{Code: e.Code, Msg: msg}
}

// 通用
var (
	OK = &Errno{Code: 0, Msg: "成功"}
)

// 通用 1xxxx
var (
	ErrInvalidParam = &Errno{Code: 10001, Msg: "参数错误"}
)

// 用户 2xxxx
var (
	ErrUsernameLen   = &Errno{Code: 20001, Msg: "用户名长度需为 2-64 个字符"}
	ErrPasswordLen   = &Errno{Code: 20002, Msg: "密码长度需为 6-64 个字符"}
	ErrUsernameExist = &Errno{Code: 20003, Msg: "用户名已存在"}
	ErrUserNotFound  = &Errno{Code: 20004, Msg: "用户不存在"}
	ErrPasswordWrong = &Errno{Code: 20005, Msg: "用户名或密码错误"}
	// ErrUnauthorized 前端据此判定跳转登录页
	ErrUnauthorized  = &Errno{Code: 20006, Msg: "未登录或凭证已失效"}
	ErrMfaCodeNeeded = &Errno{Code: 20007, Msg: "需要 MFA 验证码"}
	ErrMfaCodeWrong  = &Errno{Code: 20008, Msg: "MFA 验证码错误"}
	ErrMfaBound      = &Errno{Code: 20009, Msg: "MFA 已绑定，请勿重复绑定"}
	ErrEmptyFile     = &Errno{Code: 20010, Msg: "上传文件为空"}
	ErrFileTooLarge  = &Errno{Code: 20011, Msg: "文件大小超出限制"}
	ErrFileType      = &Errno{Code: 20012, Msg: "不支持的文件类型"}
)

// 视频 3xxxx
var (
	ErrVideoNotFound = &Errno{Code: 30001, Msg: "视频不存在"}
	// ErrVideoForbidden 操作的不是自己投稿的视频
	ErrVideoForbidden = &Errno{Code: 30002, Msg: "无权操作该视频"}
)

// 互动 5xxxx
var (
	ErrCommentNotFound = &Errno{Code: 50001, Msg: "评论不存在"}
	// ErrCommentForbidden 操作的不是自己发表的评论
	ErrCommentForbidden = &Errno{Code: 50002, Msg: "无权操作该评论"}
)

// 服务端 9xxxx
var (
	ErrInternal     = &Errno{Code: 90001, Msg: "服务器内部错误"}
	ErrUploadFailed = &Errno{Code: 90002, Msg: "上传失败"}
)
