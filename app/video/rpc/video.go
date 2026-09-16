package main

import (
	"flag"
	"fmt"

	"tiktok-go/app/video/rpc/internal/config"
	"tiktok-go/app/video/rpc/internal/server"
	"tiktok-go/app/video/rpc/internal/svc"
	"tiktok-go/app/video/rpc/video"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/video.yaml", "the config file")

// maxUploadBytes gRPC 单条消息上限，需容纳投稿的视频 + 封面
const maxUploadBytes = 128 << 20 // 128MB

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		video.RegisterVideoServiceServer(grpcServer, server.NewVideoServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// 投稿接口经 gRPC 传视频文件内容，gRPC 默认只收 4MB，按业务上限放开
	s.AddOptions(grpc.MaxRecvMsgSize(maxUploadBytes))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
