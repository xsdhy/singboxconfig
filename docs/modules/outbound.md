# Outbound 管理

## 模块职责

Outbound 模块统一管理最终会进入 sing-box `outbounds` 的出站记录，覆盖两类来源：

- `MANUAL`：后台手工维护的静态出站
- `SUBSCRIPTION`：从订阅源解析后落库的缓存出站

这次重构后，旧的“额外出站”不再作为独立数据模型存在，而是收敛为 `Outbound` 表中的 `MANUAL` 记录。

代码入口：

- 实体定义：`entity/device_management.go`
- API 实现：`service/outbound_api.go`
- 缓存刷新：`service/outbound_cache.go`
- 存储抽象：`storage/storage.go`
- 生成拼装：`convert/singbox/outbound.go`

## 数据模型

统一实体为 `entity.Outbound`，核心字段如下：

- `id`：数据库主键，API 更新和删除时使用
- `tag`：全局唯一的 sing-box outbound 标识
- `name`：展示名称
- `description`：备注说明
- `type`：出站协议类型，仅作展示与辅助筛选
- `enabled`：是否启用
- `sort`：排序值
- `visibleDevices`：逗号分隔的设备编码列表，控制设备可见性
- `configJson`：完整的 sing-box outbound JSON
- `source`：来源，取值为 `MANUAL` 或 `SUBSCRIPTION`
- `subscribeName`：所属订阅名称，仅 `SUBSCRIPTION` 时有值
- `lastFetchTime`：订阅缓存最近一次写入时间，仅 `SUBSCRIPTION` 时有意义

与 Inbound 模块类似，真正参与配置生成的是 `configJson` 反序列化后的结果，而不是文档中的展示字段本身。

## 管理接口

统一接口路径为 `/api/outbounds`：

- `GET /api/outbounds`
- `POST /api/outbounds`
- `PUT /api/outbounds/:id`
- `DELETE /api/outbounds/:id`
- `PATCH /api/outbounds/batch-enable`

接口行为说明：

- 新建接口只允许写入 `MANUAL` 记录，服务端会强制覆盖 `source=MANUAL`
- `SUBSCRIPTION` 记录允许查看、删除、批量启停，但不允许直接编辑 JSON
- 列表接口支持按 `source`、`subscribe_name`、`enabled` 和关键词搜索筛选

## 订阅缓存协作

订阅 Outbound 的抓取、缓存和失败降级由 `service/outbound_cache.go` 统一负责，相关接口挂在 `/api/subscribes` 下：

- `PUT /api/subscribes/:name/cache-config`
- `POST /api/subscribes/:name/refresh-outbound`
- `GET /api/subscribes/:name/outbounds`

刷新流程：

1. 拉取订阅原文
2. 解码为 `entity.SingBoxOut`
3. 覆盖写入统一 `Outbound` 表中的 `SUBSCRIPTION` 记录
4. 更新订阅上的缓存时间、状态和错误信息

如果刷新失败，系统不会删除旧缓存，而是保留旧数据给生成流程继续使用。

## 生成逻辑

设备配置生成时，`resolveGenerateOutbounds()` 会先按订阅的缓存状态决定是否刷新，再从统一 Outbound 表中读取当前设备可见且启用的记录。

后续处理分两步：

1. `singbox.GetExtraOutbounds(deviceCode, items)` 负责设备可见性过滤、排序和 `configJson` 反序列化
2. `singbox.GetOutbounds(...)` 负责构造节点分组并固定追加 `direct` 出站

这意味着生成链路已经不再直接发起订阅网络请求，也不再区分“额外出站”和“订阅节点”两条独立主线。

## 设备可见性规则

`visibleDevices` 采用逗号分隔的精确匹配规则：

- 为空：所有设备可见
- 非空：去空白后与当前 `deviceCode` 精确相等才视为可见

订阅来源节点会继承所属订阅的 `visibleDevices`，因此修改订阅可见性会影响后续刷新写入的新缓存记录。

## 当前限制

- 不对 `configJson` 做完整 schema 校验
- `type` 字段不会自动与 `configJson.type` 强制保持一致
- 订阅来源节点的手动删除和启用状态修改，下次刷新时可能被覆盖
- 当不同来源写入相同 `tag` 时，最终以最后一次写入结果为准

## 适合更新本文档的场景

- 增加 Outbound 结构化编辑器
- 扩展更多来源类型，例如远程模板或系统内置模板
- 调整订阅缓存覆盖策略或冲突处理规则
- 为 Outbound 增加更细粒度的引用校验
