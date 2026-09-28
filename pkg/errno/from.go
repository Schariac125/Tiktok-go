package errno

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

// FromErr 从 error 提取业务错误。
// 若 err 为 nil 返回 OK；若 err 是 *Errno 原样返回；
// 否则视为未知错误，记录原始错误后返回 ErrInternal。
func FromErr(ctx context.Context, err error) *Errno {
	if err == nil {
		return OK
	}
	if e, ok := errors.AsType[*Errno](err); ok {
		return e
	}
	// 未知错误（DB 故障、RPC 超时等）必须留痕，否则线上只剩一句"服务器内部错误"
	logx.WithContext(ctx).Errorf("unexpected error: %v", err)
	return ErrInternal
}
