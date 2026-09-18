package logic

import (
	"context"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendLogic {
	return &ListFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 好友列表 GET /relation/friends
// 好友即互相关注，user_id 由网关从登录态注入，只允许查自己的。
func (l *ListFriendLogic) ListFriend(in *social.ListFriendReq) (*social.ListFriendResp, error) {
	if in.UserId <= 0 {
		return &social.ListFriendResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}
	userId := uint64(in.UserId)

	total, err := l.svcCtx.RelationModel.CountFriends(l.ctx, userId)
	if err != nil {
		return nil, err
	}

	var ids []uint64
	if total > 0 {
		pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
		ids, err = l.svcCtx.RelationModel.FindFriends(l.ctx, userId, (pageNum-1)*pageSize, pageSize)
		if err != nil {
			return nil, err
		}
	}

	data, err := buildUserBriefPage(l.ctx, l.svcCtx, ids, total)
	if err != nil {
		return nil, err
	}

	return &social.ListFriendResp{Base: toStatus(errno.OK), Data: data}, nil
}
