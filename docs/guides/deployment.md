# 部署说明

本文档描述当前仓库已经具备的部署方式，以及代码里实际支持的配置入口。

## 部署形态

当前项目支持两类主路径：

- 直接运行 Go 二进制
- 使用仓库内 `Dockerfile` 构建镜像

前端会被打包为单文件并输出到 `cmd/server/index.html`，随后由 Go 二进制通过 `embed` 内嵌并在 `/api/admin` 提供页面。

## 一体化部署原理

构建链路来自 [Makefile](/Users/xsdhy/data/code/go/singboxconfig/Makefile) 和 [vite.config.ts](/Users/xsdhy/data/code/go/singboxconfig/web/vite.config.ts)：

1. `web/` 执行 `npm install && npm run build`
2. Vite 把前端产物输出到 `cmd/server/`
3. `cmd/server/main.go` 使用 `//go:embed index.html` 内嵌页面
4. Go 服务启动后通过 `/api/admin` 提供管理台

## 二进制部署

### 构建

```bash
make build
```

如果只想本机构建后端：

```bash
go build -o singboxconfig ./cmd/server
```

### 运行

PostgreSQL 模式：

```bash
PORT=7391 \
DATABASE_URL='postgres://user:pass@db:5432/singboxconfig?sslmode=disable' \
./singboxconfig
```

Supabase 模式：

```bash
SUPABASE_URL='https://xxx.supabase.co' \
SUPABASE_KEY='service_role_key' \
./singboxconfig
```

## Docker 部署

### 构建镜像

```bash
docker build -t singboxconfig:latest .
```

### 运行容器

PostgreSQL 模式：

```bash
docker run -d \
  --name singboxconfig \
  -p 7391:7391 \
  -e PORT=7391 \
  -e DATABASE_URL='postgres://user:pass@db:5432/singboxconfig?sslmode=disable' \
  singboxconfig:latest
```

## 反向代理

如果需要通过域名访问，建议只暴露一个 HTTP 服务端口，并在代理层做 TLS。

主要路径：

- `/api/admin`：嵌入式管理台
- `/api/auth/login`：登录接口
- `/api/*`：管理接口，需 Bearer Token
- `/open/generate/:device`：公开配置生成接口
- `/open/ruleset/:tag`：公开规则集读取接口

## 存储后端选择

服务启动时按下面优先级选择存储：

1. `SUPABASE_URL` 和 `SUPABASE_KEY` 同时存在：Supabase REST
2. `DATABASE_URL` 存在：PostgreSQL
3. 其他情况：本地 JSON 文件存储

## 生产环境注意事项

### 1. 管理员账户在首次启动时自动初始化

- 首次启动时，系统会自动检查数据库中是否存在管理员配置
- 如果不存在（首次启动），会自动初始化默认管理员账户：用户名 `admin`，密码 `admin`
- 后续启动会复用已初始化的认证配置，不会重复生成
- **重要：首次启动后请立即登录管理台修改默认密码**

### 2. 管理端认证改为 Bearer Token

当前流程：

1. 打开 `/api/admin`
2. 前端调用 `POST /api/auth/login`
3. 登录成功后本地保存 Bearer Token
4. 后续请求统一带 `Authorization: Bearer ...`

这意味着：

- 管理台页面本身不再受 Basic Auth 保护
- 反向代理和安全设备不应依赖浏览器 Basic Auth 行为
- 修改用户名或密码后，旧 token 会立即失效

### 3. 公开生成接口依赖设备 token

`/open/generate/:device` 不走管理端登录体系，而是通过查询参数 `token` 鉴权。

### 4. 必须配置数据库

项目已移除文件存储模式，启动时必须配置以下之一：

- `DATABASE_URL`（PostgreSQL/MySQL）
- `SUPABASE_URL` + `SUPABASE_KEY`

### 5. 管理员密码重置方式

命令行重置：

```bash
./singboxconfig -reset-password 'new-password-123'
```

部署配置重置：

```bash
FORCE_RESET_PASSWORD='new-password-123' ./singboxconfig
```

注意：

- `-reset-password` 执行后直接退出，不启动 HTTP 服务
- `FORCE_RESET_PASSWORD` 只要存在，就会在每次启动时继续覆盖密码
- 使用 `FORCE_RESET_PASSWORD` 后需要手动从部署配置中移除

## 相关文档

- [快速开始](./quickstart.md)
- [配置项说明](../reference/configuration.md)
- [API 接口列表](../reference/api-reference.md)
