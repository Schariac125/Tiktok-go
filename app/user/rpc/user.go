package main

import (
	"flag"
	"fmt"

	"tiktok-go/app/user/rpc/internal/config"
	"tiktok-go/app/user/rpc/internal/server"
	"tiktok-go/app/user/rpc/internal/svc"
	"tiktok-go/app/user/rpc/user"
	"tiktok-go/pkg/constant"
	"tiktok-go/pkg/env"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	env.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	// 上传头像经 gRPC 传文件内容，gRPC 默认只收 4MB，按头像业务上限放开
	s.AddOptions(grpc.MaxRecvMsgSize(constant.GRPCMsgMaxSize))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
