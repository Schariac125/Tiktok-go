package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserRelationModel = (*customUserRelationModel)(nil)

type (
	// UserRelationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserRelationModel.
	UserRelationModel interface {
		userRelationModel
		// InsertIgnore 写入一条关注关系，已关注时不报错也不重复写，返回影响行数（1 表示新增）。
		InsertIgnore(ctx context.Context, userId, toUserId uint64) (int64, error)
		// DeleteOne 删除一条关注关系（取关，硬删除），返回影响行数。
		DeleteOne(ctx context.Context, userId, toUserId uint64) (int64, error)
		// FindFollowing 分页取某人关注的人的 id，最近关注的在前。
		FindFollowing(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error)
		// CountFollowing 统计某人关注了多少人。
		CountFollowing(ctx context.Context, userId uint64) (int64, error)
		// FindFollower 分页取某人的粉丝 id，最近关注的在前。
		FindFollower(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error)
		// CountFollower 统计某人有多少粉丝。
		CountFollower(ctx context.Context, userId uint64) (int64, error)
		// FindFriends 分页取好友（互相关注）的 id。
		FindFriends(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error)
		// CountFriends 统计某人的好友数。
		CountFriends(ctx context.Context, userId uint64) (int64, error)
	}

	customUserRelationModel struct {
		*defaultUserRelationModel
	}
)

// NewUserRelationModel returns a model for the database table.
func NewUserRelationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserRelationModel {
	return &customUserRelationModel{
		defaultUserRelationModel: newUserRelationModel(conn, c, opts...),
	}
}

// InsertIgnore 依赖 uk_user_to 唯一索引去重：重复关注时影响行数为 0，调用方据此判断是否新增。
func (m *customUserRelationModel) InsertIgnore(ctx context.Context, userId, toUserId uint64) (int64, error) {
	query := fmt.Sprintf("insert ignore into %s (`user_id`, `to_user_id`) values (?, ?)", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, userId, toUserId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteOne 取关，文档要求硬删除，不留记录。
func (m *customUserRelationModel) DeleteOne(ctx context.Context, userId, toUserId uint64) (int64, error) {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `to_user_id` = ?", m.table)
	result, err := m.ExecNoCacheCtx(ctx, query, userId, toUserId)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// FindFollowing 取某人关注的人，最近关注的在前。
func (m *customUserRelationModel) FindFollowing(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error) {
	query := fmt.Sprintf("select `to_user_id` as id from %s where `user_id` = ? order by `id` desc limit ? offset ?", m.table)
	return m.queryIds(ctx, query, userId, limit, offset)
}

// CountFollowing 统计某人关注了多少人。
func (m *customUserRelationModel) CountFollowing(ctx context.Context, userId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) as total from %s where `user_id` = ?", m.table)
	return m.count(ctx, query, userId)
}

// FindFollower 取某人的粉丝，最近关注的在前。
func (m *customUserRelationModel) FindFollower(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error) {
	query := fmt.Sprintf("select `user_id` as id from %s where `to_user_id` = ? order by `id` desc limit ? offset ?", m.table)
	return m.queryIds(ctx, query, userId, limit, offset)
}

// CountFollower 统计某人有多少粉丝。
func (m *customUserRelationModel) CountFollower(ctx context.Context, userId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) as total from %s where `to_user_id` = ?", m.table)
	return m.count(ctx, query, userId)
}

// FindFriends 取好友（互相关注）的 id。
// 好友关系不单独存表：r1 是我关注的人，r2 是关注我的人，两者取交集即是好友。
func (m *customUserRelationModel) FindFriends(ctx context.Context, userId uint64, offset, limit int64) ([]uint64, error) {
	query := fmt.Sprintf(`select r1.to_user_id as id from %s r1
		join %s r2 on r1.to_user_id = r2.user_id and r2.to_user_id = r1.user_id
		where r1.user_id = ? order by r1.id desc limit ? offset ?`, m.table, m.table)
	return m.queryIds(ctx, query, userId, limit, offset)
}

// CountFriends 统计某人的好友数，连接方式同 FindFriends。
func (m *customUserRelationModel) CountFriends(ctx context.Context, userId uint64) (int64, error) {
	query := fmt.Sprintf(`select count(*) as total from %s r1
		join %s r2 on r1.to_user_id = r2.user_id and r2.to_user_id = r1.user_id
		where r1.user_id = ?`, m.table, m.table)
	return m.count(ctx, query, userId)
}

// queryIds 查出单列 id 并收集成切片。
// SQL 里统一用 `as id` 别名，因为查的可能是 to_user_id 或 user_id，
// 而 go-zero 是按 db tag 名与列名对应来扫描的。
func (m *customUserRelationModel) queryIds(ctx context.Context, query string, args ...any) ([]uint64, error) {
	var rows []struct {
		Id uint64 `db:"id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	return ids, nil
}

// count 执行单值统计
func (m *customUserRelationModel) count(ctx context.Context, query string, args ...any) (int64, error) {
	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, args...); err != nil {
		return 0, err
	}
	return row.Total, nil
}
