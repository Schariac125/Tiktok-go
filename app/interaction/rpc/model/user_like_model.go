package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tiktok-go/pkg/constant"
)

var _ UserLikeModel = (*customUserLikeModel)(nil)

type (
	// UserLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserLikeModel.
	UserLikeModel interface {
		userLikeModel
		// InsertIgnore 写入一条点赞，已存在时不报错也不重复写，返回影响行数（1 表示确实新增）。
		InsertIgnore(ctx context.Context, userId, targetType, targetId uint64) (int64, error)
		// DeleteOne 删除一条点赞，返回影响行数（1 表示确实删掉了）。
		DeleteOne(ctx context.Context, userId, targetType, targetId uint64) (int64, error)
		// FindVideoIdsByUser 分页取某个用户点赞的视频 id，最近点赞的在前。
		FindVideoIdsByUser(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error)
		// CountByUser 统计某个用户对某类目标的点赞总数。
		CountByUser(ctx context.Context, userId, targetType uint64) (int64, error)
		// FindUserIdsByTarget 取某个目标下所有点赞用户 id。
		// 用于 Redis 点赞集合丢失时从数据库重建，集合是这张表的投影。
		FindUserIdsByTarget(ctx context.Context, targetType, targetId uint64) ([]uint64, error)
		// DeleteByTarget 清理某个目标（视频或评论）下的所有点赞，返回删除条数。
		DeleteByTarget(ctx context.Context, targetType, targetId uint64) (int64, error)
		// DeleteByTargets 批量清理多个目标下的点赞（如视频被删时清理它所有评论的点赞）。
		DeleteByTargets(ctx context.Context, targetType uint64, targetIds []uint64) (int64, error)
	}

	customUserLikeModel struct {
		*defaultUserLikeModel
	}
)

// NewUserLikeModel returns a model for the database table.
func NewUserLikeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserLikeModel {
	return &customUserLikeModel{
		defaultUserLikeModel: newUserLikeModel(conn, c, opts...),
	}
}

// InsertIgnore 依赖 uk_user_target 唯一索引去重：
// 重复点赞时影响行数为 0，调用方据此决定要不要把计数 +1。
// 后台同步任务可能重放同一条记录，靠这个保证幂等。
func (m *customUserLikeModel) InsertIgnore(ctx context.Context, userId, targetType, targetId uint64) (int64, error) {
	query := fmt.Sprintf("insert ignore into %s (`user_id`, `target_type`, `target_id`) values (?, ?, ?)", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, userId, targetType, targetId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteOne 删除一条点赞，返回影响行数。
func (m *customUserLikeModel) DeleteOne(ctx context.Context, userId, targetType, targetId uint64) (int64, error) {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `target_type` = ? and `target_id` = ?", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, userId, targetType, targetId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// FindVideoIdsByUser 只取视频点赞（target_type 固定为视频），最近点赞的在前。
func (m *customUserLikeModel) FindVideoIdsByUser(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error) {
	query := fmt.Sprintf("select `target_id` from %s where `user_id` = ? and `target_type` = ? order by `id` desc limit ? offset ?",
		m.table)

	var rows []struct {
		TargetId uint64 `db:"target_id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, userId, constant.LikeTargetVideo, limit, offset); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.TargetId)
	}
	return ids, nil
}

// CountByUser 统计某个用户的点赞总数。
func (m *customUserLikeModel) CountByUser(ctx context.Context, userId, targetType uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) as total from %s where `user_id` = ? and `target_type` = ?", m.table)

	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, userId, targetType); err != nil {
		return 0, err
	}
	return row.Total, nil
}

// FindUserIdsByTarget 取某个目标下所有点赞用户 id。
// 走 idx_target (target_type, target_id)，不会全表扫。
func (m *customUserLikeModel) FindUserIdsByTarget(ctx context.Context, targetType, targetId uint64) ([]uint64, error) {
	query := fmt.Sprintf("select `user_id` from %s where `target_type` = ? and `target_id` = ?", m.table)

	var rows []struct {
		UserId uint64 `db:"user_id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, targetType, targetId); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserId)
	}
	return ids, nil
}

// DeleteByTarget 清理某个目标下的所有点赞（目标被删除时调用）。
func (m *customUserLikeModel) DeleteByTarget(ctx context.Context, targetType, targetId uint64) (int64, error) {
	query := fmt.Sprintf("delete from %s where `target_type` = ? and `target_id` = ?", m.table)
	return m.affectedRows(ctx, query, targetType, targetId)
}

// DeleteByTargets 批量清理多个目标下的点赞。
func (m *customUserLikeModel) DeleteByTargets(ctx context.Context, targetType uint64, targetIds []uint64) (int64, error) {
	if len(targetIds) == 0 {
		return 0, nil
	}

	args := make([]any, 0, len(targetIds)+1)
	args = append(args, targetType)
	for _, id := range targetIds {
		args = append(args, id)
	}

	query := fmt.Sprintf("delete from %s where `target_type` = ? and `target_id` in (%s)",
		m.table, strings.TrimSuffix(strings.Repeat("?,", len(targetIds)), ","))
	return m.affectedRows(ctx, query, args...)
}

// affectedRows 执行批量写并返回影响行数
func (m *customUserLikeModel) affectedRows(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := m.ExecNoCacheCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
