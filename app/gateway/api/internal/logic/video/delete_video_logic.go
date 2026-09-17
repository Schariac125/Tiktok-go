// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVideoLogic {
	return &DeleteVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除视频 DELETE /video/delete
// 视频归属由 rpc 侧校验，操作者身份取自登录态
func (l *DeleteVideoLogic) DeleteVideo(req *types.DeleteVideoReq) (*types.DeleteVideoResp, error) {
	userId := jwt.UserIdFromContext(l.ctx)
	if userId <= 0 {
		return &types.DeleteVideoResp{Base: toStatus(errno.ErrUnauthorized)}, nil
	}

	rpcResp, err := l.svcCtx.VideoRpc.DeleteVideo(l.ctx, &video.DeleteVideoReq{
		UserId: userId,
		Id:     req.Id,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.DeleteVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	// 视频删成功后，连带清理它的评论与点赞（数据在 interaction 服务里）。
	// 顺序是先删主体再清附属：反过来一旦删视频失败，用户会看到视频还在、评论却没了。
	// 清理失败不回滚也不报错——视频已经删掉了，残留的互动数据不影响主流程，记日志即可。
	if rpcResp.Base.Code == errno.OK.Code {
		if _, err := l.svcCtx.InteractionRpc.CleanByVideo(l.ctx, &interaction.CleanByVideoReq{VideoId: req.Id}); err != nil {
			l.Errorf("清理视频 %d 的互动数据失败: %v", req.Id, err)
		}
	}

	return &types.DeleteVideoResp{Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg}}, nil
}
