package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/internal/task"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消点赞 DELETE /like/delete
// 与点赞同样投递消息；落库时是硬删除，不留记录。
func (l *UnlikeLogic) Unlike(in *interaction.UnlikeReq) (*interaction.UnlikeResp, error) {
	if in.UserId <= 0 {
		return &interaction.UnlikeResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	targetType, targetId, e := resolveLikeTarget(in.VideoId, in.CommentId)
	if e != nil {
		return &interaction.UnlikeResp{Base: toStatus(e)}, nil
	}

	if err := task.PushUnlike(l.ctx, l.svcCtx.LikePublisher, uint64(in.UserId), targetType, targetId); err != nil {
		l.Errorf("投递取消点赞消息失败: %v", err)
		return &interaction.UnlikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &interaction.UnlikeResp{Base: toStatus(errno.OK)}, nil
}
