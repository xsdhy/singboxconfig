# 需求：规则集 open 接口路径改为 query 参数 + 少量规则改为内联展开

- 状态：已实现（implemented）
- 提出时间：2026-06-11
- 完成时间：2026-06-11
- 关联模块：[规则集管理](../../modules/rule-set.md)、[配置生成流程](../../architecture/config-generation.md)、[API 接口列表](../../reference/api-reference.md)、[设备管理](../../modules/device.md)
- 前置需求：[规则集独立 open 接口 + 系统 Host 设置 + 生成时规则集改为 URL 引用](./ruleset-open-endpoint-and-url-reference.md)

## 背景

[前置需求](./ruleset-open-endpoint-and-url-reference.md)落地后存在两点待优化：

1. **接口路径把 `software`、`device` 放在路径段上**（`/open/rules/:tag/:software/:device`），导致设备名 / 规则集 tag 含斜杠等字符时路径层级易混淆，也不便于扩展更多可选参数。
2. **生成整份配置时，只要配置了 `system_host`，所有 local / inline 规则集都改成远程 URL 引用**。对只有一两条规则的规则集，为此额外引入一次远程请求并不划算，客户端逐条内联反而更轻量直观。

## 需求

### 1. open 接口路径改用 query 参数

- 路由由 `GET /open/rules/:tag/:software/:device` 改为 `GET /open/rules/:tag`。
- `software`、`device` 与既有的 `token` 一律走 query 参数：`GET /open/rules/:tag?software=...&device=...&token=...`。
- 鉴权、可见性、software/remote 校验语义保持不变。

### 2. 规则条数少于阈值时内联展开

- 生成整份配置（`/open/generate`、`/open/surge`、`/open/shadowrocket`）时，local / inline 规则集仅在**规则条数 ≥ 3** 时才改为指向 open 接口的远程 URL 引用。
- 规则条数少于 3 条时，即便 `system_host` 已配置，也直接逐条展开（Surge / Shadowrocket）或内联（sing-box），不引用 open 接口。
- 阈值与计数口径统一收敛在 `convert/ruleset` 包，三软件共用，避免各自实现漂移。

## 实现要点

| 关注点 | 位置 | 说明 |
|--------|------|------|
| 阈值常量 | `convert/ruleset/ruleset.go` | 新增 `InlineThreshold = 3` |
| 规则计数 | `convert/ruleset/ruleset.go` | 新增 `CountLines()`，复用 `RenderLines` 计数（每个域名 / CIDR / 进程项各计一条） |
| inline 规则数组 | `convert/ruleset/ruleset.go` | 新增 `InlineRules()`，剥离外层 `{"version":1,...}` 包装，返回纯 rules 数组 |
| URL 拼接 | `convert/ruleset/ruleset.go` | `BuildRuleSetURL` 改为 `tag` 走路径、`software`/`device`/`token` 走 query 参数 |
| sing-box | `convert/singbox/route.go` | `baseRuleSets()` 按条数判定 remote URL / inline；inline 改用 `InlineRules()` 输出正确的 rules 数组 |
| Surge | `convert/surge/surge.go` | `renderRuleSection()` 按条数判定 RULE-SET URL / 逐条展开 |
| Shadowrocket | `convert/shadowrocket/shadowrocket.go` | 同 Surge |
| 路由注册 | `cmd/server/main.go` | `r.GET("/open/rules/:tag", ...)` |
| 接口处理 | `service/rules_open.go` | `software`、`device` 改用 `c.Query(...)` 读取 |
| 前端复制地址 | `web/src/utils/ruleSetUrl.ts`、`web/src/components/RuleSetCopyURLModal.tsx` | 按新路径 + query 参数拼接绝对地址 |

## 兼容性说明

- 旧路径 `/open/rules/:tag/:software/:device` 不再注册，使用旧地址的客户端会 404，需要重新拉取整份配置或重新复制规则集地址。
- 修复了一处既有隐患：原 inline 回退把整份 `{"version":1,"rules":[...]}` 塞进 sing-box `rule_set.rules` 字段（应为 rules 数组），现已剥离外层包装输出正确数组；本次新增的「少于 3 条内联」路径同样走该正确实现。

## 验收标准

1. ✅ `GET /open/rules/:tag?software=singbox&device=...&token=...`（surge / shadowrocket 同理）返回与改造前一致的内容，鉴权与可见性语义不变。
2. ✅ `BuildRuleSetURL` 输出 `<host>/open/rules/<escaped-tag>?device=...&software=...&token=...`，`tag` 走 path escape，`software`/`device`/`token` 走 query escape（单测覆盖特殊字符）。
3. ✅ 配置合法 `system_host` 且规则条数 ≥ 3：sing-box 输出 `type:"remote"`、`format:"source"` 远程 URL；Surge / Shadowrocket 输出单行 `RULE-SET,<url>,<policy>`。
4. ✅ 配置合法 `system_host` 但规则条数 < 3：sing-box 输出 `type:"inline"`（rules 为正确的规则数组，不带 version 外层）；Surge / Shadowrocket 逐条展开，不输出 RULE-SET URL。
5. ✅ `CountLines` 计数口径与逐行展开一致，内容非法时返回 error 并回退到展开/内联。
6. ✅ 前端「复制地址」生成的 URL 与后端 `BuildRuleSetURL` 路径结构一致，类型检查通过。
7. ✅ 既有单测全部通过，新增阈值边界、URL 格式、inline 数组形态单测。
