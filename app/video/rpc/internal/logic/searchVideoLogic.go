package logic

import (
	"context"
	"strings"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchVideoLogic {
	return &SearchVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 搜索视频 POST /video/search
// 关键词匹配标题或描述，可叠加作者、时间范围筛选；零值表示该条件不限。
// username 到 user_id 的换算由网关完成，这里只认 user_id。
func (l *SearchVideoLogic) SearchVideo(in *video.SearchVideoReq) (*video.SearchVideoResp, error) {
	q := model.SearchQuery{
		Keywords: strings.TrimSpace(in.Keywords),
		UserId:   uint64(in.UserId),
		From:     in.FromDate,
		To:       in.ToDate,
	}

	total, err := l.svcCtx.VideoModel.CountSearch(l.ctx, q)
	if err != nil {
		return nil, err
	}
	// 没有命中时不必再查列表
	if total == 0 {
		return &video.SearchVideoResp{
			Base: toStatus(errno.OK),
			Data: &video.VideoPageData{Items: convertVideos(nil)},
		}, nil
	}

	pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
	q.Offset = (pageNum - 1) * pageSize
	q.Limit = pageSize

	list, err := l.svcCtx.VideoModel.Search(l.ctx, q)
	if err != nil {
		return nil, err
	}

	return &video.SearchVideoResp{
		Base: toStatus(errno.OK),
		Data: &video.VideoPageData{Items: convertVideos(list), Total: total},
	}, nil
}
