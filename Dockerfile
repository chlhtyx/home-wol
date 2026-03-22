# ============================================
# 阶段 1: 编译应用
# ============================================
FROM registry.cn-shanghai.aliyuncs.com/xm69/go-build AS build

# 设置工作目录
WORKDIR /build

# 设置 Go 环境变量
ENV GOPROXY=https://goproxy.cn \
    GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 复制依赖文件
COPY go.mod go.sum ./
RUN set -eux && \
    echo "=== 下载 Go 依赖 ===" && \
    go mod download && \
    go mod verify

# 复制源代码
COPY . .

# 编译应用
RUN set -eux && \
    echo "=== 编译应用 ===" && \
    go build -tags musl -ldflags="-w -s" -o main && \
    echo "=== 编译完成 ===" && \
    ls -lh main

# ============================================
# 阶段 2: 运行时镜像
# ============================================
FROM alpine:3.19

# 安装运行时依赖
RUN set -eux && \
    echo "=== 安装依赖 ===" && \
    # 使用阿里云镜像源
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
        ca-certificates \
        tzdata \
        wget \
    && \
    echo "=== 设置时区 ===" && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    \
    echo "=== 清理缓存 ===" && \
    rm -rf /var/cache/apk/* && \
    echo "=== 镜像构建完成 ==="

# 设置工作目录
WORKDIR /home

# 从构建阶段复制编译好的程序
COPY --from=build /build/main .

# 从构建阶段复制静态文件
COPY --from=build /build/static ./static

# 暴露端口
EXPOSE 1323

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:1323/ || exit 1

# 启动应用
ENTRYPOINT ["./main"]
