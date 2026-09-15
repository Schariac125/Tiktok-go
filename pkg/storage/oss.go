package storage

import (
	"fmt"
	"io"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 确保 ossClient 实现 Object 接口
var _ Object = (*ossClient)(nil)

type ossClient struct {
	client *oss.Client
	bucket *oss.Bucket
	cfg    OSSConfig
}

// NewOSS 创建阿里云 OSS 客户端
func NewOSS(cfg OSSConfig) (Object, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 oss client: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取 bucket %s: %w", cfg.BucketName, err)
	}
	return &ossClient{client: client, bucket: bucket, cfg: cfg}, nil
}

// Upload 上传对象并返回其访问 URL
func (c *ossClient) Upload(key string, reader io.Reader, size int64, contentType string) (string, error) {
	if reader == nil {
		return "", ErrEmptyContent
	}

	options := []oss.Option{oss.ContentType(contentType)}
	if size > 0 {
		options = append(options, oss.ContentLength(size))
	}
	if err := c.bucket.PutObject(key, reader, options...); err != nil {
		return "", fmt.Errorf("上传对象 %s: %w", key, err)
	}

	// 拼接对象 URL：Domain 为空时用默认公网域名
	domain := c.cfg.Domain
	if domain == "" {
		domain = "https://" + c.cfg.BucketName + "." + c.cfg.Endpoint
	}
	return fmt.Sprintf("%s/%s", domain, url.PathEscape(key)), nil
}
