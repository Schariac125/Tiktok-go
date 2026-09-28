package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentModel = (*customCommentModel)(nil)

type (
	// CommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentModel.
	CommentModel interface {
		commentModel
		// FindByVideo 取某个视频下的一级评论（parent_id 为 0），按时间倒序。
		FindByVideo(ctx context.Context, videoId uint64, offset, limit int64) ([]*Comment, error)
		// FindByParent 取某条评论下的回复，按时间正序（回复按先后展示）。
		FindByParent(ctx context.Context, parentId uint64, offset, limit int64) ([]*Comment, error)
		// SoftDelete 软删除单条评论，已删除时返回 ErrNotFound。
		SoftDelete(ctx context.Context, id uint64) error
		// SoftDeleteByParent 级联软删除某条评论下的所有回复，返回删除条数。
		SoftDeleteByParent(ctx context.Context, parentId uint64) (int64, error)
		// SoftDeleteByVideo 级联软删除某个视频下的所有评论，返回删除条数。
		SoftDeleteByVideo(ctx context.Context, videoId uint64) (int64, error)
		// FindIdsByVideo 取某个视频下所有未删除评论的 id，用于连带清理点赞。
		FindIdsByVideo(ctx context.Context, videoId uint64) ([]uint64, error)
		// FindIdsByParent 取某条评论下所有未删除回复的 id。
		FindIdsByParent(ctx context.Context, parentId uint64) ([]uint64, error)
		// SetLikeCount 把点赞数设为绝对值，供物化任务写回。
		SetLikeCount(ctx context.Context, id uint64, count int64) error
		// IncrChildCount 增减子评论数，不会减到 0 以下。
		IncrChildCount(ctx context.Context, id uint64, delta int64) error
	}

	customCommentModel struct {
		*defaultCommentModel
	}
)

// NewCommentModel returns a model for the database table.
func NewCommentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentModel {
	return &customCommentModel{
		defaultCommentModel: newCommentModel(conn, c, opts...),
	}
}

// FindByVideo 取某个视频下的一级评论，最新的在前。
func (m *customCommentModel) FindByVideo(ctx context.Context, videoId uint64, offset, limit int64) ([]*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `video_id` = ? and `parent_id` = 0 and `deleted_at` is null order by `id` desc limit ? offset ?",
		commentRows, m.table)

	var list []*Comment
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, videoId, limit, offset); err != nil {
		return nil, err
	}
	return list, nil
}

// FindByParent 取某条评论下的回复，先回复的在前。
func (m *customCommentModel) FindByParent(ctx context.Context, parentId uint64, offset, limit int64) ([]*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `parent_id` = ? and `deleted_at` is null order by `id` asc limit ? offset ?",
		commentRows, m.table)

	var list []*Comment
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, parentId, limit, offset); err != nil {
		return nil, err
	}
	return list, nil
}

// SoftDelete 软删除单条评论，记录保留在表里。
func (m *customCommentModel) SoftDelete(ctx context.Context, id uint64) error {
	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set `deleted_at` = now() where `id` = ? and `deleted_at` is null", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, fmt.Sprintf("%s%v", cacheCommentIdPrefix, id))
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

// SoftDeleteByParent 级联软删除某条评论下的所有回复，返回实际删除条数。
// 评论只有两级，子评论的 parent_id 直接指向被删的评论，不需要递归。
func (m *customCommentModel) SoftDeleteByParent(ctx context.Context, parentId uint64) (int64, error) {
	query := fmt.Sprintf("update %s set `deleted_at` = now() where `parent_id` = ? and `deleted_at` is null", m.table)
	return m.affectedRows(ctx, query, parentId)
}

// SoftDeleteByVideo 级联软删除某个视频下的所有评论（含回复），返回实际删除条数。
func (m *customCommentModel) SoftDeleteByVideo(ctx context.Context, videoId uint64) (int64, error) {
	query := fmt.Sprintf("update %s set `deleted_at` = now() where `video_id` = ? and `deleted_at` is null", m.table)
	return m.affectedRows(ctx, query, videoId)
}

// FindIdsByVideo 取某个视频下所有未删除评论的 id，含回复，用于连带清理它们的点赞。
func (m *customCommentModel) FindIdsByVideo(ctx context.Context, videoId uint64) ([]uint64, error) {
	query := fmt.Sprintf("select `id` from %s where `video_id` = ? and `deleted_at` is null", m.table)

	var rows []struct {
		Id uint64 `db:"id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, videoId); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	return ids, nil
}

// FindIdsByParent 取某条评论下所有未删除回复的 id，用于连带清理它们的点赞。
func (m *customCommentModel) FindIdsByParent(ctx context.Context, parentId uint64) ([]uint64, error) {
	query := fmt.Sprintf("select `id` from %s where `parent_id` = ? and `deleted_at` is null", m.table)

	var rows []struct {
		Id uint64 `db:"id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, parentId); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	return ids, nil
}

// SetLikeCount 把点赞数设为绝对值。
//
// 点赞数的权威来源是 Redis 的点赞集合，这一列只是物化出来的副本，
// 只能整体覆盖，不能做增量——增量会在「副本比 Redis 旧」时越算越偏。
func (m *customCommentModel) SetLikeCount(ctx context.Context, id uint64, count int64) error {
	query := fmt.Sprintf("update %s set `like_count` = ? where `id` = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, count, id)
	return err
}

// IncrChildCount 增减子评论数，同样做 0 兜底。
func (m *customCommentModel) IncrChildCount(ctx context.Context, id uint64, delta int64) error {
	query := fmt.Sprintf("update %s set `child_count` = greatest(cast(`child_count` as signed) + ?, 0) where `id` = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, delta, id)
	return err
}

// affectedRows 执行批量 update 并返回影响行数
func (m *customCommentModel) affectedRows(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := m.ExecNoCacheCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
