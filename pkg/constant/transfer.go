package constant

// GRPCMsgMaxSize 是 gRPC 单条消息的上限。
// gRPC 默认只收 4MB，而上传接口（如头像）要经 gRPC 传文件内容，需按业务上限放开。
const GRPCMsgMaxSize = 8 << 20 // 8MB

// MultipartMaxMemory 是 multipart 解析时驻留内存的阈值，
// 超出部分由标准库落到临时文件，避免大文件整段占内存。
const MultipartMaxMemory = 32 << 20 // 32MB
