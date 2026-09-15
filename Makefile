.PHONY: api-build model-user-build

# 生成网关代码。
# goctl api 生成器是扁平的：所有 service 的 logic 都生成在 internal/logic 包下，
# 不会保留 api/ 目录的模块划分。生成后用 mv-logic 把骨架移入模块子包；
# 已实现过的 logic 文件不会被覆盖（goctl 检测到存在即跳过），只需移走新生成的骨架。
# 新增模块时，在 mv-logic 里补一条对应的 mv + sed。
api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero
	$(MAKE) mv-logic

# 将 goctl 生成的扁平 logic 骨架归位到模块子包（user 模块）。
# 注意：目标位置已有同名文件时跳过 mv，防止覆盖已实现的代码——
# goctl 生成骨架是因为扁平位置没有该文件，移过去只会得到空壳。
mv-logic:
	@mkdir -p app/gateway/api/internal/logic/user
	@for f in register_logic login_logic get_user_info_logic upload_avatar_logic get_m_f_a_qrcode_logic bind_m_f_a_logic; do \
		if [ -f app/gateway/api/internal/logic/$$f.go ] && [ ! -f app/gateway/api/internal/logic/user/$$f.go ]; then \
			mv app/gateway/api/internal/logic/$$f.go app/gateway/api/internal/logic/user/; \
			sed -i 's/^package logic$$/package user/' app/gateway/api/internal/logic/user/$$f.go; \
			sed -i 's|^\t"tiktok-go/app/gateway/api/internal/logic"|\tuserlogic "tiktok-go/app/gateway/api/internal/logic/user"|; s|\blog\.New|userlogic.New|g' app/gateway/api/internal/handler/$$(echo $$f | sed 's/_logic//')_handler.go; \
		fi \
	done

# 从 DDL 生成 user model（带 redis 缓存）
model-user-build:
	goctl model mysql ddl -src migration/user.sql -dir app/user/rpc/model -style go_zero --cache
