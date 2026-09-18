package logic

import (
	"context"

	"tiktok-go/app/social/rpc/internal/svc"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关注 POST /relation/follow
// 重复关注按幂等处理：唯一索引挡掉重复写入，接口仍返回成功。
func (l *FollowLogic) Follow(in *social.FollowReq) (*social.FollowResp, error) {
	if in.UserId <= 0 || in.ToUserId <= 0 {
		return &social.FollowResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}
	if in.UserId == in.ToUserId {
		return &social.FollowResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("不能关注自己"))}, nil
	}

	// 确认被关注者存在，否则会留下指向不存在用户的关系记录
	users, err := l.svcCtx.UserRpc.BatchGetUser(l.ctx, &user.BatchGetUserReq{Ids: []int64{in.ToUserId}})
	if err != nil {
		return nil, err
	}
	if users.GetBase().GetCode() != errno.OK.Code {
		return nil, err
	}
	if len(users.GetItems()) == 0 {
		return &social.FollowResp{Base: toStatus(errno.ErrUserNotFound)}, nil
	}

	if _, err := l.svcCtx.RelationModel.InsertIgnore(l.ctx, uint64(in.UserId), uint64(in.ToUserId)); err != nil {
		l.Errorf("写入关注关系失败: %v", err)
		return nil, err
	}

	return &social.FollowResp{Base: toStatus(errno.OK)}, nil
}
