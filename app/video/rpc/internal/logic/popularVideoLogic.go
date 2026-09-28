package logic

import (
	"context"
	"encoding/json"

	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/model"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/errno"
	"tiktok-go/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type PopularVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPopularVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PopularVideoLogic {
	return &PopularVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 热门排行榜 GET /video/popular
// 榜单按观看量倒序，整份缓存在 redis 里，分页从缓存切片；
// 缓存未命中或 redis 故障时回退直查数据库，不影响接口可用性。
func (l *PopularVideoLogic) PopularVideo(in *video.PopularVideoReq) (*video.PopularVideoResp, error) {
	rank, err := l.loadRank()
	if err != nil {
		return nil, err
	}

	pageNum, pageSize := utils.NormalizePage(in.PageNum, in.PageSize)
	offset := (pageNum - 1) * pageSize

	return &video.PopularVideoResp{
		Base: toStatus(errno.OK),
		Data: &video.VideoFeedData{Items: convertVideos(l.ctx, l.svcCtx.Redis, utils.SlicePage(rank, offset, pageSize))},
	}, nil
}

// loadRank 取榜单：先读缓存，未命中或不可用时查库并回填。
// 榜单允许滞后，靠 TTL 自然过期，不做主动失效。
func (l *PopularVideoLogic) loadRank() ([]*model.Video, error) {
	if l.svcCtx.Redis != nil {
		// 单独设超时，redis 挂掉时快速失败去查库，而不是拖到 gRPC 整体超时
		ctx, cancel := context.WithTimeout(l.ctx, constant.RedisOpTimeout)
		raw, err := l.svcCtx.Redis.GetCtx(ctx, constant.PopularRankCacheKey)
		cancel()

		switch {
		case err == nil:
			var cached []*model.Video
			if err := json.Unmarshal([]byte(raw), &cached); err == nil {
				return cached, nil
			}
			// 反序列化失败说明缓存内容不可用，当作未命中走查库
			l.Errorf("热门榜单缓存内容异常，回退查库")
		case err != redis.Nil:
			// Nil 是正常的未命中，其他错误才需要记录
			l.Errorf("读取热门榜单缓存失败，回退查库: %v", err)
		}
	}

	rank, err := l.svcCtx.VideoModel.FindPopular(l.ctx, constant.PopularRankSize)
	if err != nil {
		return nil, err
	}

	// 空榜单也缓存，避免缓存被击穿时每次都查库
	if l.svcCtx.Redis != nil {
		if data, err := json.Marshal(rank); err == nil {
			ctx, cancel := context.WithTimeout(l.ctx, constant.RedisOpTimeout)
			err := l.svcCtx.Redis.SetexCtx(ctx, constant.PopularRankCacheKey, string(data), int(constant.PopularRankTTL.Seconds()))
			cancel()
			if err != nil {
				l.Errorf("回填热门榜单缓存失败: %v", err)
			}
		}
	}
	return rank, nil
}
