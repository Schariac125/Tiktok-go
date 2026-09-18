// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package social

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFollowerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowerLogic {
	return &ListFollowerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 粉丝列表 GET /follow/follower
// 无需登录
func (l *ListFollowerLogic) ListFollower(req *types.ListFollowerReq) (*types.ListFollowerResp, error) {
	rpcResp, err := l.svcCtx.SocialRpc.ListFollower(l.ctx, &social.ListFollowerReq{
		UserId:   req.UserId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListFollowerResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.ListFollowerResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.UserBriefPageData{
			Items: convertUserBriefs(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
