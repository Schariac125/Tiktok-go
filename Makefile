# tiktok-go 微服务脚手架
#
# 目录约定
#   api/                 接口定义，只读源文件，不放生成产物
#     main.api             网关入口，import 各模块的 .api
#     model.api            跨模块共享的数据模型
#     <模块>/<模块>.api      网关 HTTP 接口（无需鉴权）
#     <模块>/<模块>_auth.api 网关 HTTP 接口（需鉴权，走 AuthMiddleware）
#     <模块>/<模块>.proto    rpc 接口
#   app/                 生成产物 + 业务实现
#     gateway/api          由 api/main.api 生成
#     <模块>/rpc            由 api/<模块>/<模块>.proto 生成
#
# 新增模块：在 api/ 下建同名目录放 .api/.proto，并把模块名加进 MODULES。

SHELL := /bin/bash

# go module 名，goctl 据此推导生成代码的 import 路径
MODULE := tiktok-go

# 业务模块列表，决定 rpc-build / model-build / build 要处理哪些模块
MODULES := user video interaction social

# goctl 命名风格
#   api 网关用 go_zero：bind_m_f_a_handler.go
#   rpc 服务用 goZero：bindMFALogic.go
#   model 用 go_zero：user_model.go
API_STYLE   := go_zero
RPC_STYLE   := goZero
MODEL_STYLE := go_zero

GATEWAY_DIR := app/gateway/api
GATEWAY_API := api/main.api
BIN_DIR     := bin

.DEFAULT_GOAL := generate
.PHONY: generate api-build rpc-build model-build build \
        fmt vet \
        image-build \
        infra-up infra-down infra-logs \
        rocketmq-hosts rocketmq-check \
        run-gateway clean

# 生成全部代码：网关 + 各模块 rpc
generate: api-build rpc-build

api-build:
	goctl api go --api $(GATEWAY_API) --dir $(GATEWAY_DIR) --style $(API_STYLE)

rpc-build:
	@for m in $(MODULES); do \
		if [ ! -f api/$$m/$$m.proto ]; then echo "跳过 $$m：缺少 api/$$m/$$m.proto"; continue; fi; \
		echo ">>> 生成 $$m rpc"; \
		mkdir -p app/$$m/rpc; \
		goctl rpc protoc api/$$m/$$m.proto \
			--go_out=app/$$m/rpc \
			--go-grpc_out=app/$$m/rpc \
			--zrpc_out=app/$$m/rpc \
			--module=$(MODULE) \
			--style $(RPC_STYLE) || exit 1; \
	done

model-build:
	@for m in $(MODULES); do \
		if [ ! -f migration/$$m.sql ]; then echo "跳过 $$m：缺少 migration/$$m.sql"; continue; fi; \
		if [ ! -d app/$$m/rpc ]; then echo "跳过 $$m：app/$$m/rpc 不存在，请先 make rpc-build"; continue; fi; \
		echo ">>> 生成 $$m model"; \
		goctl model mysql ddl --src migration/$$m.sql --dir app/$$m/rpc/model \
			--style $(MODEL_STYLE) --cache || exit 1; \
	done

# 编译所有服务到 bin/
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/gateway ./$(GATEWAY_DIR)
	@for m in $(MODULES); do \
		if [ ! -d app/$$m/rpc ]; then echo "跳过 $$m-rpc：app/$$m/rpc 不存在，请先 make rpc-build"; continue; fi; \
		go build -o $(BIN_DIR)/$$m-rpc ./app/$$m/rpc || exit 1; \
	done

# 代码格式化。CI 里装的是 gofumpt（见 .github/workflows/ci.yml），本地没装就退回 gofmt。
# 安装：go install mvdan.cc/gofumpt@v0.12.0
fmt:
	@if command -v gofumpt >/dev/null 2>&1; then \
		echo ">>> gofumpt -l -w ."; \
		gofumpt -l -w .; \
	else \
		echo ">>> 未找到 gofumpt，退回 gofmt -l -w ."; \
		gofmt -l -w .; \
	fi

# 静态检查，只报问题不改代码
vet:
	go vet ./...

# 构建各服务镜像。Dockerfile 在服务目录下，但构建 context 是项目根（go.mod 在那里）
image-build:
	@for m in $(MODULES); do \
		echo ">>> 构建 $$m-rpc 镜像"; \
		docker build -f app/$$m/rpc/Dockerfile -t tiktok-go/$$m-rpc:latest . || exit 1; \
	done
	@echo ">>> 构建 gateway 镜像"
	@docker build -f app/gateway/api/Dockerfile -t tiktok-go/gateway:latest .

# 启动依赖中间件（etcd、mysql、redis、rocketmq），配置见 docker-compose.yml。
# 数据卷由 compose 自动创建，不需要手工准备。
infra-up:
	docker compose up -d

rocketmq-hosts:
	@grep -q "host.docker.internal" /etc/hosts \
		&& echo "已存在，跳过" \
		|| echo "127.0.0.1 host.docker.internal" | sudo tee -a /etc/hosts

	@docker run --rm --network tiktok-go_default --entrypoint /bin/bash apache/rocketmq:4.9.7 \
		-c 'ROCKETMQ_BIN=$$(echo /home/rocketmq/rocketmq-*/bin); \
		    sh $$ROCKETMQ_BIN/mqadmin clusterList -n rocketmq-namesrv:9876'

# 停止中间件，数据卷保留
infra-down:
	docker compose down

# 跟踪中间件日志
infra-logs:
	docker compose logs -f

# 启动网关 make run-gateway
run-gateway:
	go run ./$(GATEWAY_DIR) -f $(GATEWAY_DIR)/etc/gateway.yaml

# 启动某个模块的 rpc 服务 make run-user / make run-video
run-%:
	go run ./app/$*/rpc -f app/$*/rpc/etc/$*.yaml

# 清理编译产物
clean:
	rm -rf $(BIN_DIR)
