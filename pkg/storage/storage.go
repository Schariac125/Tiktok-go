// Package storage 提供对象存储抽象，当前实现为阿里云 OSS。
package storage

import (
	"errors"
	"io"
)

// 通用错误
var (
	ErrEmptyContent = errors.New("上传内容为空")
)

// Object 对象存储抽象接口，供业务侧依赖（便于替换云厂商/本地存储/测试 mock）
type Object interface {
	// Upload 上传对象，返回可访问的对象 URL。
	//   - key: 对象键（含路径与文件名），如 avatar/2/xxx.png
	//   - reader: 对象内容
	//   - size: 内容长度（字节），未知传 -1
	Upload(key string, reader io.Reader, size int64, contentType string) (url string, err error)
}
