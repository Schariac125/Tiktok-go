package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VideoModel = (*customVideoModel)(nil)

type (
	// VideoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVideoModel.
	VideoModel interface {
		videoModel
		// SoftDelete 软删除：把 delete_at 置为当前时间，记录保留在表里。
		// 记录不存在或已被删除时返回 ErrNotFound。
		SoftDelete(ctx context.Context, id uint64) error
	}

	customVideoModel struct {
		*defaultVideoModel
	}
)

// NewVideoModel returns a model for the database table.
func NewVideoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VideoModel {
	return &customVideoModel{
		defaultVideoModel: newVideoModel(conn, c, opts...),
	}
}

// SoftDelete 软删除，只置 delete_at，不做物理删除。
// delete_at is null 的条件保证重复删除不会覆盖首次删除时间。
func (m *customVideoModel) SoftDelete(ctx context.Context, id uint64) error {
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set `delete_at` = now() where `id` = ? and `delete_at` is null", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, fmt.Sprintf("%s%v", cacheVideoIdPrefix, id))
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
