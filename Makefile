.PHONY: api-build model-user-build

# 生成网关代码。
# goctl 会按 api 文件名(user.api)重新生成 etc/user.yaml 等模板文件，
# 与我们维护的 gateway.yaml/gateway.go 冲突，生成后统一清理。
api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero
	rm -f app/gateway/api/etc/user.yaml
	rm -f app/gateway/api/user.go

# 从 DDL 生成 user model（带 redis 缓存）
model-user-build:
	goctl model mysql ddl -src migration/user.sql -dir app/user/rpc/model -style go_zero --cache
