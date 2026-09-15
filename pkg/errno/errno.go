// Package errno 统一定义业务错误码与错误信息。
//
// 约定：
//   - Code 0 表示成功
//   - 4xx 表示客户端错误（参数、鉴权、冲突等）
//   - 5xx 表示服务端错误
package errno

// Errno 业务错误
type Errno struct {
	Code int32  `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
}

// Error 实现 error 接口
func (e *Errno) Error() string {
	return e.Msg
}

// WithMsg 基于已有 Errno 替换提示信息（如附带具体字段名）
func (e *Errno) WithMsg(msg string) *Errno {
	return &Errno{Code: e.Code, Msg: msg}
}

// Status 返回 code/msg 二元组，供各层构造响应体的 base 字段使用
func (e *Errno) Status() (int32, string) {
	return e.Code, e.Msg
}

// 通用
var (
	OK = &Errno{Code: 0, Msg: "成功"}
)

// 客户端错误 4xx
var (
	ErrInvalidParam  = &Errno{Code: 400, Msg: "参数错误"}
	ErrUsernameLen   = &Errno{Code: 400, Msg: "用户名长度需为 2-64 个字符"}
	ErrPasswordLen   = &Errno{Code: 400, Msg: "密码长度需为 6-64 个字符"}
	ErrUsernameExist = &Errno{Code: 400, Msg: "用户名已存在"}
	ErrUserNotFound  = &Errno{Code: 404, Msg: "用户不存在"}
	ErrPasswordWrong = &Errno{Code: 401, Msg: "用户名或密码错误"}
	ErrUnauthorized  = &Errno{Code: 401, Msg: "未登录或凭证已失效"}
	ErrMfaCodeNeeded = &Errno{Code: 401, Msg: "需要 MFA 验证码"}
	ErrMfaCodeWrong  = &Errno{Code: 401, Msg: "MFA 验证码错误"}
)

// 服务端错误 5xx
var (
	ErrInternal = &Errno{Code: 500, Msg: "服务器内部错误"}
)
