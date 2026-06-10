# 规则集管理

## 模块职责

规则集管理负责维护规则内容和命中后的出站策略。sing-box 输出会把它转换为 `route.rule_set` 与 `route.rules`，Surge / Shadowrocket 输出会把它转换为 `[Rule]` 段中的 `RULE-SET` 或展开后的规则行。

代码入口：

- 实体定义：`entity/rule_set.go`
- 管理接口：`service/service.go`
- sing-box 生成转换：`convert/singbox/route.go`
- Surge 生成转换：`convert/surge/surge.go`
- Shadowrocket 生成转换：`convert/shadowrocket/shadowrocket.go`

## 数据模型

`entity.RuleSet` 字段：

- `name`：后台展示名称
- `tag`：规则集唯一标识
- `ruleSetType`：通常为 `remote` 或 `local`
- `format`：规则集格式，远程常用 `binary`，本地常用 `source`
- `content`：本地规则集 JSON 文本
- `url`：远程规则集地址
- `outbound`：命中该规则集后的默认出站
- `downloadDetour`：远程规则集下载时使用的出站
- `ableDevices`：逗号分隔设备编码，用来限制可见设备
- `sort`：输出顺序，越小越靠前

## 管理接口

接口路径为 `/api/rule-sets`：

- `POST /api/rule-sets`
- `PUT /api/rule-sets/:tag`
- `DELETE /api/rule-sets/:tag`
- `GET /api/rule-sets`
- `GET /open/ruleset/:tag`

说明：

- 创建时会检查 `tag` 唯一性
- 更新时会把 `content` 中的换行和空格去掉，便于以紧凑 JSON 持久化
- `GET /open/ruleset/:tag` 会把 `content` 反序列化后直接输出，主要面向本地规则集内容查看

## 生成到 sing-box 的方式

`convert/singbox.GetRoute()` 会把规则集拆成两层结构：

1. `route.rule_set`
2. `route.rules` 中按规则集引用生成的路由规则

### `route.rule_set`

由 `baseRuleSets()` 生成：

- `ruleSetType == "remote"`：输出 `type=remote`、`tag`、`format`、`url`、`download_detour`
- 其他情况：尝试把 `content` 解析成 JSON，成功后输出 `type=inline`

如果本地 `content` 不是合法 JSON，该规则集会被静默跳过，不会阻断整个配置生成。

### `route.rules`

由 `GetRoute()` 生成：

- 先写入一组基础规则
- 再按 `sort` 升序遍历规则集
- 每条规则集追加一条 `{"rule_set":[tag], "outbound": outbound}`
- 规则集的 `outbound` 不在当前设备最终出站列表（含节点分组出站与 `direct`）中时，**跳过该条规则并记录 warning**，避免生成指向不存在出站的路由规则

## 生成到 Surge / Shadowrocket 的方式

`convert/surge.Render()` 和 `convert/shadowrocket.Render()` 会把规则集输出到 `[Rule]` 段：

- `ruleSetType == "remote"`：输出 `RULE-SET,<url>,<outbound>`
- `local` / `inline`：解析 `content` 中常见的 `domain`、`domain_suffix`、`domain_keyword`、`domain_regex`、`ip_cidr`、`geoip` 字段，并展开为对应客户端的规则行
- 规则的 `outbound` 既不是内置 `DIRECT` / `REJECT`，也未命中任何已导出的代理或策略组时，**跳过该条规则并记录 warning**（按各软件实际可生成的代理/策略组判定）
- 非法本地 JSON 会跳过并记录 warning，不中断整体配置生成
- 最后固定追加 `FINAL,general`（兜底策略不参与上述存在性校验）

## 规则集 URL 引用模式（依赖系统 Host）

除“展开/内联”外，规则集还支持以**远程 URL 引用**方式输出，由全局设置 `system_host`（系统 Host）控制：

- 配置合法 `system_host` 后，三条整份配置链路会把**有效的** local / inline 规则集改为指向本服务规则集 open 接口的远程引用：
  - sing-box：`baseRuleSets()` 输出 `type:"remote"`、`format:"source"`、`url` 指向 `.../open/rules/<tag>/singbox/<device>?token=<token>`
  - Surge / Shadowrocket：`renderRuleSection()` 输出单行 `RULE-SET,<url>,<policy>`，不再逐条展开
- URL 由 `convert/ruleset.BuildRuleSetURL()` 拼接：`tag`、`device` 使用 path escape，`token` 使用 query escape，`system_host` 先去掉尾斜杠
- 客户端通过 `GET /open/rules/:tag/:software/:device?token=...` 拉取该规则集内容（见[API 文档](../reference/api-reference.md)），解析/渲染逻辑统一收敛在 `convert/ruleset` 包
- **鉴权模型**：规则集 open 接口复用整份配置接口的设备解析、启用状态、token 校验，并额外校验 `AbleDevices` 可见性（不可见按 404）
- **降级与兼容**：
  - `system_host` 未配置或非法：回退到原“展开/内联”行为，旧部署零配置可用
  - local / inline `Content` 非法且 host 已配置：不生成指向坏内容的远程 URL，回退到展开/内联（仍会跳过坏内容并记录 warning）
  - remote 规则集：始终保持原 `URL` 引用，不受影响
  - 有效规则集过滤：sing-box 先按 `AbleDevices` 与 `Outbound` 存在性过滤，仅对有效规则集同时输出 `route.rule_set` 与 `route.rules`，避免输出“会被客户端下载却不被引用”的远程规则集
- **设备 token 轮换**：生成的规则集 URL 携带设备 token，token 修改后旧整份配置中的规则集 URL 会失效，需重新拉取整份配置

## 设备可见性控制

`ableDevices` 用于限制哪些设备会加载该规则集，当前逻辑是：

- 为空：所有设备可见
- 非空：使用 `strings.Contains(ableDevices, deviceCode)` 判断

这里是简单子串匹配，不是严格的“逗号分隔精确匹配”。例如 `ableDevices="pad,phone"` 时，`deviceCode="ph"` 也可能误命中。这是当前实现的一个已知边界。

## 基础路由规则

除规则集派生规则外，系统还会固定输出几条基础规则：

- `protocol=dns -> action=hijack-dns`
- `clash_mode=direct -> outbound=direct`
- `clash_mode=global -> outbound=select`
- `protocol=quic -> action=reject`

同时路由根对象默认值为：

- `final: "general"`
- `find_process: false`
- `auto_detect_interface: true`

这里的 `final="general"` 假设系统中存在名为 `general` 的出站或节点分组；FINAL/Final 兜底策略不做存在性校验（仅规则集派生的普通规则会校验出站是否存在）。

## 与其他模块的关系

- 常把 `outbound` 指向“节点分组”的 `tag`
- `downloadDetour` 也通常引用某个已存在出站
- DNS 默认规则中会引用 `cnip`、`cnsite` 两个规则集 tag，因此常见部署会创建这两个规则集

## 当前限制

- 规则集派生的普通规则会校验 `outbound` 是否存在（不存在则跳过），但 `downloadDetour`、FINAL/Final 兜底等其它引用仍不做校验
- `ableDevices` 采用子串匹配，存在误匹配空间
- 本地规则集保存为字符串，不做结构化字段校验
- `GET /open/ruleset/:tag` 是无鉴权的历史兼容接口，仅适合本地规则集；面向客户端的多软件输出请用 `GET /open/rules/:tag/:software/:device`（带设备鉴权）。远程规则集没有单独下载代理接口
- Surge / Shadowrocket 规则集文件第一版仅映射 `domain`/`domain_suffix`/`domain_keyword`/`domain_regex`/`ip_cidr`/`geoip`，其它字段跳过并记录 warning

## 适合更新本文档的场景

- 调整 `ableDevices` 匹配逻辑
- 新增规则集类型或格式
- 修改基础路由规则
- 引入规则预校验、依赖检查或更严格的引用约束
