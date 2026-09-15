package errno

import "errors"

// FromErr 从 error 提取业务错误。
// 若 err 为 nil 返回 OK；若 err 是 *Errno 原样返回；否则视为未知错误返回 ErrInternal。
func FromErr(err error) *Errno {
	if err == nil {
		return OK
	}
	if e, ok := errors.AsType[*Errno](err); ok {
		return e
	}
	return ErrInternal
}
