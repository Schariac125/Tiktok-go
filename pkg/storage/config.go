package storage

// OSSConfig 阿里云 OSS 配置。
// 字段名与 go-zero yaml 配置一一对应（小写蛇形）。
type OSSConfig struct {
	Endpoint        string // OSS 区域节点，如 oss-cn-beijing.aliyuncs.com
	AccessKeyID     string // 阿里云 AccessKey ID
	AccessKeySecret string // 阿里云 AccessKey Secret
	BucketName      string // 存储桶名
	Domain          string // 自定义域名或 bucket 公网访问域名（用于拼接对象 URL），如 https://bucket.oss-cn-beijing.aliyuncs.com
}
