package logic

import (
	"context"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取关 DELETE /relation/unfollow
// 硬删除；没关注过也返回成功——取关是幂等的，重复调用不该报错。
func (l *UnfollowLogic) Unfollow(in *social.UnfollowReq) (*social.UnfollowResp, error) {
	if in.UserId <= 0 || in.ToUserId <= 0 {
		return &social.UnfollowResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	if _, err := l.svcCtx.RelationModel.DeleteOne(l.ctx, uint64(in.UserId), uint64(in.ToUserId)); err != nil {
		l.Errorf("删除关注关系失败: %v", err)
		return nil, err
	}

	return &social.UnfollowResp{Base: toStatus(errno.OK)}, nil
}
