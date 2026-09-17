# syntax=docker/dockerfile:1

# 构建 context 必须是项目根目录（go.mod 在那里）：
#   docker build -f app/interaction/rpc/Dockerfile -t tiktok-go/interaction-rpc .

# ---- 构建阶段 ----
FROM golang:1.26-alpine AS builder

WORKDIR /build

# 先只拷依赖清单，源码改动时这一层仍能命中缓存
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# CGO_ENABLED=0 静态编译，运行镜像就不必带 libc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./app/interaction/rpc

# ---- 运行阶段 ----
FROM alpine:3.20

# ca-certificates：本服务会经 video rpc 间接用到 OSS
# tzdata：响应里的时间按本地时区格式化
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 app

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /out/server .
# 打进一份默认配置，部署时挂卷覆盖即可（容器内中间件地址与本地不同）
COPY app/interaction/rpc/etc ./etc

USER app
EXPOSE 8082

ENTRYPOINT ["/app/server"]
CMD ["-f", "etc/interaction.yaml"]
