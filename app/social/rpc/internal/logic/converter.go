package logic

import (
	"context"
	"fmt"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

// toStatus 将业务错误转换为 proto 的 Status 响应
func toStatus(e *errno.Errno) *social.Status {
	return &social.Status{Code: e.Code, Msg: e.Msg}
}

// buildUserBriefPage 把一批用户 id 补齐成 {id, username} 列表。
// 关系表里只有 id，用户名要跨服务从 user 取；
// 取不到的（用户已注销）直接跳过，不影响其余条目。
func buildUserBriefPage(ctx context.Context, svcCtx *svc.ServiceContext, ids []uint64, total int64) (*social.UserBriefPageData, error) {
	if len(ids) == 0 {
		// 保持非 nil slice，序列化成 [] 而不是 null
		return &social.UserBriefPageData{Items: []*social.UserBrief{}, Total: total}, nil
	}

	resp, err := svcCtx.UserRpc.BatchGetUser(ctx, &user.BatchGetUserReq{Ids: toInt64Slice(ids)})
	if err != nil {
		logx.WithContext(ctx).Errorf("补齐用户名失败: %v", err)
		return nil, err
	}
	if resp.GetBase().GetCode() != errno.OK.Code {
		return nil, fmt.Errorf("查询用户信息被拒: %s", resp.GetBase().GetMsg())
	}

	// BatchGetUser 不保证顺序，按 ids 的顺序（即关注时间序）重排
	byId := make(map[int64]*user.User, len(resp.GetItems()))
	for _, u := range resp.GetItems() {
		byId[u.GetId()] = u
	}

	items := make([]*social.UserBrief, 0, len(ids))
	for _, id := range ids {
		if u, ok := byId[int64(id)]; ok {
			items = append(items, &social.UserBrief{Id: u.GetId(), Username: u.GetUsername()})
		}
	}

	return &social.UserBriefPageData{Items: items, Total: total}, nil
}

// toInt64Slice 把 id 列表转成 rpc 需要的类型
func toInt64Slice(ids []uint64) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		out = append(out, int64(id))
	}
	return out
}
