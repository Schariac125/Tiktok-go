package main

import (
	"context"
	"flag"
	"fmt"

	"tiktok-go/app/interaction/rpc/interaction"
	"tiktok-go/app/interaction/rpc/internal/config"
	"tiktok-go/app/interaction/rpc/internal/server"
	"tiktok-go/app/interaction/rpc/internal/svc"
	"tiktok-go/pkg/env"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interaction.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	env.MustLoad(*configFile, &c)

	// 点赞落库的消费端随进程启停：接口只投递消息，由它异步刷进数据库。
	// ctx 先建，传给 ServiceContext 用于订阅；退出时 cancel 停掉消费，
	// Close 把消费进度同步回 broker。
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svcCtx := svc.NewServiceContext(appCtx, c)
	defer svcCtx.Close()

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
