package storage

import (
	"errors"
	"fmt"
	"io"
)

// ErrNotConfigured 存储未配置（凭据缺失）时上传返回的错误
var ErrNotConfigured = errors.New("对象存储未配置")

// LazyOSS 惰性初始化的 OSS 客户端：
// 凭据缺失时构造不报错（避免拖垮整个服务启动），真正上传时才返回错误。
// 凭据补齐后（如热加载配置）下一次上传即生效。
type LazyOSS struct {
	cfg     OSSConfig
	inner   Object // 缓存的客户端，nil 表示未创建
	invalid bool   // true 表示已尝试创建但凭据无效
}

// NewLazyOSS 创建惰性 OSS 客户端
func NewLazyOSS(cfg OSSConfig) Object {
	return &LazyOSS{cfg: cfg}
}

func (l *LazyOSS) configured() bool {
	return l.cfg.Endpoint != "" && l.cfg.AccessKeyID != "" &&
		l.cfg.AccessKeySecret != "" && l.cfg.BucketName != ""
}

// Upload 上传对象并返回其访问 URL
func (l *LazyOSS) Upload(key string, reader io.Reader, size int64, contentType string) (string, error) {
	if !l.configured() {
		return "", ErrNotConfigured
	}
	if l.inner == nil && !l.invalid {
		client, err := NewOSS(l.cfg)
		if err != nil {
			l.invalid = true // 凭据无效不重试，直到配置变更
			return "", fmt.Errorf("初始化 oss client: %w", err)
		}
		l.inner = client
	}
	if l.inner == nil {
		return "", ErrNotConfigured
	}
	return l.inner.Upload(key, reader, size, contentType)
}
