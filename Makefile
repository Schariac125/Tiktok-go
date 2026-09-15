.PHONY: api-build model-user-build

# 生成网关代码。
# goctl api 生成器是扁平的：所有 service 的 handler/logic 都生成在
# internal/handler、internal/logic 包下，不保留 api/ 目录的模块划分。
# 生成后执行 mv-logic 将骨架归位到模块子包并修正 routes.go。
api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero
	$(MAKE) mv-logic

# 将 goctl 生成的扁平 handler/logic 骨架归位到模块子包（user 模块），并修正 routes.go 引用。
# 注意：
#   - 目标位置已有同名文件时跳过 mv，防止空骨架覆盖已实现的代码（残留骨架直接删除）
#   - routes.go 头部注明了新模块接入方式
mv-logic:
	@mkdir -p app/gateway/api/internal/handler/user app/gateway/api/internal/logic/user
	@for f in register login get_user_info upload_avatar get_m_f_a_qrcode bind_m_f_a; do \
		if [ -f app/gateway/api/internal/handler/$$f\_handler.go ]; then \
			if [ ! -f app/gateway/api/internal/handler/user/$$f\_handler.go ]; then \
				mv app/gateway/api/internal/handler/$$f\_handler.go app/gateway/api/internal/handler/user/; \
				sed -i 's/^package handler$$/package user/' app/gateway/api/internal/handler/user/$$f\_handler.go; \
				sed -i 's|^\t"tiktok-go/app/gateway/api/internal/logic"|\t"tiktok-go/app/gateway/api/internal/logic/user"|' app/gateway/api/internal/handler/user/$$f\_handler.go; \
				sed -i 's|\blog\.New|user.New|g' app/gateway/api/internal/handler/user/$$f\_handler.go; \
				sed -i 's/^\tfunc \([a-zA-Z]\)/\tfunc \U\1\E/' app/gateway/api/internal/handler/user/$$f\_handler.go; \
			else \
				rm -f app/gateway/api/internal/handler/$$f\_handler.go; \
			fi \
		fi \
	done
	@for f in register_logic login_logic get_user_info_logic upload_avatar_logic get_m_f_a_qrcode_logic bind_m_f_a_logic; do \
		if [ -f app/gateway/api/internal/logic/$$f.go ] && [ ! -f app/gateway/api/internal/logic/user/$$f.go ]; then \
			mv app/gateway/api/internal/logic/$$f.go app/gateway/api/internal/logic/user/; \
			sed -i 's/^package logic$$/package user/' app/gateway/api/internal/logic/user/$$f.go; \
		fi \
	done
	@rm -f app/gateway/api/internal/logic/*_logic.go
	@sed -i 's|^\t"tiktok-go/app/gateway/api/internal/svc"|\tuserhandler "tiktok-go/app/gateway/api/internal/handler/user"\n\t"tiktok-go/app/gateway/api/internal/svc"|' app/gateway/api/internal/handler/routes.go
	@sed -i 's|\bloginHandler(|userhandler.LoginHandler(|; s|\bregisterHandler(|userhandler.RegisterHandler(|; s|\bgetUserInfoHandler(|userhandler.GetUserInfoHandler(|; s|\buploadAvatarHandler(|userhandler.UploadAvatarHandler(|; s|\bgetMFAQrcodeHandler(|userhandler.GetMFAQrcodeHandler(|; s|\bbindMFAHandler(|userhandler.BindMFAHandler(|' app/gateway/api/internal/handler/routes.go

# 从 DDL 生成 user model（带 redis 缓存）
model-user-build:
	goctl model mysql ddl -src migration/user.sql -dir app/user/rpc/model -style go_zero --cache
