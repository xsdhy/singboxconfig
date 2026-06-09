# 配置生成流程

## 总体说明

配置生成是本项目最核心的能力，主入口在 `service/generated.go` 的 `Generated` 方法，对应接口：

```text
GET /open/generate/:device?token=...
```

它的职责是根据设备身份、后台维护的数据和一组默认兜底规则，实时生成完整的 sing-box JSON。

当前还提供一条平行输出链路：

```text
GET /open/surge/:device?token=...
```

Surge 输出复用同一套设备解析、token 鉴权、DNS 读取、订阅缓存刷新、Outbound 可见性过滤、节点分组筛选和规则集过滤，最后由 `convert/surge` 渲染为 INI 风格文本。它不导出 Inbound 与 WireGuard endpoint。

这个流程不是简单地“读数据库后原样返回”，而是一个逐步组装过程：

- 先确认目标设备和权限
- 再分别拼装 DNS、入站、endpoint、出站、路由、experimental
- 最后输出 `entity.SingBoxConfig`

## 主流程分解

### 1. 解析设备与鉴权

`Generated` 首先读取：

- 路径参数 `:device`
- 查询参数 `token`

然后调用 `resolveGenerateDevice(deviceCode)`。

设备解析规则是：

1. 先读取 `storage.ListDevices()`
2. 如果存储中设备列表为空，则回退到 `singbox.GetDefaultDevices()`
3. 如果设备列表非空，则必须从存储中精确获取该设备

这意味着“默认设备兜底”只会在完全空库时触发，不会与管理员已配置的数据混用。

鉴权规则随后执行：

- 设备不存在 -> `404`
- 设备被禁用 -> `403`
- token 不匹配 -> `401`

## 2. 读取 DNS 配置

`resolveDNSConfigJSON()` 从全局设置中读取 key 为 `dns_config` 的值：

- 读取成功 -> 返回原始 JSON 字符串
- 未找到 -> 返回空字符串
- 其它错误 -> 直接终止请求

之后 `singbox.ResolveDNS(configJSON)` 决定最终 DNS：

- 空字符串 -> 使用 `GetDefaultDNS()`
- 非法 JSON -> 记录告警并使用 `GetDefaultDNS()`
- 合法 JSON -> 直接使用用户配置

所以 DNS 是“尽量使用后台配置，失败时始终兜底”的设计，不会因为坏配置导致整个生成接口失败。

## 3. 解析设备入站

入站生成分成两步：

### 3.1 解析绑定关系

`resolveGenerateInbounds(deviceCode)` 先调用 `storage.ListDeviceInbounds(deviceCode)`：

- 如果设备没有绑定记录，则回退到 `GetDefaultDeviceInbounds(deviceCode)`

绑定关系决定某台设备会带哪些入站模板。

### 3.2 解析入站模板

然后读取 `storage.ListInbounds()`：

- 如果存储为空，则回退到 `GetDefaultInbounds()`

接下来流程会：

1. 用 `tag -> inbound` 建立索引
2. 按绑定的 `Sort`、`InboundTag` 稳定排序
3. 根据绑定关系筛出真正需要的模板

最后再交给 `singbox.GetInbounds(items)` 做 JSON 反序列化：

- 跳过 `nil` 和 `Enabled=false` 的模板
- `ConfigJSON` 反序列化失败时记录告警并跳过
- 输出 `[]entity.SingInbound`

这使后台可以直接维护原始 sing-box inbound JSON，而生成层只负责校验和注入。

## 4. 生成 WireGuard endpoint

如果设备 `WireGuardTag` 为空，则直接不生成 endpoint。

否则 `resolveGenerateEndpoints(device)` 会：

1. 读取对应 `WireGuard` 模板
2. 如果未找到且是空存储兼容场景，则尝试默认模板
3. 读取该模板下所有 `WireGuardPeer`
4. 如果 Peer 为空，则尝试默认 Peer
5. 调用 `singbox.GetWireGuardEndpoints(wg, peers, device)`

`GetWireGuardEndpoints` 的关键规则：

- `wg == nil`、`device == nil`、模板禁用、标签不匹配时返回空
- 设备必须提供 `WireGuardClientAddr` 和 `WireGuardClientKey`
- `WireGuardClientAddr` 不带掩码时自动补 `/32`
- Peer 按 `Sort`、`ID` 排序
- 仅保留 `Enabled=true` 的 Peer

最终输出到 `SingBoxConfig.endpoints`。

## 5. 处理 Outbound 缓存与设备过滤

当前生成链路统一通过 `resolveGenerateOutbounds(ctx, deviceCode)` 处理 Outbound，不再区分“额外出站”和“订阅节点拉取”两条独立主线。

### 5.1 订阅筛选与禁用收集

系统会先读取全部 `Subscribe`，按以下条件分为两组：

**禁用/不可见组**（收集到 `disabledSubscribes` 集合）：
- `Subscribe.Status == false`
- 订阅 `VisibleDevices` 对当前设备不可见

**活跃组**（参与后续刷新）：
- `Subscribe.Status == true`
- 订阅 `VisibleDevices` 对当前设备可见

禁用/不可见组的订阅名称会被记录下来，用于在步骤 5.5 中过滤其已缓存的 Outbound。

### 5.2 缓存过期判断

对每个命中的订阅，生成链路会调用 `needsRefresh(subscribe)` 判断是否需要刷新缓存：

- `OutboundLastFetchTime == nil`：从未刷新过，需要刷新
- `OutboundCacheDuration <= 0`：无缓存模式，每次生成都尝试刷新
- 当前时间超过 `OutboundLastFetchTime + OutboundCacheDuration`：缓存过期，需要刷新
- 其它情况：直接使用已缓存的 Outbound

### 5.3 刷新订阅缓存

需要刷新时，`RefreshSubscriptionOutbound` 会执行：

1. `httpGetBytes(ctx, subscribe.URL, subscribe.UserAgent)` 拉取订阅原文
2. `parseSubscriptionOutbounds(body)` 解析为 `entity.SingBoxOut`
3. 将结果转换为统一 `entity.Outbound`，来源标记为 `SUBSCRIPTION`
4. 删除该订阅旧缓存
5. 批量写入新缓存
6. 更新订阅的 `OutboundLastFetchTime`、`OutboundLastFetchStatus`、`OutboundLastFetchError`

网络拉取规则：

- 超时 30 秒
- 自定义 `User-Agent` 为空时回退到内置桌面浏览器 UA
- 非 `200` 响应视为失败

内容解析规则：

1. 先尝试把响应体当作 Base64 解码
2. 解码失败则按明文继续处理
3. 按换行拆成多条节点 URL
4. 根据 `scheme://` 选择协议解码器

当前启用的协议映射为：

- `ss`
- `trojan`
- `vmess`

无法识别或解析失败的节点会被跳过，并记录 warning。

### 5.4 刷新失败时的降级

如果订阅刷新失败：

- 生成流程不会直接失败
- 服务会把订阅状态更新为 `FAILED`
- 已有旧缓存会被保留，供本次生成继续使用
- 如果既没有旧缓存又刷新失败，该订阅等价于本次不产生任何出站

### 5.5 从统一 Outbound 表读取最终候选项

完成刷新判断后，系统调用 `storage.GetOutboundsByDevice(deviceCode)` 读取当前设备可见且启用的所有 Outbound。

过滤规则：

- 只保留 `Enabled=true` 的记录
- `VisibleDevices` 为空表示全部设备可见
- `VisibleDevices` 非空时，仅当前设备精确命中才可见

### 5.6 过滤禁用订阅的缓存 Outbound

从 Outbound 表读取结果后，系统会根据步骤 5.1 中收集的 `disabledSubscribes` 集合，过滤掉属于禁用/不可见订阅的 `SUBSCRIPTION` 来源 Outbound：

- 仅针对 `Source=SUBSCRIPTION` 的记录进行检查
- `MANUAL` 来源的 Outbound 不受订阅状态影响
- 这确保了禁用某个订阅后，其已缓存的节点不会出现在最终生成的配置中

当过滤后存储中完全没有任何 Outbound 时，仍会回退到 `GetDefaultExtraOutbounds()`，用于兼容历史默认行为。

## 6. 构造节点分组出站

最终候选项会先交给 `singbox.GetExtraOutbounds(deviceCode, items)` 做排序、过滤和 JSON 反序列化，再由 `GetOutbounds` 根据 `NodeGroup` 构造出站组。

### 6.1 成员筛选

`constructOutboundGroup` 内部调用 `outboundGroupRuleFilter(groupRule, tags)`：

- `Include` 为空时，默认包含全部标签
- `Include` 非空时，按逗号拆分，只要标签包含任一关键字就命中
- `Exclude` 非空时，再从命中集合中剔除包含排除关键字的标签
- 输出顺序保持与原始标签顺序一致，避免 map 遍历导致结果不稳定

### 6.2 分组类型

目前主要支持两种：

- `selector`
- `urltest`

规则：

- `selector` 的 `default` 设为第一个成员
- `urltest` 默认测试地址为 `https://www.gstatic.com/generate_204`
- `urltest` 额外带 `interval=10m`、`tolerance=50`

如果某分组筛选后没有成员，则该分组不会输出。

## 7. 注入固定出站

`GetOutbounds` 在最后固定追加：

```json
{ "type": "direct", "tag": "direct" }
```

这条固定出站用于配合：

- 路由基础规则中的 `clash_mode=direct`
- DNS 默认配置中的直连解析链路

## 8. 生成路由

路由由 `singbox.GetRoute(device.Code, ruleSets)` 负责。

### 8.1 基础规则

`baseRules()` 固定注入：

- `protocol=dns` -> `action=hijack-dns`
- `clash_mode=direct` -> `outbound=direct`
- `clash_mode=global` -> `outbound=select`
- `protocol=quic` -> `action=reject`

### 8.2 规则集过滤

所有 `RuleSet` 会先按 `Sort` 升序排序，然后按设备过滤：

- `AbleDevices` 为空 -> 对所有设备生效
- `AbleDevices` 非空 -> 仅当字符串包含当前 `device` 时生效

这里用的是字符串包含判断，不是独立数组字段，因此设备编码命名应避免产生歧义。

### 8.3 规则集输出

`baseRuleSets` 会根据 `RuleSetType` 构造：

- `remote` -> 输出远程规则集，包含 `format`、`url`、`download_detour`
- 其它 -> 视为 inline，本地 `Content` 反序列化为 `rules`

非法的本地 JSON 规则集会被跳过。

### 9.4 路由规则输出

对每个生效规则集，再追加一条：

```text
rule_set = [tag]
outbound = ruleSet.Outbound
```

最终 `route.final` 固定为 `general`，`auto_detect_interface=true`。

## 9. 生成 experimental

`singbox.GetExperimental(device)` 按设备类型返回：

- `phone` -> `nil`
- `tv` -> 开启 Clash API，地址固定为 `192.168.10.66:9090`
- 其它设备 -> Clash API 地址固定为 `127.0.0.1:9090`

这部分仍带有明显设备类型假设，是当前生成链路里仍然偏硬编码的一段逻辑。

## 10. 汇总为最终配置

当所有部件准备完成后，`Generated` 组装：

```go
entity.SingBoxConfig{
    DNS:          singbox.ResolveDNS(dnsConfigJSON),
    Endpoints:    endpoints,
    Route:        singbox.GetRoute(device.Code, ruleSets),
    Experimental: singbox.GetExperimental(device.Code),
    Inbounds:     singbox.GetInbounds(inbounds),
    Outbounds:    outbounds,
}
```

然后通过 `c.JSON(http.StatusOK, singBoxConfig)` 返回给客户端。

Surge 输出入口 `SurgeGenerated` 复用前面的致命错误处理和 `resolveGenerateOutbounds(ctx, deviceCode)`，但不组装 sing-box 专属的 Inbound、Endpoint、Experimental。它读取 `ListNodeGroups()` 和 `ListRuleSets()` 后调用 `surge.Render(...)`，通过 `text/plain` 返回包含 `[General]`、`[Proxy]`、`[Proxy Group]`、`[Rule]` 的配置文本。

Surge 渲染规则：

- `[General]` 从 sing-box DNS server 中提取上游地址，跳过 `rcode://` 这类非上游地址
- `[Proxy]` 导出 Shadowsocks、Trojan、VMess；不支持或关键字段缺失的节点跳过并记录 warning
- `[Proxy Group]` 复用同一份 include / exclude 筛选逻辑，且只引用已成功导出的代理名称
- `[Rule]` 对 remote 规则集输出 `RULE-SET,<url>,<outbound>`，对本地规则集展开常见域名、CIDR、GEOIP 规则，最后追加 `FINAL,general`

## 错误处理策略

生成链路不是“任一子步骤失败就整体失败”，而是区分致命错误和可降级错误。

### 致命错误

以下情况直接返回错误响应：

- 设备不存在、禁用、token 错误
- 存储层读取失败
- DNS 全局设置读取失败
- `GetOutbounds` 返回整体错误

### 可降级错误

以下情况通常只影响局部输出，不中断整体生成：

- 某个订阅拉取失败
- 某条节点 URL 解析失败
- 某个 Inbound `ConfigJSON` 非法
- 某个 Extra Outbound `ConfigJSON` 非法
- DNS 配置 JSON 非法
- 某个本地 RuleSet `Content` 非法

这种策略保证了系统在配置存在局部脏数据时仍尽量生成可用结果。

## 稳定性设计

为了让生成结果尽可能稳定、可测试，代码中做了几处显式处理：

- Inbound 按 `Sort`、`Tag` 排序
- DeviceInbound 按 `Sort`、`InboundTag` 排序
- WireGuard Peer 按 `Sort`、`ID` 排序
- RuleSet 按 `Sort` 排序
- 节点分组成员按原始订阅标签顺序输出

这些细节对测试和最终配置 diff 都很重要。

## 当前生成链路的已知限制

- 订阅抓取是请求时同步完成的，慢订阅会直接拉长接口耗时
- 订阅失败目前只写日志，没有缓存和重试机制
- `RuleSet.AbleDevices` 使用字符串包含判断，匹配边界不够严格
- `experimental` 配置仍以设备编码常量驱动，尚未完全数据化
- 固定路由最终出站为 `general`，依赖外部确保存在对应节点组

这些限制不影响主功能，但决定了后续性能优化和配置一致性治理的重点。
