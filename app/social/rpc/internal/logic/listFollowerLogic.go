package logic

import (
	"context"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowerLogic {
	return &ListFollowerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 粉丝列表 GET /follow/follower
// 与关注列表方向相反：查 to_user_id 是我的人。
func (l *ListFollowerLogic) ListFollower(in *social.ListFollowerReq) (*social.ListFollowerResp, error) {
	if in.UserId <= 0 {
		return &social.ListFollowerResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("用户 id 不能为空"))}, nil
	}
	userId := uint64(in.UserId)

	total, err := l.svcCtx.RelationModel.CountFollower(l.ctx, userId)
	if err != nil {
		return nil, err
	}

	var ids []uint64
	if total > 0 {
		pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
		ids, err = l.svcCtx.RelationModel.FindFollower(l.ctx, userId, (pageNum-1)*pageSize, pageSize)
		if err != nil {
			return nil, err
		}
	}

	data, err := buildUserBriefPage(l.ctx, l.svcCtx, ids, total)
	if err != nil {
		return nil, err
	}

	return &social.ListFollowerResp{Base: toStatus(errno.OK), Data: data}, nil
}
