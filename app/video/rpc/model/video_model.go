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
		// FindByUserId 分页查询某个用户未删除的视频，按 id 倒序。
		FindByUserId(ctx context.Context, userId uint64, offset, limit int64) ([]*Video, error)
		// CountByUserId 统计某个用户未删除的视频总数。
		CountByUserId(ctx context.Context, userId uint64) (int64, error)
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

// FindByUserId 分页查询某个用户未删除的视频，按 id 倒序（自增主键即投稿顺序）。
// select 的列取自 videoRows，与 Video 的字段一一对应，避免 go-zero 的
// strict 模式因列数少于字段数而报 not matching destination to scan。
func (m *customVideoModel) FindByUserId(ctx context.Context, userId uint64, offset, limit int64) ([]*Video, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `delete_at` is null order by `id` desc limit ? offset ?",
		videoRows, m.table)

	var list []*Video
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, userId, limit, offset); err != nil {
		return nil, err
	}
	return list, nil
}

// CountByUserId 统计某个用户未删除的视频总数，用于分页返回。
func (m *customVideoModel) CountByUserId(ctx context.Context, userId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) as total from %s where `user_id` = ? and `delete_at` is null", m.table)

	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, userId); err != nil {
		return 0, err
	}
	return row.Total, nil
}
