# 订阅管理

## 模块职责

订阅管理负责维护远程节点订阅源，并在配置生成时把订阅内容转换为统一 `entity.SingBoxOut` / `entity.Outbound`。这些缓存节点既可进入 sing-box JSON，也可被 Surge / Shadowrocket 输出链路继续渲染。它覆盖两部分能力：

- 管理端 CRUD：增删改查订阅源
- 生成端解析：拉取远程文本、拆分节点 URL、调用协议解码器转换为统一出站

代码入口：

- 实体定义：`entity/subscribe.go`
- 服务接口：`service/service.go`
- 缓存刷新与生成过滤：`service/outbound_cache.go`
- 协议解析：`protocol/*.go`

## 数据模型

后端实体 `entity.Subscribe`：

- `name`：订阅名称，也是后台主键
- `url`：远程订阅地址
- `userAgent`：可选请求头，留空时回退到默认桌面浏览器 UA
- `status`：是否启用；只有启用的订阅才会参与生成

- `visibleDevices`：逗号分隔的设备编码，控制哪些设备可使用当前订阅的节点
- `outboundCacheDuration`：Outbound 缓存时长（分钟），0 表示不缓存
- `outboundLastFetchTime`：最近一次成功刷新缓存的时间
- `outboundLastFetchStatus`：最近一次刷新状态（SUCCESS / FAILED）
- `outboundLastFetchError`：最近一次刷新失败原因

订阅通过缓存机制将解析后的节点持久化到统一 Outbound 表中，避免每次生成配置都重新拉取。

## 管理接口

管理接口挂在 Bearer Token 保护下的 `/api/subscribes`：

- `POST /api/subscribes`：创建订阅
- `PUT /api/subscribes/:name`：更新订阅
- `DELETE /api/subscribes/:name`：删除订阅
- `GET /api/subscribes`：列出订阅

服务层基本是薄封装，直接读写 `storage.SubscribeStorage`。这一层没有额外校验 URL 可达性，也不会在保存时预解析订阅内容。

## 配置生成中的处理流程

生成接口 `/open/generate/:device?token=...`、`/open/surge/:device?token=...` 与 `/open/shadowrocket/:device?token=...` 都会调用 `resolveGenerateOutbounds()`，订阅处理流程如下：

1. 读取全部订阅
2. 收集禁用（`status=false`）或对当前设备不可见的订阅名称集合
3. 对启用且可见的订阅，判断缓存是否过期，过期则刷新
4. 从统一 Outbound 表读取当前设备可见且启用的记录
5. 过滤掉属于禁用/不可见订阅的 `SUBSCRIPTION` 来源 Outbound
6. 将最终 Outbound 列表交给节点分组构造

**订阅禁用的完整效果**：当某个订阅被禁用后，不仅不会触发刷新，其已缓存的 Outbound 也会在生成时被过滤掉，不会出现在任一输出格式的最终配置中。

失败行为是”单订阅降级、继续整体生成”：

- 拉取失败：记录日志，保留旧缓存供生成继续使用
- 内容解析失败：记录日志，跳过该订阅
- 单个节点不支持或解码失败：跳过该节点

这意味着只要还有其他订阅或手动 Outbound 可用，整个设备配置仍可继续生成。

## 请求行为

HTTP 拉取逻辑位于 `service/outbound_cache.go`：

- 超时时间：60 秒
- 使用独立的超时 context，避免受上层 HTTP 请求取消影响
- 默认 UA：桌面 Chrome UA
- 非 200 状态码视为失败

`userAgent` 仅影响订阅拉取请求，不会进入最终 sing-box 配置。

## 与其他模块的关系

- 与“协议解码器”模块强耦合：节点 URL 最终由 `protocol` 包转换
- 与“节点分组”模块耦合：解析出来的节点标签会继续参与分组筛选
- 与“额外出站”模块并列：两者共同组成候选 `outbounds`

## 当前支持范围与限制

代码现状需要明确区分“订阅解析已接入”和“某个客户端输出格式是否支持”：

- 已接入订阅解析链路：`ss`、`ssr`、`trojan`、`vmess`、`vless`
- sing-box 输出会保留这些统一 Outbound
- Surge 输出不展开订阅节点：每个订阅源输出为携带 `policy-path=<订阅地址>` 的策略组，由 Surge 客户端自行拉取订阅（详见 [requirements/implemented/surge-subscription-policy-path.md](../requirements/implemented/surge-subscription-policy-path.md)）；手工节点中 Surge 不导出 SSR / VLESS，会跳过并记录 warning
- Shadowrocket 不支持 `policy-path`，输出仍展开订阅缓存节点，会导出 SSR / VLESS，因此能覆盖比 Surge 手工节点更完整的协议

另外还有几个实现边界：

- 不做重复节点去重，多个订阅可产生同名标签
- 不支持订阅级代理、重试策略、ETag/If-None-Match 等优化

## 适合更新本文档的场景

以下变更需要同步更新本文档：

- 新增或删除支持的订阅协议
- 修改订阅拉取策略、超时或 UA 逻辑
- 增加节点缓存、去重或预解析能力
- 调整订阅接口路径或校验规则
