.PHONY: help clean web build build-local docker-build docker-dev docker-up docker-down dev dev-web dev-backend sip test lint

# 默认目标
help:
	@echo "SingBox Config - 构建命令"
	@echo ""
	@echo "构建命令："
	@echo "  make web          - 构建前端 (React SPA -> cmd/server/index.html)"
	@echo "  make build        - 构建 Linux amd64 二进制 (包含前端)"
	@echo ""
	@echo "其他命令："
	@echo "  make test         - 运行测试"
	@echo "  make lint         - 运行代码检查"
	@echo "  make clean        - 清理构建产物"
	@echo "  make sip          - 构建并部署到 sip 服务器"

# 清理构建产物
clean:
	@echo "清理构建产物..."
	rm -f singboxconfig
	rm -f cmd/server/index.html
	rm -rf tmp/
	cd web && rm -rf dist node_modules/.vite


# ============================================
# 构建命令
# ============================================

# 构建前端 (React SPA 打包为单文件 HTML)
web:
	@echo "构建前端..."
	cd web && pnpm install && pnpm run build
	@echo "前端构建完成: cmd/server/index.html"

# 构建 Linux amd64 二进制 (用于生产部署)
build: web
	@echo "构建 Linux amd64 二进制..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o singboxconfig ./cmd/server
	@echo "构建完成: ./singboxconfig"


# ============================================
# Docker 命令
# ============================================

# 构建生产 Docker 镜像
docker-build:
	@echo "构建生产 Docker 镜像..."
	docker build -t singboxconfig:latest .
	@echo "镜像构建完成: singboxconfig:latest"



# ============================================
# 测试和检查
# ============================================

# 运行测试
test:
	@echo "运行后端测试..."
	go test -v ./...
	@echo "运行前端测试..."
	cd web && npm test

# 运行代码检查
lint:
	@echo "运行 Go 代码检查..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "请先安装 golangci-lint"; exit 1; }
	golangci-lint run ./...
	@echo "运行前端代码检查..."
	cd web && npm run lint || echo "前端 lint 命令未配置"

# ============================================
# 部署命令
# ============================================

# 部署到服务器
sip: build
	@echo "部署到服务器..."
	scp ./singboxconfig sip:/data/docker/singboxconfig/singboxconfig.new
	ssh sip "/data/docker/singboxconfig/update.sh"
	@echo "部署完成"
