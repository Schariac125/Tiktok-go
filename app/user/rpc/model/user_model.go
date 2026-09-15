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
		UpdateMFA(ctx context.Context, id uint64, secret string, enabled bool) error
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

// UpdateMFA 更新 MFA 绑定状态：secret 传空串且 enabled=false 表示解绑
func (m *customUserModel) UpdateMFA(ctx context.Context, id uint64, secret string, enabled bool) error {
	userKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
	enabledVal := 0
	if enabled {
		enabledVal = 1
	}
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set `mfa_secret` = ?, `mfa_enabled` = ? where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, secret, enabledVal, id)
	}, userKey)
	return err
}
