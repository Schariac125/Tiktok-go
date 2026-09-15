package logic

import (
	"database/sql"
	"time"

	"tiktok-go/app/user/rpc/model"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"
)

// toStatus 将业务错误转换为 proto 的 Status 响应
func toStatus(e *errno.Errno) *user.Status {
	return &user.Status{Code: e.Code, Msg: e.Msg}
}

// convertUser 将 model.User 转换为 proto 的 User；
// 时间字段按接口文档约定输出字符串，空值（NULL）输出空串
func convertUser(u *model.User) *user.User {
	return &user.User{
		Id:        int64(u.Id),
		Username:  u.Username,
		Password:  "",
		AvatarUrl: u.AvatarUrl,
		CreateAt:  formatTime(u.CreateAt),
		UpdatedAt: formatTime(u.UpdatedAt),
		DeleteAt:  formatNullTime(u.DeleteAt),
	}
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05")
}

func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return formatTime(t.Time)
}
