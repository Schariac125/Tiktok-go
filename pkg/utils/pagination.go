package utils

import "tiktok-go/pkg/constant"

// NormalizePage 归一化分页参数：页码从 1 起，每页数量落在 [1, constant.MaxPageSize]。
// 列表类接口在查询前都先过一遍，避免调用方传入 0、负数或过大的值——
// 直接拿调用方的值拼 LIMIT/OFFSET 会出现 LIMIT 0 返回空、offset 为负导致 SQL 报错。
func NormalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = constant.DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = constant.DefaultPageSize
	}
	if pageSize > constant.MaxPageSize {
		pageSize = constant.MaxPageSize
	}
	return pageNum, pageSize
}
