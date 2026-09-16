package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
		// FindPopular 按观看量倒序取热门视频（只含未删除的），用于排行榜。
		FindPopular(ctx context.Context, limit int64) ([]*Video, error)
		// FeedLatest 取首页视频流：未删除的视频按发布时间倒序。
		FeedLatest(ctx context.Context, since time.Time, limit int64) ([]*Video, error)
		// Search 按关键词/作者/时间范围搜索视频，分页返回，最新发布的在前。
		Search(ctx context.Context, q SearchQuery) ([]*Video, error)
		// CountSearch 统计符合搜索条件的视频总数，用于分页。
		CountSearch(ctx context.Context, q SearchQuery) (int64, error)
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

// FindPopular 按观看量倒序取热门视频（只含未删除的）。
// 观看量相同时再按 id 倒序，保证榜单顺序稳定、分页不会出现重复或漏项。
func (m *customVideoModel) FindPopular(ctx context.Context, limit int64) ([]*Video, error) {
	query := fmt.Sprintf("select %s from %s where `delete_at` is null order by `visit_count` desc, `id` desc limit ?",
		videoRows, m.table)

	var list []*Video
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, limit); err != nil {
		return nil, err
	}
	return list, nil
}

// FeedLatest 取首页视频流：未删除的视频按发布时间倒序。
// since 为零值时取最新的一批；非零值时只取该时间之后的，用于下拉刷新。
// 带上 id 排序是为了同一时间戳下的次序稳定，避免刷新时出现重复或漏项。
func (m *customVideoModel) FeedLatest(ctx context.Context, since time.Time, limit int64) ([]*Video, error) {
	query := fmt.Sprintf("select %s from %s where `delete_at` is null", videoRows, m.table)
	args := make([]any, 0, 2)

	if !since.IsZero() {
		query += " and `created_at` > ?"
		args = append(args, since)
	}
	query += " order by `created_at` desc, `id` desc limit ?"
	args = append(args, limit)

	var list []*Video
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// SearchQuery 视频搜索的过滤条件，零值字段表示不限制。
type SearchQuery struct {
	Keywords string // 匹配标题或描述
	UserId   uint64 // 按作者筛选
	From     int64  // 起始时间，Unix 秒
	To       int64  // 结束时间，Unix 秒
	Offset   int64  // 分页偏移
	Limit    int64  // 每页数量
}

// buildSearchWhere 拼出搜索的 where 子句与参数，Search 与 CountSearch 共用。
// 关键词用 LIKE %kw% 匹配标题和描述，前后都有通配符走不了索引；
// 数据量上来后应换成全文索引（FULLTEXT）或独立检索引擎。
// 时间条件用 from_unixtime 转换，会按 MySQL 会话时区（现为 +08:00）解释。
func buildSearchWhere(q SearchQuery) (string, []any) {
	where := "`delete_at` is null"
	args := make([]any, 0, 6)

	if kw := strings.TrimSpace(q.Keywords); kw != "" {
		where += " and (`title` like ? or `description` like ?)"
		like := "%" + kw + "%"
		args = append(args, like, like)
	}
	if q.UserId > 0 {
		where += " and `user_id` = ?"
		args = append(args, q.UserId)
	}
	if q.From > 0 {
		where += " and `created_at` >= from_unixtime(?)"
		args = append(args, q.From)
	}
	if q.To > 0 {
		where += " and `created_at` <= from_unixtime(?)"
		args = append(args, q.To)
	}

	return where, args
}

// Search 按条件搜索视频并分页返回，最新发布的在前。
func (m *customVideoModel) Search(ctx context.Context, q SearchQuery) ([]*Video, error) {
	where, args := buildSearchWhere(q)
	query := fmt.Sprintf("select %s from %s where %s order by `id` desc limit ? offset ?",
		videoRows, m.table, where)
	args = append(args, q.Limit, q.Offset)

	var list []*Video
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

// CountSearch 统计符合搜索条件的视频总数。
func (m *customVideoModel) CountSearch(ctx context.Context, q SearchQuery) (int64, error) {
	where, args := buildSearchWhere(q)
	query := fmt.Sprintf("select count(*) as total from %s where %s", m.table, where)

	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, args...); err != nil {
		return 0, err
	}
	return row.Total, nil
}
