package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/internal/task"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 点赞 POST /like/action
// 高频操作，只投递消息就返回，不碰数据库。
// 落库与计数由消费端异步完成，滞后在毫秒级。
func (l *LikeLogic) Like(in *interaction.LikeReq) (*interaction.LikeResp, error) {
	if in.UserId <= 0 {
		return &interaction.LikeResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	targetType, targetId, e := resolveLikeTarget(in.VideoId, in.CommentId)
	if e != nil {
		return &interaction.LikeResp{Base: toStatus(e)}, nil
	}

	if err := task.PushLike(l.ctx, l.svcCtx.LikePublisher, uint64(in.UserId), targetType, targetId); err != nil {
		l.Errorf("投递点赞消息失败: %v", err)
		return &interaction.LikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	return &interaction.LikeResp{Base: toStatus(errno.OK)}, nil
}
