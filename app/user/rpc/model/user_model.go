package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		UpdateAvatarUrl(ctx context.Context, id uint64, avatarUrl string) error
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}

// UpdateAvatarUrl 只更新头像字段并清除相关缓存
func (m *customUserModel) UpdateAvatarUrl(ctx context.Context, id uint64, avatarUrl string) error {
	userKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set `avatar_url` = ? where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, avatarUrl, id)
	}, userKey)
	return err
}
