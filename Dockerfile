# ============================================
# 阶段 1: 前端构建 (Node.js)
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/web

# 复制前端依赖文件
COPY web/package*.json ./

# 安装依赖（包括 devDependencies，构建时需要）
RUN npm ci

# 复制前端源码
COPY web/ ./

# 构建前端 (输出到 ../cmd/server/index.html)
RUN npm run build


# ============================================
# 阶段 2: 后端构建 (Go)
# ============================================
FROM golang:1.25-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git

WORKDIR /build

# 复制 Go 依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY . .

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /build/cmd/server/index.html ./cmd/server/index.html

# 构建后端 (使用 TARGETARCH 自动适配多架构)
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath \
    -o singboxconfig ./cmd/server


# ============================================
# 阶段 3: 运行时镜像
# ============================================
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /build/singboxconfig .

# 创建数据目录
RUN mkdir -p /app/data && chown -R app:app /app

# 切换到非 root 用户
USER app

# 暴露端口
EXPOSE 7391

# 设置默认环境变量
ENV PORT=7391

# 启动服务
CMD ["./singboxconfig"]
