// Package constant 集中存放业务上的各项限制值。
// 同类限制统一在此定义，避免散落在各个 logic 里各写一份、改一处漏一处。
package constant

// 上传文件的大小上限
const (
	AvatarMaxSize = 5 << 20   // 头像，5MB
	CoverMaxSize  = 5 << 20   // 视频封面，5MB
	VideoMaxSize  = 500 << 20 // 视频，500MB
)

// AllowedImageExts 允许上传的图片格式及其 MIME，头像与视频封面共用
var AllowedImageExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// AllowedVideoExts 允许上传的视频格式及其 MIME
var AllowedVideoExts = map[string]string{
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".webm": "video/webm",
	".flv":  "video/x-flv",
}
