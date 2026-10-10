package logic

import (
	"context"
	"errors"
	"strings"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishCommentLogic {
	return &PublishCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发表评论 POST /comment/publish
// video_id 与 comment_id 至少给一个：给 comment_id 表示回复某条评论。
func (l *PublishCommentLogic) PublishComment(in *interaction.PublishCommentReq) (*interaction.PublishCommentResp, error) {
	if in.UserId <= 0 {
		return &interaction.PublishCommentResp{Base: toStatus(errno.ErrInvalidParam)}, nil
	}

	content := strings.TrimSpace(in.Content)
	if content == "" {
		return &interaction.PublishCommentResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("评论内容不能为空"))}, nil
	}
	if len([]rune(content)) > constant.CommentMaxLen {
		return &interaction.PublishCommentResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("评论内容过长"))}, nil
	}

	videoId, parentId := in.VideoId, in.CommentId
	if videoId <= 0 && parentId <= 0 {
		return &interaction.PublishCommentResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("视频 id 和评论 id 至少指定一个"))}, nil
	}

	// 回复评论时从父评论补全所属视频，调用方就不必重复传 video_id
	if parentId > 0 && videoId <= 0 {
		parent, err := l.svcCtx.CommentModel.FindOne(l.ctx, uint64(parentId))
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return &interaction.PublishCommentResp{Base: toStatus(errno.ErrCommentNotFound)}, nil
			}
			return nil, err
		}
		if parent.DeletedAt.Valid {
			return &interaction.PublishCommentResp{Base: toStatus(errno.ErrCommentNotFound)}, nil
		}
		videoId = int64(parent.VideoId)
	}

	if _, err := l.svcCtx.CommentModel.Insert(l.ctx, &model.Comment{
		UserId:   uint64(in.UserId),
		VideoId:  uint64(videoId),
		ParentId: uint64(parentId),
		Content:  content,
	}); err != nil {
		l.Errorf("写入评论失败: %v", err)
		return nil, err
	}

	// 回复场景下父评论的子评论数要 +1
	if parentId > 0 {
		if err := l.svcCtx.CommentModel.IncrChildCount(l.ctx, uint64(parentId), 1); err != nil {
			l.Errorf("更新父评论子评论数失败: %v", err)
		}
	}
	// 视频的评论数同样 +1
	syncVideoCommentCount(l.ctx, l.svcCtx, videoId, 1)

	return &interaction.PublishCommentResp{Base: toStatus(errno.OK)}, nil
}
