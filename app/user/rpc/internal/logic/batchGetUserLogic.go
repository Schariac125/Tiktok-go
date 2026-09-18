package logic

import (
	"context"

	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserLogic {
	return &BatchGetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按 id 批量查用户，供其他服务补齐用户名等信息（如社交的关注/粉丝列表）。
// 不保证返回顺序与入参一致，调用方按 id 自己对齐；
// 已删除的用户不会返回，调用方需要容忍数量少于入参。
func (l *BatchGetUserLogic) BatchGetUser(in *user.BatchGetUserReq) (*user.BatchGetUserResp, error) {
	ids := make([]uint64, 0, len(in.Ids))
	for _, id := range in.Ids {
		if id > 0 {
			ids = append(ids, uint64(id))
		}
	}
	if len(ids) == 0 {
		return &user.BatchGetUserResp{
			Base:  toStatus(errno.OK),
			Items: convertUsers(nil),
		}, nil
	}

	list, err := l.svcCtx.UserModel.FindByIds(l.ctx, ids)
	if err != nil {
		return nil, err
	}

	return &user.BatchGetUserResp{
		Base:  toStatus(errno.OK),
		Items: convertUsers(list),
	}, nil
}
