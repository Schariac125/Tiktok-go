// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package video

import (
	"context"
	"strings"

	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/app/video/rpc/video"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchVideoLogic {
	return &SearchVideoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 搜索视频 POST /video/search
// 无需登录。username 到 user_id 的换算在这里完成——video 服务不掌握用户数据，
// 由网关编排 user 与 video 两个 rpc，video 侧始终只认 user_id。
func (l *SearchVideoLogic) SearchVideo(req *types.SearchVideoReq) (*types.SearchVideoResp, error) {
	var userId int64

	if name := strings.TrimSpace(req.Username); name != "" {
		userResp, err := l.svcCtx.UserRpc.GetUserByUsername(l.ctx, &user.GetUserByUsernameReq{Username: name})
		if err != nil {
			return nil, err
		}

		base := userResp.GetBase()
		switch base.GetCode() {
		case errno.OK.Code:
			userId = userResp.GetData().GetId()
		case errno.ErrUserNotFound.Code:
			// 没有这个用户，自然搜不到他的视频：按空结果返回，不当作错误
			return &types.SearchVideoResp{
				Base: toStatus(errno.OK),
				Data: types.VideoPageData{Items: []types.Video{}},
			}, nil
		default:
			// 其余非成功状态原样透出
			return &types.SearchVideoResp{
				Base: types.Status{Code: base.GetCode(), Msg: base.GetMsg()},
			}, nil
		}
	}

	rpcResp, err := l.svcCtx.VideoRpc.SearchVideo(l.ctx, &video.SearchVideoReq{
		Keywords: req.Keywords,
		PageSize: req.PageSize,
		PageNum:  req.PageNum,
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}
	if rpcResp == nil || rpcResp.Base == nil {
		return &types.SearchVideoResp{Base: toStatus(errno.ErrInternal)}, nil
	}

	data := rpcResp.GetData()
	return &types.SearchVideoResp{
		Base: types.Status{Code: rpcResp.Base.Code, Msg: rpcResp.Base.Msg},
		Data: types.VideoPageData{
			Items: convertVideos(data.GetItems()),
			Total: data.GetTotal(),
		},
	}, nil
}
