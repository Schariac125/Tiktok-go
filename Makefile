.PHONY: api-build model-user-build

api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero

# 从 DDL 生成 user model（带 redis 缓存）
model-user-build:
	goctl model mysql ddl -src migration/user.sql -dir app/user/rpc/model -style go_zero --cache
