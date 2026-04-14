# 配置项说明

本文档描述当前启动入口实际读取的环境变量与认证相关保留配置。

## 服务启动读取的环境变量

启动逻辑位于 [main.go](/Users/xsdhy/data/code/go/singboxconfig/cmd/server/main.go)。

### `PORT`

用途：

- 指定 HTTP 服务监听端口

默认值：

- `7391`

### `SUPABASE_URL`

用途：

- 指定 Supabase 项目地址

配套变量：

- `SUPABASE_KEY`

### `SUPABASE_KEY`

用途：

- Supabase API Key

### `DATABASE_URL`

用途：

- 指定数据库连接串

当前行为：

- 当前入口固定用 PostgreSQL 驱动打开它

### `DATA_FILE`

用途：

- 指定 JSON 文件存储路径

默认值：

- `data.json`

## 存储选择优先级

启动时按下面顺序决定存储后端：

1. `SUPABASE_URL` + `SUPABASE_KEY`
2. `DATABASE_URL`
3. `DATA_FILE` / 默认 `data.json`

## 管理员初始化相关变量

### `ADMIN_USERNAME`

用途：

- 首次初始化管理员用户名
- 在未初始化状态下执行密码重置时作为默认用户名来源

默认值：

- `admin`

说明：

- 只在认证尚未初始化时使用
- 已初始化后，普通启动不会覆盖现有管理员用户名

### `ADMIN_PASSWORD`

用途：

- 首次初始化管理员密码

说明：

- 若未提供，服务会生成随机强密码并写入启动日志
- 若提供，需满足最小密码长度要求

### `FORCE_RESET_PASSWORD`

用途：

- 启动时强制重置管理员密码

说明：

- 这是持续覆盖型启动参数
- 只要环境变量存在，每次启动都会执行一次重置
- 使用完成后必须从部署配置中移除

## 命令行参数

### `-reset-password`

用途：

- 重置管理员密码后直接退出

说明：

- 不依赖旧密码
- 若认证尚未初始化，会顺带补齐管理员配置

## 管理端认证配置

管理端不再使用浏览器 HTTP Basic Auth，而是：

1. `POST /api/auth/login` 登录
2. 返回 Bearer Token
3. 后续 `/api/*` 请求带 `Authorization: Bearer ...`

认证配置仍保存在保留的全局设置 key 中：

- `auth.username`
- `auth.password_hash`
- `auth.initialized_at`
- `auth.password_changed_at`
- `auth.token_secret`
- `auth.session_version`

说明：

- `auth.*` 是系统保留配置
- `/api/settings` 不会暴露，也不允许修改这些 key
- 配置导入导出默认不会导出或导入这些 key

## 数据库配置

项目必须使用数据库存储，支持以下方式：

### PostgreSQL

当前已在启动入口接通，可以直接通过 `DATABASE_URL` 使用。

### MySQL

当前状态：

- 依赖已存在
- 存储层也具备通用 `DatabaseStorage`
- 但启动入口没有根据变量切换到 `mysql.Open(...)`

## Supabase 模式说明

Supabase 存储通过 PostgREST HTTP API 工作，实现在 [supabase.go](/Users/xsdhy/data/code/go/singboxconfig/storage/supabase.go)。

## 前端开发配置

前端开发代理配置在 [vite.config.ts](/Users/xsdhy/data/code/go/singboxconfig/web/vite.config.ts)：

- `/api` -> `http://localhost:7391`

前端登录态存储在浏览器本地存储，由 `web/src/api/index.ts` 统一给请求追加 Bearer Token。

## 示例配置

### 文件模式

```bash
PORT=7391
DATA_FILE=./data.json
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me-123
```

### PostgreSQL 模式

```bash
PORT=7391
DATABASE_URL=postgres://user:pass@127.0.0.1:5432/singboxconfig?sslmode=disable
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me-123
```

### Supabase 模式

```bash
PORT=7391
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=your-service-role-key
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change-me-123
```

## 相关文档

- [快速开始](../guides/quickstart.md)
- [部署说明](../guides/deployment.md)
- [存储抽象层](../architecture/storage-layer.md)
