# 常见问题

## 服务启动后访问 `/` 返回 404

原因：

- 当前后端没有把前端挂在根路径

正确入口：

- 管理台：`/api/admin`
- sing-box 公开生成接口：`/open/generate/:device`
- Surge 公开生成接口：`/open/surge/:device`

## 为什么管理台能打开，但接口提示未认证

原因：

- 管理台页面本身是公开的
- 真正的管理接口 `/api/*` 需要 Bearer Token

检查：

- 是否已经通过 `/api/auth/login` 登录
- 请求头里是否带了 `Authorization: Bearer ...`
- token 是否已经过期或被账号修改操作作废

## 为什么前端页面打开了，但接口还是失败

检查：

- 后端是否运行在 `7391`
- 前端是否通过 Vite 开发服务器访问
- 浏览器本地存储里是否还有有效 token
- 请求是否真的走到了 `/api/*`

## 忘记管理员密码怎么办

可选方式：

- 运行 `./singboxconfig -reset-password 'new-password-123'`
- 或临时设置 `FORCE_RESET_PASSWORD='new-password-123'` 再启动服务

注意：

- `-reset-password` 会重置后直接退出
- `FORCE_RESET_PASSWORD` 会在每次启动时继续覆盖密码，使用完要删除这个环境变量

## 用户名也忘了怎么办

当前仍是单管理员模型：

- 如果还能登录，直接在“账号设置”里修改用户名
- 如果已经无法登录，先通过重置密码恢复访问，再用当前用户名进入后台修改
- 如果你不确定历史用户名是什么，可以直接查看存储中的 `auth.username`

## 为什么修改用户名或密码后，旧登录态立刻失效

原因：

- 后端会轮换 `auth.session_version`
- 所有旧 token 都会立即失效

这是预期行为，用来避免账号修改后旧会话继续可用。

## 为什么 `DATABASE_URL` 配了却连不上 MySQL

原因：

- 当前启动入口只使用 PostgreSQL 驱动打开 `DATABASE_URL`

结论：

- 现在可直接用的是 PostgreSQL
- MySQL 还不是当前入口的可运行模式

## 为什么没有任何数据时也能生成配置

原因：

- 生成链路里保留了若干默认回退逻辑

例如：

- 默认设备
- 默认 DNS
- 默认 Inbound 绑定
- 默认额外出站

这些逻辑主要用于兼容历史行为，不代表后台已经完成初始化。

## 为什么我创建了设备后，`default` 设备访问不到了

原因：

- 只有在设备存储完全为空时，生成接口才会回退到内置默认设备
- 一旦设备表已有数据，就只按真实设备列表查找

## 为什么规则集限制设备时有误匹配

原因：

- 当前 `AbleDevices` 的判断使用 `strings.Contains`

## 为什么 DNS 配置保存后生成接口仍像默认配置

可能原因：

- `dns_config` 保存的 JSON 非法

当前行为：

- 生成链路在解析失败时会记录警告并回退到默认 DNS，而不是直接报错

## 为什么 Surge 配置里缺少部分节点

可能原因：

- Surge 输出第一版只导出 Shadowsocks、Trojan 和 best-effort VMess
- VLESS、Hysteria、Hysteria2、TUIC 等协议会被跳过并记录 warning
- 某个节点缺少必要字段，例如 `server`、`server_port`、密码或 UUID

这类问题只影响单个节点，不会中断整个 Surge 配置生成。

## 为什么订阅里配置了 SSR 但生成结果没有节点

原因：

- `protocol/ssr.go` 虽然存在解码器实现
- 但当前订阅解析映射没有注册 `ssr`

## 数据持久化要求

项目已移除文件存储模式，必须使用数据库：

- PostgreSQL（推荐）
- MySQL
- Supabase

配置方式：

- 设置 `DATABASE_URL` 环境变量
- 或设置 `SUPABASE_URL` + `SUPABASE_KEY`

生产环境建议使用 PostgreSQL 或 Supabase 以确保数据安全。

## 相关文档

- [快速开始](../guides/quickstart.md)
- [配置项说明](./configuration.md)
- [API 接口列表](./api-reference.md)
