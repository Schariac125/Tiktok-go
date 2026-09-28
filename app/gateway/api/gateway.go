// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"tiktok-go/app/gateway/api/internal/config"
	"tiktok-go/app/gateway/api/internal/handler"
	"tiktok-go/app/gateway/api/internal/svc"
	"tiktok-go/app/gateway/api/internal/types"
	"tiktok-go/pkg/env"
	"tiktok-go/pkg/errno"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

// errorBody 异常路径的响应体，与接口文档约定的 {base:{code,msg}} 保持一致
type errorBody struct {
	Base types.Status `json:"base"`
}

func main() {
	flag.Parse()

	var c config.Config
	env.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 框架异常出口：让参数解析失败、RPC 传输失败等也走统一契约
	// （HTTP 200 + {base:{code,msg}}），不再落到 go-zero 默认的
	// HTTP 400 + {code,msg}。未匹配路由不经过此处，仍是 HTTP 404。
	httpx.SetErrorHandlerCtx(func(ctx context.Context, err error) (int, any) {
		e := errno.FromErr(ctx, err)
		return http.StatusOK, errorBody{Base: types.Status{Code: e.Code, Msg: e.Msg}}
	})

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
