package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/model"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentLogic {
	return &ListCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看评论列表 GET /comment/list
// 给 video_id 取该视频的一级评论（最新在前），给 comment_id 取该评论下的回复（先回复的在前）。
// 接口不返回总数，所以这里不做 count。
func (l *ListCommentLogic) ListComment(in *interaction.ListCommentReq) (*interaction.ListCommentResp, error) {
	pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
	offset := (pageNum - 1) * pageSize

	var (
		list []*model.Comment
		err  error
	)
	switch {
	case in.CommentId > 0:
		list, err = l.svcCtx.CommentModel.FindByParent(l.ctx, uint64(in.CommentId), offset, pageSize)
	case in.VideoId > 0:
		list, err = l.svcCtx.CommentModel.FindByVideo(l.ctx, uint64(in.VideoId), offset, pageSize)
	default:
		return &interaction.ListCommentResp{
			Base: toStatus(errno.ErrInvalidParam.WithMsg("视频 id 和评论 id 至少指定一个")),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &interaction.ListCommentResp{
		Base:  toStatus(errno.OK),
		Items: convertComments(list),
	}, nil
}
