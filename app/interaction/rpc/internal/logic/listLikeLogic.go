package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLikeLogic {
	return &ListLikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看某个用户点赞的视频列表 GET /like/list
// 点赞关系在本服务，视频数据在 video 服务，这里跨服务取详情后按点赞时间序拼装。
// 分页对的是点赞关系（最近点赞的在前），再按这个顺序去取视频详情。
func (l *ListLikeLogic) ListLike(in *interaction.ListLikeReq) (*interaction.ListLikeResp, error) {
	if in.UserId <= 0 {
		return &interaction.ListLikeResp{Base: toStatus(errno.ErrInvalidParam.WithMsg("用户 id 不能为空"))}, nil
	}
	userId := uint64(in.UserId)

	total, err := l.svcCtx.UserLikeModel.CountByUser(l.ctx, userId, constant.LikeTargetVideo)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &interaction.ListLikeResp{
			Base: toStatus(errno.OK),
			Data: &interaction.VideoPageData{Items: []*interaction.Video{}},
		}, nil
	}

	pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
	offset := (pageNum - 1) * pageSize

	ids, err := l.svcCtx.UserLikeModel.FindVideoIdsByUser(l.ctx, userId, offset, pageSize)
	if err != nil {
		return nil, err
	}

	videoResp, err := l.svcCtx.VideoRpc.BatchGetVideo(l.ctx, &video.BatchGetVideoReq{Ids: toInt64Slice(ids)})
	if err != nil {
		return nil, err
	}
	if videoResp.GetBase().GetCode() != errno.OK.Code {
		return &interaction.ListLikeResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	// 批量查不保证顺序，按点赞时间序（ids 的顺序）重排；
	// 期间被删掉的视频不会返回，跳过即可
	byId := make(map[int64]*video.Video, len(videoResp.GetItems()))
	for _, v := range videoResp.GetItems() {
		byId[v.GetId()] = v
	}
	items := make([]*interaction.Video, 0, len(ids))
	for _, id := range ids {
		if v, ok := byId[int64(id)]; ok {
			items = append(items, convertVideo(v))
		}
	}

	return &interaction.ListLikeResp{
		Base: toStatus(errno.OK),
		Data: &interaction.VideoPageData{Items: items, Total: total},
	}, nil
}

// toInt64Slice 把 id 列表转成 rpc 需要的类型
func toInt64Slice(ids []uint64) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		out = append(out, int64(id))
	}
	return out
}
