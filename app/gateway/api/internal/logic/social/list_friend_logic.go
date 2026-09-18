// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package social

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/social/rpc/social"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendLogic {
	return &ListFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 好友列表 GET /relation/friends
// 只允许获取自己的：查询者身份取自登录态，不接收调用方传入的 user_id
func (l *ListFriendLogic) ListFriend(req *types.ListFriendReq) (*types.ListFriendResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.ListFriendResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.SocialRpc.ListFriend(l.ctx, &social.ListFriendReq{
		UserId:   userId,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.ListFriendResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.ListFriendResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.UserBriefPageData{
			Items: convertUserBriefs(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
