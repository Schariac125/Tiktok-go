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

type ListFollowingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowingLogic {
	return &ListFollowingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 查看某个人的关注列表 GET /follow/following
// 无需登录；用户名由 social 服务跨服务从 user 取回
func (l *ListFollowingLogic) ListFollowing(req *types.ListFollowingReq) (*types.ListFollowingResp, error) {
	rpcResp, err := l.svcCtx.SocialRpc.ListFollowing(l.ctx, &social.ListFollowingReq{
		UserId:   req.UserId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListFollowingResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.ListFollowingResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.UserBriefPageData{
			Items: convertUserBriefs(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
