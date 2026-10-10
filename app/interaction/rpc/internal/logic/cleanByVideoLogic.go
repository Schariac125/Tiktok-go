package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type CleanByVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCleanByVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CleanByVideoLogic {
	return &CleanByVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 视频被删除后清理它的互动数据（内部接口，由网关在删视频后调用）。
// 清理范围：该视频下所有评论（软删）+ 这些评论和视频本身的点赞（硬删）。
func (l *CleanByVideoLogic) CleanByVideo(in *interaction.CleanByVideoReq) (*interaction.CleanByVideoResp, error) {
	if in.VideoId <= 0 {
		return &interaction.CleanByVideoResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}
	videoId := uint64(in.VideoId)

	// 先取出该视频下所有未删评论的 id（含回复），它们身上的点赞也要一起清
	commentIds, err := l.svcCtx.CommentModel.FindIdsByVideo(l.ctx, videoId)
	if err != nil {
		return nil, err
	}

	likes, err := l.svcCtx.UserLikeModel.DeleteByTargets(l.ctx, constant.LikeTargetComment, commentIds)
	if err != nil {
		return nil, err
	}
	n, err := l.svcCtx.UserLikeModel.DeleteByTarget(l.ctx, constant.LikeTargetVideo, videoId)
	if err != nil {
		return nil, err
	}
	likes += n

	comments, err := l.svcCtx.CommentModel.SoftDeleteByVideo(l.ctx, videoId)
	if err != nil {
		return nil, err
	}

	l.Infof("清理视频 %d 的互动数据：评论 %d 条，点赞 %d 条", videoId, comments, likes)
	return &interaction.CleanByVideoResp{
		Base:     toStatus(errno.OK),
		Comments: comments,
		Likes:    likes,
	}, nil
}
