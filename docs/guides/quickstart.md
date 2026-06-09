# 快速开始

本文档对应当前仓库实际实现，适合第一次把项目跑起来时使用。

## 环境要求

- Go `1.25.0` 或兼容版本
- Node.js `18+`
- npm
- 可选：
  - PostgreSQL，用于 `DATABASE_URL` 模式
  - Supabase 项目与 service key，用于 `SUPABASE_URL` + `SUPABASE_KEY` 模式

## 本地启动方式

项目分为前端管理台和后端服务两部分：

- 前端开发模式：Vite 默认端口 `5173`
- 后端服务：Gin 默认端口 `7391`
- 前端开发服务器会把 `/api` 代理到 `http://localhost:7391`

## 1. 启动后端

在仓库根目录执行：

```bash
go run ./cmd/server
```

默认行为：

- 端口默认 `7391`
- 必须配置 `SUPABASE_URL` / `SUPABASE_KEY` 或 `DATABASE_URL`，否则服务无法启动

PostgreSQL：

```bash
PORT=7391 DATABASE_URL='postgres://user:pass@127.0.0.1:5432/singboxconfig?sslmode=disable' go run ./cmd/server
```

Supabase：

```bash
SUPABASE_URL='https://xxx.supabase.co' SUPABASE_KEY='service_role_key' go run ./cmd/server
```

## 2. 启动前端开发服务器

在 `web/` 目录执行：

```bash
npm install
npm run dev
```

访问：

- 前端开发页：`http://localhost:5173`
- 后端 API：`http://localhost:7391`

## 3. 初始化管理员账户

首次启动时：

- 系统会自动检查数据库中是否存在管理员配置
- 如果不存在（首次启动），会自动初始化默认管理员账户
- 默认用户名: `admin`
- 默认密码: `admin`

示例日志：

```text
========================================
首次启动，已初始化管理员账户
用户名: admin
密码: admin
请登录后尽快修改密码
========================================
```

**重要提示：首次启动后请立即登录管理台修改默认密码**

## 4. 登录管理台

当前管理台入口是 `/api/admin`，但页面本身不再依赖浏览器 Basic Auth。

实际流程是：

1. 打开 `http://localhost:7391/api/admin`
2. 输入管理员用户名和密码
3. 前端调用 `POST /api/auth/login`
4. 登录成功后本地保存 Bearer Token
5. 后续所有 `/api/*` 请求自动带 `Authorization: Bearer ...`

如果你使用的是前端开发服务器，则直接访问：

- `http://localhost:5173`

## 5. 首次建议录入的数据

建议按下面顺序初始化：

1. 订阅源
2. 节点分组
3. 规则集
4. Inbound 模板
5. 设备
6. 设备与 Inbound 绑定
7. 可选的 WireGuard / 额外出站 / DNS

## 6. 验证接口是否可用

先登录拿 token：

```bash
curl -X POST http://localhost:7391/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

再用 Bearer Token 请求管理接口：

```bash
curl http://localhost:7391/api/subscribes \
  -H "Authorization: Bearer <access-token>"
```

公开生成接口不依赖管理端登录态，而是按设备 token 鉴权：

```bash
curl "http://localhost:7391/open/generate/default?token=996007"
```

Surge 客户端可使用平行的公开接口：

```bash
curl "http://localhost:7391/open/surge/default?token=996007"
```

Shadowrocket 客户端可使用另一条平行的公开接口：

```bash
curl "http://localhost:7391/open/shadowrocket/default?token=996007"
```

## 常见启动问题

### 登录页能打开，但接口返回 401

检查：

- 登录是否成功返回了 token
- 请求头里是否带了 `Authorization: Bearer ...`
- token 是否已经因为修改用户名/密码而失效

### 忘记管理员密码怎么办

命令行重置：

```bash
go run ./cmd/server -reset-password 'new-password-123'
```

容器启动时临时注入：

```bash
FORCE_RESET_PASSWORD='new-password-123' go run ./cmd/server
```

注意：

- `FORCE_RESET_PASSWORD` 是持续覆盖型参数
- 重置完成后要把它从部署配置里删掉，否则每次重启都会继续覆盖密码

### 前端页面能打开，但请求失败

检查：

- 后端是否运行在 `7391`
- Vite 代理是否生效
- 浏览器控制台里请求是否真的打到了 `/api/*`

### 使用 `DATABASE_URL` 但连接失败

当前入口只接了 PostgreSQL 驱动：

- `DATABASE_URL` 必须是 PostgreSQL 可用连接串
- 虽然依赖里包含 MySQL 驱动，但 `cmd/server/main.go` 当前没有根据环境变量切换到这些驱动

### 服务启动失败提示 "No storage backend configured"

检查：

- 是否配置了 `DATABASE_URL` 或 `SUPABASE_URL` + `SUPABASE_KEY`
- 必须配置其中之一，否则服务无法启动

## 相关文档

- [部署说明](./deployment.md)
- [开发指南](./development.md)
- [前端架构](../frontend/architecture.md)
- [配置项说明](../reference/configuration.md)
