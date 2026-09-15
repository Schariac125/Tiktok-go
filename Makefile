.PHONY: api-build

api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero