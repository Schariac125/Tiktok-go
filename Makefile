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
        image-build \
        infra-up infra-down infra-logs \
        rocketmq-hosts rocketmq-check \
        run-gateway clean

# 生成全部代码：网关 + 各模块 rpc
generate: api-build rpc-build

# 生成网关代码。
# 各模块的 .api 用 group 属性声明分组，handler/logic 会落到 internal/handler/<group>
# 与 internal/logic/<group> 子包，无需生成后再搬目录。
# 已存在的文件（config、service_context、以及各 handler/logic 业务实现）不会被覆盖。
api-build:
	goctl api go --api $(GATEWAY_API) --dir $(GATEWAY_DIR) --style $(API_STYLE)

# 生成各模块 rpc 代码。
# proto 里的 go_package 为 "./<模块>"，因此 pb 文件落在 app/<模块>/rpc/<模块>。
# 已实现的 logic、model、etc 配置不会被覆盖。
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

# 从 migration/<模块>.sql 生成带 redis 缓存的 model。
# 生成 <模块>_model_gen.go / vars.go，手写的 <模块>_model.go 不会被覆盖。
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

# 让宿主机也能解析 host.docker.internal，指向 127.0.0.1。
#
# 只有「宿主机与容器混跑」才需要：这时 broker 注册的是 host.docker.internal，
# 容器侧由 Docker Desktop 解析，宿主机侧要这一行。同时 BIND_ADDR 得设成 0.0.0.0，
# 否则端口只绑环回，容器够不着。
#
# 纯本地开发（服务都在宿主机上）用 ROCKETMQ_BROKER_IP=127.0.0.1，不需要这一步。
rocketmq-hosts:
	@grep -q "host.docker.internal" /etc/hosts \
		&& echo "已存在，跳过" \
		|| echo "127.0.0.1 host.docker.internal" | sudo tee -a /etc/hosts

# 查看 broker 注册到 NameServer 的地址；客户端连不上时先看这里。
#
# 在独立的临时容器里跑，不在 broker 容器内跑：broker 注册的是
# host.docker.internal，从它自己容器里连这个地址会绕回宿主机的 published
# 端口再 DNAT 回自身，形成 hairpin 而失败——真实客户端（interaction-rpc）
# 是第三方，不受影响，所以这里也用一个第三方容器来模拟。
rocketmq-check:
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
