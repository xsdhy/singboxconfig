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

## 存储选择优先级

启动时按下面顺序决定存储后端：

1. `SUPABASE_URL` + `SUPABASE_KEY`
2. `DATABASE_URL`

如果都未配置，服务将无法启动。

## 管理员初始化相关变量

### 首次启动自动初始化

系统在首次启动时会自动检查数据库中是否存在 auth 配置：

- 如果不存在（首次启动），自动初始化默认管理员账户：
  - 用户名: `admin`
  - 密码: `admin`
- 如果已存在，使用已保存的管理员配置

**重要提示：**

- 首次启动后请立即登录管理台修改默认密码
- 后续启动不会重置已保存的管理员配置

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

## 全局设置项（key/value）

通过 `/api/settings` 管理的普通全局设置（区别于上面的 `auth.*` 保留 key），常用项：

### `dns_config`

- 保存 sing-box 标准 DNS JSON，前端「DNS」页独立维护
- 生成 sing-box 配置时读取；未配置时回退内置默认 DNS

### `system_host`

- 含义：本服务对外可访问的基础地址，如 `https://config.example.com`
- 用途：生成整份配置时，把有效且规则条数 ≥ 3 的 local / inline 规则集由“展开/内联”改为指向本服务规则集 open 接口（`/open/rules/:tag?software=...&device=...&token=...`）的远程 URL 引用（条数少于 3 时仍直接展开/内联）
- 取值约束：
  - 必须是合法的 `http` / `https` 绝对地址；保存时会校验，非法值直接拒绝（返回 400）
  - 读取与保存都会去掉首尾空白与尾部斜杠
  - 留空表示未配置：生成时回退到原展开/内联行为，remote 规则集行为不变
- 编辑入口：前端「Global Settings」页提供独立的「系统 Host」输入框，也可通过通用 key/value 设置项编辑
- 可访问性要求：客户端必须能访问 `system_host`；如走反向代理需保证 HTTPS 与公网/内网可达
- 注意：生成的规则集 URL 携带设备 token，设备 token 轮换后旧整份配置中的规则集 URL 会失效，需重新拉取整份配置

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

### PostgreSQL 模式

```bash
PORT=7391
DATABASE_URL=postgres://user:pass@127.0.0.1:5432/singboxconfig?sslmode=disable
```

### Supabase 模式

```bash
PORT=7391
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=your-service-role-key
```

## 相关文档

- [快速开始](../guides/quickstart.md)
- [部署说明](../guides/deployment.md)
- [存储抽象层](../architecture/storage-layer.md)
