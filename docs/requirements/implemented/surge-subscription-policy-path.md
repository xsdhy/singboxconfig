# Surge 输出改用订阅 policy-path 引用

## 背景

此前 `/open/surge/:device` 输出会把订阅缓存节点全部展开为 `[Proxy]` 行，与手工节点混在一起。
Surge 原生支持在策略组中通过 `policy-path=<url>` 直接引用订阅地址（[Policy Including](https://manual.nssurge.com/policy-group/policy-including.html)），
由客户端自行拉取并更新订阅，无需服务端展开。

## 需求

- Surge 配置中不再展开订阅节点：
  - 手工添加（`Source=MANUAL`）的 Outbound 仍然逐条输出为 `[Proxy]` 行；
  - 来自订阅（`Source=SUBSCRIPTION`）的节点不输出，改为每个订阅源生成一个
    `select` 策略组，携带 `policy-path=<订阅地址>`；
  - 订阅的缓存时长（分钟）映射为 `update-interval`（秒）。
- 节点分组通过 `include-other-group=<订阅策略组,...>` 引用订阅策略组，
  并把 `Include` / `Exclude` 关键字翻译为 `policy-regex-filter` 正则
  （Include 为或关系，Exclude 用负向先行断言剔除，关键字做正则转义），
  与 sing-box 链路 `FilterOutboundGroupTags` 的子串匹配语义保持一致。
- 订阅可见性约束沿用现有规则：禁用（`Status=false`）或对当前设备不可见
  （`VisibleDevices` 不含设备编码）的订阅不输出。
- Surge 生成时不再触发订阅缓存刷新（订阅由客户端拉取，服务端缓存与 Surge 输出无关）。

## Shadowrocket 调研结论

Shadowrocket **不支持** `policy-path`。其社区手册（lowertop.github.io/Shadowrocket、
GMOogway/shadowrocket-rules 配置文档）中 `[Proxy Group]` 仅支持
`url / interval / timeout / tolerance / policy-regex-filter / policy-select-name / hidden` 等参数；
订阅引用只能在 App 界面通过“订阅”开关选择，没有配置文本语法。
因此 Shadowrocket 输出链路维持原有行为：继续展开订阅缓存节点。

## 实现

- `convert/surge/surge.go`
  - `Render` 新增 `subscribes []*entity.Subscribe` 参数；
  - 新增 `renderSubscriptionGroupSection`：每个可见订阅渲染为
    `name = select, policy-path=<url>[, update-interval=<秒>]`；
  - `renderProxyGroupSection`：存在订阅策略组时追加
    `include-other-group` 与 `policy-regex-filter`；
    分组成员为空但有订阅策略组时不再跳过该分组；
  - 新增 `buildPolicyRegexFilter` / `quoteKeywords` / `isSubscribeVisibleForDevice`。
- `service/outbound_cache.go`
  - 新增 `resolveManualGenerateOutbounds`：只返回设备可见且启用的手工 Outbound，
    不触发订阅刷新；存储完全为空时保留默认手工节点兜底；
  - 新增 `resolveVisibleSubscribes`：返回启用且设备可见的订阅源。
- `service/generated.go`
  - `SurgeGenerated` 改为调用上述两个函数，并把订阅列表传入 `surge.Render`。

## 测试

- `convert/surge/surge_test.go`
  - `TestRenderSubscriptionPolicyPath`：验证订阅策略组输出、禁用/不可见订阅过滤、
    手工节点保留、`include-other-group` 与 `policy-regex-filter`；
  - `TestRenderSubscriptionOnlyGroup`：验证无手工节点命中时分组仍输出。
