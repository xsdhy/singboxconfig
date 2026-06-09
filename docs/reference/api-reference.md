# API 接口列表

本文档基于当前 [main.go](/Users/xsdhy/data/code/go/singboxconfig/cmd/server/main.go) 实际注册的路由整理。

## 认证方式

接口分两类：

- 公开接口：`/open/*`
- 管理接口：`/api/*`

### 管理接口认证

管理接口使用 Bearer Token。

当前行为：

- `POST /api/auth/login` 使用用户名密码登录
- 登录成功后返回 `access_token`
- 受保护的 `/api/*` 请求需要带 `Authorization: Bearer <access_token>`
- 首次启动会初始化单管理员账户
- 凭据持久化存储在 `auth.*` 保留配置中
- 可通过 `POST /api/auth/change-credentials` 修改用户名和密码
- 可通过 `-reset-password` 或 `FORCE_RESET_PASSWORD` 执行运维重置

### 公开接口认证

- `/open/generate/:device`：通过查询参数 `token` 鉴权
- `/open/surge/:device`：通过查询参数 `token` 鉴权
- `/open/shadowrocket/:device`：通过查询参数 `token` 鉴权
- `/open/ruleset/:tag`：无需额外鉴权

## 返回风格

当前后端没有统一响应包装，大多数接口直接返回：

- 实体对象
- 实体数组
- 原始 `map`
- 或 `{ "message": "Deleted successfully" }`
- 失败时返回 `{ "error": "..." }`

## 公开接口

### `GET /open/generate/:device`

用途：

- 为设备生成 sing-box 配置

查询参数：

- `token`：设备 token

状态码：

- `200`：返回完整 sing-box JSON
- `401`：token 不匹配
- `403`：设备被禁用
- `404`：设备不存在

### `GET /open/surge/:device`

用途：

- 为设备生成 Surge 配置文本

查询参数：

- `token`：设备 token

说明：

- 复用 `/open/generate/:device` 的设备解析、启用状态和 token 鉴权
- 复用订阅缓存刷新、Outbound 设备可见性过滤、节点分组筛选和规则集过滤
- `Content-Type` 为 `text/plain`
- Shadowsocks / Trojan 会完整映射，VMess 做 best-effort 映射
- VLESS / Hysteria / Hysteria2 / TUIC 等 Surge 第一版未导出的协议会跳过并记录 warning

状态码：

- `200`：返回 Surge INI 风格配置文本
- `401`：token 不匹配
- `403`：设备被禁用
- `404`：设备不存在

### `GET /open/shadowrocket/:device`

用途：

- 为设备生成 Shadowrocket 配置文本

查询参数：

- `token`：设备 token

说明：

- 复用 `/open/generate/:device` 的设备解析、启用状态和 token 鉴权
- 复用订阅缓存刷新、Outbound 设备可见性过滤、节点分组筛选和规则集过滤
- `Content-Type` 为 `text/plain`
- Shadowsocks / ShadowsocksR / Trojan / VMess / VLESS 会完整映射
- Hysteria2 / TUIC 做 best-effort 映射；Hysteria v1 等暂不导出的协议会跳过并记录 warning

状态码：

- `200`：返回 Shadowrocket INI 风格配置文本
- `401`：token 不匹配
- `403`：设备被禁用
- `404`：设备不存在

### `GET /open/ruleset/:tag`

用途：

- 返回本地规则集内容

说明：

- 只会从当前保存的规则集里按 `tag` 查找
- 远程规则集不会在这里转发其远程 URL 内容

## 管理接口

### 认证管理

- `POST /api/auth/login`
- `GET /api/auth/me`
- `POST /api/auth/change-credentials`

`POST /api/auth/login` 请求体：

```json
{
  "username": "admin",
  "password": "your-password"
}
```

成功响应包含：

- `access_token`
- `token_type`
- `expires_at`
- `username`

`GET /api/auth/me` 返回：

- `username`
- `initialized_at`
- `password_changed_at`（存在时返回）

`POST /api/auth/change-credentials` 请求体：

```json
{
  "old_password": "current-password",
  "new_username": "ops-admin",
  "new_password": "new-password-456"
}
```

说明：

- `old_password` 必填
- `new_username` 和 `new_password` 至少提供一项
- 成功响应会返回新的 `access_token`

### 订阅管理

- `GET /api/subscribes`
- `POST /api/subscribes`
- `PUT /api/subscribes/:name`
- `DELETE /api/subscribes/:name`
- `PUT /api/subscribes/:name/cache-config`
- `POST /api/subscribes/:name/refresh-outbound`
- `GET /api/subscribes/:name/outbounds`

请求体字段：

- `name`
- `url`
- `userAgent`
- `status`
- `visibleDevices`
- `outboundCacheDuration`
- `outboundLastFetchTime`
- `outboundLastFetchStatus`
- `outboundLastFetchError`

补充说明：

- `PUT /api/subscribes/:name/cache-config` 目前只更新 `outboundCacheDuration`
- `POST /api/subscribes/:name/refresh-outbound` 会立即拉取订阅并刷新统一 Outbound 缓存
- `GET /api/subscribes/:name/outbounds` 返回分页列表和 `subscribeCacheInfo` 摘要

### 节点分组

- `GET /api/node-groups`
- `POST /api/node-groups`
- `PUT /api/node-groups/:tag`
- `DELETE /api/node-groups/:tag`

请求体字段：

- `tag`
- `name`
- `groupType`
- `testURL`
- `include`
- `exclude`

### 规则集

- `GET /api/rule-sets`
- `POST /api/rule-sets`
- `PUT /api/rule-sets/:tag`
- `DELETE /api/rule-sets/:tag`

请求体字段：

- `tag`
- `name`
- `ruleSetType`
- `format`
- `content`
- `url`
- `outbound`
- `downloadDetour`
- `ableDevices`
- `sort`

### 全局设置

- `GET /api/settings`
- `POST /api/settings`
- `PUT /api/settings/:key`
- `DELETE /api/settings/:key`
- `GET /api/settings/:key`
- `GET /api/settings/key/:key`

请求体字段：

- `key`
- `value`

注意：

- `GET /api/settings` 当前返回的是 `map[string]string`
- `GET /api/settings/:key` 和 `GET /api/settings/key/:key` 当前行为相同
- `auth.*` 是保留 key，这组接口不会返回，也不允许读写删

### 配置导入导出

- `GET /api/config-transfer/export`
- `POST /api/config-transfer/import`

说明：

- 导出返回 JSON 文件下载
- 导入使用 `multipart/form-data`，字段名是 `file`
- `auth.*` 默认不会导出，也不会在导入时写入

### 设备管理

- `GET /api/devices`
- `POST /api/devices`
- `GET /api/devices/:code`
- `PUT /api/devices/:code`
- `DELETE /api/devices/:code`
- `GET /api/devices/:code/inbounds`
- `PUT /api/devices/:code/inbounds`

设备字段：

- `code`
- `name`
- `description`
- `token`
- `enabled`
- `sort`
- `wireGuardTag`
- `wireGuardClientAddr`
- `wireGuardClientKey`

绑定字段：

- `deviceCode`
- `inboundTag`
- `sort`

### Inbound 模板

- `GET /api/inbounds`
- `POST /api/inbounds`
- `GET /api/inbounds/:tag`
- `PUT /api/inbounds/:tag`
- `DELETE /api/inbounds/:tag`

字段：

- `tag`
- `name`
- `description`
- `type`
- `enabled`
- `sort`
- `configJson`

### WireGuard

- `GET /api/wire-guards`
- `POST /api/wire-guards`
- `GET /api/wire-guards/:tag`
- `PUT /api/wire-guards/:tag`
- `DELETE /api/wire-guards/:tag`
- `GET /api/wire-guards/:tag/peers`
- `POST /api/wire-guards/:tag/peers`
- `PUT /api/wire-guards/:tag/peers/:id`
- `DELETE /api/wire-guards/:tag/peers/:id`

WireGuard 字段：

- `tag`
- `name`
- `description`
- `enabled`
- `sort`
- `endpointTag`
- `mtu`

Peer 字段：

- `id`
- `wireGuardTag`
- `sort`
- `address`
- `port`
- `publicKey`
- `preSharedKey`
- `allowedIps`
- `persistentKeepaliveInterval`
- `enabled`

### Outbound 管理

- `GET /api/outbounds`
- `POST /api/outbounds`
- `PUT /api/outbounds/:id`
- `DELETE /api/outbounds/:id`
- `PATCH /api/outbounds/batch-enable`

列表查询参数：

- `source`：`MANUAL` 或 `SUBSCRIPTION`
- `subscribe_name`：按订阅名称筛选
- `enabled`：`true` 或 `false`
- `search`：按 `tag`、`name` 模糊搜索
- `page`：页码，默认 `1`
- `limit`：每页数量，默认 `20`

字段：

- `id`
- `tag`
- `name`
- `description`
- `type`
- `enabled`
- `sort`
- `visibleDevices`
- `configJson`
- `source`
- `subscribeName`
- `lastFetchTime`
- `createdAt`
- `updatedAt`

补充说明：

- `POST /api/outbounds` 只允许创建手工节点，服务端会强制写入 `source=MANUAL`
- `PUT /api/outbounds/:id` 只允许编辑 `MANUAL` 记录；订阅缓存节点是只读的
- `PATCH /api/outbounds/batch-enable` 请求体为 `{ "ids": [1,2], "enabled": true }`

### 管理台页面

- `GET /api/admin`

说明：

- 返回嵌入式 `index.html`
- 页面本身不要求已登录
- 页面内的接口请求仍通过 Bearer Token 访问受保护的 `/api/*`

## 请求示例

### 登录并获取 Token

```bash
curl -X POST http://localhost:7391/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

### 读取当前管理员信息

```bash
curl http://localhost:7391/api/auth/me \
  -H "Authorization: Bearer <access-token>"
```

### 修改管理员用户名和密码

```bash
curl -X POST http://localhost:7391/api/auth/change-credentials \
  -H "Authorization: Bearer <access-token>" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"old-password","new_username":"ops-admin","new_password":"new-password"}'
```

### 生成设备配置

```bash
curl -X GET "http://localhost:7391/open/generate/phone?token=device-token-123"
```

**成功响应 (200)**：

```json
{
  "inbounds": [...],
  "outbounds": [...],
  "route": {...},
  "dns": {...},
  "experimental": {...}
}
```

**认证失败 (401)**：

```json
{
  "error": "Invalid token"
}
```

### 生成 Surge 设备配置

```bash
curl -X GET "http://localhost:7391/open/surge/phone?token=device-token-123"
```

**成功响应 (200)**：

```ini
[General]
dns-server = system

[Proxy]
Proxy-SS = ss, 1.2.3.4, 8000, encrypt-method=chacha20-ietf-poly1305, password=abcd1234

[Proxy Group]
general = select, Proxy-SS

[Rule]
FINAL,general
```

### 生成 Shadowrocket 设备配置

```bash
curl -X GET "http://localhost:7391/open/shadowrocket/phone?token=device-token-123"
```

**成功响应 (200)**：

```ini
[General]
dns-server = system

[Proxy]
Proxy-SS = ss, 1.2.3.4, 8000, encrypt-method=chacha20-ietf-poly1305, password=abcd1234

[Proxy Group]
general = select, Proxy-SS

[Rule]
FINAL,general
```

### 创建订阅

```bash
curl -X POST "http://localhost:7391/api/subscribes" \
  -H "Authorization: Basic eHNkaHk6eHNkaHkxMjM0NTY=" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-sub",
    "url": "https://example.com/subscribe?token=xxx",
    "userAgent": "ClashX/1.0",
    "status": 1
  }'
```

**成功响应 (200)**：

```json
{
  "name": "my-sub",
  "url": "https://example.com/subscribe?token=xxx",
  "userAgent": "ClashX/1.0",
  "status": 1
}
```

### 列出设备

```bash
curl -X GET "http://localhost:7391/api/devices" \
  -H "Authorization: Basic eHNkaHk6eHNkaHkxMjM0NTY="
```

**成功响应 (200)**：

```json
[
  {
    "code": "phone",
    "name": "My Phone",
    "token": "device-token-123",
    "enabled": true,
    "description": "Primary mobile device"
  },
  {
    "code": "tv",
    "name": "Living Room TV",
    "token": "device-token-456",
    "enabled": true,
    "description": ""
  }
]
```

### 配置导出

```bash
curl -X GET "http://localhost:7391/api/config-transfer/export" \
  -H "Authorization: Basic eHNkaHk6eHNkaHkxMjM0NTY=" \
  --output config.json
```

**成功响应 (200)**：返回完整配置数据 JSON 文件

## 常见错误码

- `400`：请求体非法、路径参数和 body 不一致、文件扩展名不合法等
- `401`：管理端未认证或公开生成接口 token 错误
- `403`：设备被禁用
- `404`：资源不存在
- `500`：存储或内部处理异常

## 当前接口与规划的差异

- 根路径 `/` 没有提供前端页面
- 没有专门的登录或会话接口

## 相关文档

- [配置项说明](./configuration.md)
- [sing-box 配置结构](./singbox-config-schema.md)
- [前端 API 客户端](../frontend/api-client.md)
