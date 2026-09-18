package logic

import (
	"context"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowingLogic {
	return &ListFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看某个人的关注列表 GET /follow/following
// 关系表只存 id，用户名跨服务从 user 取回来再拼装。
func (l *ListFollowingLogic) ListFollowing(in *social.ListFollowingReq) (*social.ListFollowingResp, error) {
	if in.UserId <= 0 {
		return &social.ListFollowingResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("用户 id 不能为空"))}, nil
	}
	userId := uint64(in.UserId)

	total, err := l.svcCtx.RelationModel.CountFollowing(l.ctx, userId)
	if err != nil {
		return nil, err
	}

	var ids []uint64
	if total > 0 {
		pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
		ids, err = l.svcCtx.RelationModel.FindFollowing(l.ctx, userId, (pageNum-1)*pageSize, pageSize)
		if err != nil {
			return nil, err
		}
	}

	data, err := buildUserBriefPage(l.ctx, l.svcCtx, ids, total)
	if err != nil {
		return nil, err
	}

	return &social.ListFollowingResp{Base: toStatus(errno.OK), Data: data}, nil
}
