package utils

import "tiktok-go/pkg/constant"

// NormalizePage 归一化分页参数：页码从 1 起，每页数量落在 [1, constant.MaxPageSize]。
// 列表类接口在查询前都先过一遍，避免调用方传入 0、负数或过大的值——
// 直接拿调用方的值拼 LIMIT/OFFSET 会出现 LIMIT 0 返回空、offset 为负导致 SQL 报错。
//
// 注意这里不能写成 max(pageSize, DefaultPageSize)：DefaultPageSize 是「调用方没传时
// 用多少」的兜底值，不是「最少给多少」，用 max 会把 page_size=2 这种正常值也撑成默认值。
func NormalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = constant.DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = constant.DefaultPageSize
	}
	pageSize = min(pageSize, constant.MaxPageSize)
	return pageNum, pageSize
}

// SlicePage 从已排好序的列表里切出第 offset 条起、长度 limit 的一段，
// 供「整份缓存 + 内存分页」的场景使用（如热门榜单：榜单整体缓存，
// 各页都从这份榜单里切，而不是按页缓存）。
// offset 越界时返回 nil，调用方按空列表处理即可。
func SlicePage[T any](list []T, offset, limit int64) []T {
	if offset < 0 || offset >= int64(len(list)) {
		return nil
	}
	return list[offset:min(offset+limit, int64(len(list)))]
}
