.PHONY: api-build model-user-build

# 生成网关代码。
# api 文件中 service 名为 gateway，goctl 生成的入口与配置文件名
# (gateway.go / etc/gateway.yaml) 与仓库维护的文件一致，已存在则跳过，不会产生多余文件。
api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero

# 从 DDL 生成 user model（带 redis 缓存）
model-user-build:
	goctl model mysql ddl -src migration/user.sql -dir app/user/rpc/model -style go_zero --cache
