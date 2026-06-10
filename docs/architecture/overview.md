# 项目概述

## 项目定位

SingBox Config 是一个面向自建代理环境的配置管理系统，目标不是运行代理本身，而是集中维护 sing-box 所需的配置数据，并按设备动态生成最终 JSON。项目主要解决三类问题：

- 订阅源、节点分组、规则集、DNS、入站、WireGuard、额外出站等配置分散且难以统一维护
- 不同设备需要不同的 token、入站组合、WireGuard 参数与可见出站
- 配置生成逻辑容易沉淀为硬编码，难以迁移、审计和扩展

系统因此采用“管理数据 + 生成配置”的设计。后台负责维护结构化数据，生成接口在请求时按设备实时拼装 sing-box 配置。

## 系统边界

项目包含三个核心子系统：

- 后端服务：提供 REST API、开放配置生成接口、配置导入导出能力
- 存储抽象层：屏蔽内存 JSON、GORM 数据库、Supabase REST 三种后端差异
- 前端控制台：单页管理界面，构建后输出到 `cmd/server/index.html` 并嵌入 Go 二进制

项目不负责：

- 运行 sing-box 进程
- 校验生成结果是否完全符合 sing-box 每个版本的最新 schema
- 托管订阅内容，只在生成时拉取订阅 URL

## 整体架构

```text
React SPA
  -> /api/*（Bearer Token 管理接口）
     -> service.Service
        -> storage.Storage
           -> DatabaseStorage | SupabaseStorage

设备 / 客户端
  -> /open/generate/:device?token=...
     -> service.Generated
        -> 存储读取 + 默认值回退
        -> convert/singbox
        -> 输出 sing-box JSON
  -> /open/surge/:device?token=...
     -> service.SurgeGenerated
        -> 复用设备鉴权 + Outbound/规则集/分组读取
        -> convert/surge
        -> 输出 Surge 文本
  -> /open/shadowrocket/:device?token=...
     -> service.ShadowrocketGenerated
        -> 复用设备鉴权 + Outbound/规则集/分组读取
        -> convert/shadowrocket
        -> 输出 Shadowrocket 文本
```

管理面和生成面是分开的：

- `/api/admin` 返回嵌入式管理台页面
- `/api/auth/login` 是公开登录入口
- 其他受保护的 `/api/*` 走 Bearer Token
- `/open/*` 面向设备使用，不依赖后台登录态，改为设备编码 + token 认证

## 核心数据模型

`entity/` 中定义了整个系统的数据边界，主要实体包括：

- `Subscribe`：远程订阅源，控制节点抓取地址、UA、启用状态
- `NodeGroup`：订阅节点的筛选与聚合规则，生成 selector/urltest 出站组
- `RuleSet`：路由规则集定义，支持 inline/local 与 remote 两类
- `Device`：生成配置的目标设备，包含 token、启用状态、WireGuard 客户端参数
- `Inbound` 与 `DeviceInbound`：入站模板及设备绑定关系
- `WireGuard` 与 `WireGuardPeer`：可复用的 WireGuard endpoint 模板
- `ExtraOutbound`：订阅之外的静态出站模板
- `SingBoxConfig`：最终输出给客户端的 sing-box 根配置

这些实体在服务层与存储层之间保持稳定，具体后端存储模型由 `storage/models.go` 和 `storage/supabase.go` 负责映射。

## 运行时主流程

### 1. 服务启动

`cmd/server/main.go` 的启动顺序如下：

1. 按环境变量选择存储实现
2. 创建 `service.Service`
3. 注册 Gin 路由
4. 监听 `PORT`，默认 `7391`

当前优先级为：

1. `SUPABASE_URL` + `SUPABASE_KEY` -> `SupabaseStorage`
2. `DATABASE_URL` -> `DatabaseStorage`
3. 否则 -> 启动失败，要求配置数据库
2. `DATABASE_URL` -> `DatabaseStorage`
3. 否则回退到 `MemoryStorage`，数据文件默认 `data.json`

### 2. 后台管理

前端控制台通过 `/api/*` 调用 CRUD 接口，服务层只做轻量校验和排序，然后把实体写入存储层。此处采用“薄服务层”设计：

- 参数绑定与路径参数一致性校验在 `service/`
- 持久化细节放在 `storage/`
- sing-box 结构拼装放在 `convert/singbox/`

### 3. 配置生成

`GET /open/generate/:device?token=...` 是 sing-box JSON 输出入口，处理顺序为：

1. 解析设备并校验 token
2. 读取 DNS 全局配置
3. 解析设备绑定的入站模板
4. 按设备绑定生成 WireGuard endpoint
5. 读取额外出站
6. 拉取启用的订阅并解码节点
7. 根据节点分组规则构造 selector/urltest
8. 组装 route、experimental、inbounds、outbounds
9. 返回 `entity.SingBoxConfig`

`GET /open/surge/:device?token=...` 是 Surge 文本输出入口。它复用相同的数据层能力，并将 Shadowsocks、Trojan、VMess、HTTP(HTTPS) 节点、由 WireGuard endpoint 转换的 `wireguard` 代理、节点分组和规则集渲染为 Surge 的 `[General]`、`[Proxy]`、`[Proxy Group]`、`[Rule]`、`[WireGuard]` 分段文本。

`GET /open/shadowrocket/:device?token=...` 是 Shadowrocket 文本输出入口。它复用相同的数据层能力，并将 Shadowsocks、ShadowsocksR、Trojan、VMess、VLESS 节点以及 best-effort Hysteria2 / TUIC 节点、节点分组和规则集渲染为 Shadowrocket 的 `[General]`、`[Proxy]`、`[Proxy Group]`、`[Rule]` 分段文本。

## 兼容性设计

项目保留了一层“空存储回退默认配置”的兼容逻辑，集中在 `service/generated.go` 和 `convert/singbox/device_management_defaults.go`：

- 当设备列表为空时，生成接口回退到内置默认设备
- 当设备入站绑定为空时，回退到默认 Inbound 组合
- 当 Inbound 列表为空时，回退到默认 TUN/HTTP/SOCKS/Mixed 模板
- 当 DNS 未配置或 JSON 非法时，回退到默认 DNS
- 当额外出站为空时，回退到空列表
- WireGuard 默认模板当前返回 `nil`，表示旧逻辑中没有启用默认模板

这套机制的目的不是长期承载业务配置，而是在空数据状态下保证系统仍可生成可用配置，兼容历史使用方式。

## 模块协作关系

项目内的依赖方向较清晰：

- `cmd/server` 依赖 `service` 和 `storage`
- `service` 依赖 `storage`、`entity`、`convert/singbox`、`transfer`
- `convert/singbox` 依赖 `entity` 和 `protocol`
- `storage` 依赖 `entity`
- `web` 独立构建，通过 HTTP API 与后端交互

依赖方向基本保持单向，没有前端代码反向影响后端生成逻辑，也没有存储层反向依赖服务层。

## 设计特点

- 数据驱动：设备、DNS、入站、WireGuard、额外出站都已从硬编码迁移为存储驱动
- 运行时生成：不预生成静态配置文件，而是在请求时按设备即时拼装
- 多后端存储：业务逻辑只依赖 `storage.Storage`
- 前后端一体部署：前端静态资源嵌入后端二进制，部署简单
- 兼容旧行为：保留默认值回退，避免空库时直接失效

## 当前实现的关键限制

- `cmd/server/main.go` 在 `DATABASE_URL` 分支中固定使用 PostgreSQL 驱动；虽然 `go.mod` 已引入 MySQL 依赖，当前启动入口并未暴露对应切换逻辑
- 管理端当前是单管理员模型，不包含多用户和 RBAC
- 前端是单页菜单切换，不使用前端路由系统
- 订阅协议解码当前启用 `ss`、`ssr`、`trojan`、`vmess`、`vless`；Hysteria2 / TUIC 可通过手工 Outbound JSON 进入 Shadowrocket best-effort 输出

这些限制不影响项目主链路，但在后续架构演进时需要优先处理。
