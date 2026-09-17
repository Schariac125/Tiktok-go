package logic

import (
	"context"
	"errors"

	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除评论 DELETE /comment/delete
// 评论本身软删除，但会连带清理：它下面的回复（软删）以及这些评论上的点赞（硬删）。
// 只允许作者删自己的评论。
func (l *DeleteCommentLogic) DeleteComment(in *interaction.DeleteCommentReq) (*interaction.DeleteCommentResp, error) {
	if in.UserId <= 0 || in.CommentId <= 0 {
		return &interaction.DeleteCommentResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	c, err := l.svcCtx.CommentModel.FindOne(l.ctx, uint64(in.CommentId))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &interaction.DeleteCommentResp{Base: toStatus(errno.ErrCommentNotFound)}, nil
		}
		return nil, err
	}
	if c.DeletedAt.Valid {
		return &interaction.DeleteCommentResp{Base: toStatus(errno.ErrCommentNotFound)}, nil
	}
	if c.UserId != uint64(in.UserId) {
		return &interaction.DeleteCommentResp{Base: toStatus(errno.ErrCommentForbidden)}, nil
	}

	// 评论只有两级，子评论的 parent_id 直接指向本条，取一次即可
	childIds, err := l.svcCtx.CommentModel.FindIdsByParent(l.ctx, uint64(in.CommentId))
	if err != nil {
		return nil, err
	}
	doomed := append(childIds, uint64(in.CommentId))

	// 连带删掉这些评论上的点赞
	if _, err := l.svcCtx.UserLikeModel.DeleteByTargets(l.ctx, constant.LikeTargetComment, doomed); err != nil {
		l.Errorf("清理评论点赞失败: %v", err)
	}

	// 软删子评论与本条
	if _, err := l.svcCtx.CommentModel.SoftDeleteByParent(l.ctx, uint64(in.CommentId)); err != nil {
		return nil, err
	}
	if err := l.svcCtx.CommentModel.SoftDelete(l.ctx, uint64(in.CommentId)); err != nil {
		return nil, err
	}

	// 计数要减掉被连带删除的数量：本条 + 它的回复
	removed := int64(len(doomed))
	if c.ParentId > 0 {
		// 本条是回复，父评论的子评论数要相应减少
		if err := l.svcCtx.CommentModel.IncrChildCount(l.ctx, c.ParentId, -removed); err != nil {
			l.Errorf("更新父评论子评论数失败: %v", err)
		}
	}
	syncVideoCommentCount(l.ctx, l.svcCtx, int64(c.VideoId), -removed)

	return &interaction.DeleteCommentResp{Base: toStatus(errno.OK)}, nil
}
