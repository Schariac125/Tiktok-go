package logic

import (
	"context"

	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

// syncVideoCommentCount 通知 video 服务调整视频的评论数。
// 视频表归 video 服务管，评论的增删要同步过去。
// 失败只记日志：评论本身已经写好了，不该因为计数同步失败而让整个操作失败，
// 计数偏差可以由后续的清理或订正任务兜底。
func syncVideoCommentCount(ctx context.Context, svcCtx *svc.ServiceContext, videoId, delta int64) {
	if videoId <= 0 || delta == 0 {
		return
	}

	resp, err := svcCtx.VideoRpc.IncrCommentCount(ctx, &video.IncrCommentCountReq{
		VideoId: videoId,
		Delta:   delta,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("更新视频 %d 评论数失败: %v", videoId, err)
		return
	}
	if resp.GetBase().GetCode() != errno.OK.Code {
		logx.WithContext(ctx).Errorf("更新视频 %d 评论数被拒: %s", videoId, resp.GetBase().GetMsg())
	}
}
