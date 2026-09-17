package main

import (
	"context"
	"flag"
	"fmt"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/config"
	"tiktok-go/app/interaction/rpc/internal/server"
	"tiktok-go/app/interaction/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interaction.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	svcCtx := svc.NewServiceContext(c)

	// 点赞的落库任务随进程启停：接口只写队列，由它按周期刷进数据库
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svcCtx.LikeWorker.Start(appCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		interaction.RegisterInteractionServiceServer(grpcServer, server.NewInteractionServiceServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
