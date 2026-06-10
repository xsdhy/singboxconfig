# 规则出站存在性校验

> 状态：已实现

## 背景

`/open/*` 公开接口的三条配置输出链路（sing-box / Surge / Shadowrocket）在生成 `[Rule]` / `route.rules` 时，会把每个规则集（`entity.RuleSet`）的 `outbound` 直接写成路由目标。但 `RuleSet.Outbound` 是后台维护的自由字符串，可能指向：

- 已被删除或改名的节点分组 `tag`
- 当前设备实际不可见的订阅节点（被设备可见性过滤后不存在）
- 该客户端协议能力不支持、因而被跳过的代理（如 Surge 跳过 VLESS）
- 单纯的拼写错误

改动前各链路的处理并不一致，且都会产生“悬空引用”：

- **sing-box**：`GetRoute` 甚至不接收出站列表，直接输出 `{"rule_set":[tag], "outbound": ruleSet.Outbound}`，不做任何校验。
- **Surge / Shadowrocket**：`policyReference` 在出站既不是内置 `DIRECT`/`REJECT`、又未命中已导出代理或策略组时，会回退到 `policyName(原始字符串)`，生成一条指向不存在策略的规则行。

悬空引用会让目标客户端在导入时报错或让该规则静默失效，排查成本高。

## 目标

- 三条配置输出链路在输出规则时，统一校验规则引用的 `outbound`/proxy 是否真实存在，**按目标软件各自实际可生成的目标集合判定**
- 引用不存在目标的规则直接跳过并记录 warning，不输出悬空规则、也不中断整体生成
- 不改变其它既有行为

## 非目标

- 不校验 `FINAL` / `final` 兜底策略（按现状固定为 `general`，由部署方保证存在）
- 不校验 `downloadDetour` 等其它出站引用
- 不在后台保存规则集时做前置校验（仍是生成期校验）
- 不改动节点分组、订阅、设备可见性等数据层逻辑

## 各软件的“有效目标集合”

| 软件 | 有效出站/策略目标 |
|------|------|
| sing-box | 当前设备最终出站列表中的所有 `tag`（含节点分组出站与固定 `direct`） |
| Surge | 已成功导出的代理名（`proxyNames`）、策略组名（`groupNames`）、内置 `DIRECT` / `REJECT`，空标签按兜底处理 |
| Shadowrocket | 同 Surge |

按各软件实际能生成的目标判定是关键：Surge 不支持的协议（如 VLESS）不会进入 `proxyNames`，因此指向它的规则在 Surge 输出里会被正确跳过，而同一规则在协议覆盖更广的 Shadowrocket 输出里则可能保留。

## 实现方案

### 1. sing-box（`convert/singbox/route.go`）

- `GetRoute` 新增 `outbounds []entity.SingBoxOut` 入参，构建已存在出站标签集合
- 遍历规则集时，`ruleSet.Outbound` 非空且不在集合中则跳过该条 + warning
- 调用方 `service/generated.go` 传入已解析的 `outbounds`（其解析顺序本就在路由构造之前）
- `Final` 兜底保持 `general` 不变，不参与校验

### 2. Surge / Shadowrocket（`convert/surge`、`convert/shadowrocket`）

- 把原 `policyReference(tag) string` 拆出 `resolvePolicy(tag) (name, ok)`：空标签与内置 `DIRECT`/`REJECT` 始终 `ok=true`，其余必须命中 `proxyNames` 或 `groupNames`
- `renderRuleSection` 改用 `resolvePolicy`，`ok=false` 时跳过该规则（remote 与 local/inline 两种类型均覆盖）+ warning
- `policyReference` 保留为 `resolvePolicy` 的薄封装，仅供 `FINAL` 兜底使用，行为不变

## 实现范围

### 后端

- [x] `convert/singbox/route.go`：`GetRoute` 增加出站列表入参并跳过悬空规则
- [x] `service/generated.go`：调用 `GetRoute` 时传入 `outbounds`
- [x] `convert/surge/surge.go`：`resolvePolicy` 校验 + 规则跳过
- [x] `convert/shadowrocket/shadowrocket.go`：`resolvePolicy` 校验 + 规则跳过
- [x] 单元测试：三条链路各新增 `TestRenderSkipsRuleWithUnknownOutbound` / `TestGetRouteSkipsRuleWithUnknownOutbound`，覆盖已存在出站（分组/代理/`direct`）保留、不存在出站（含 remote）跳过、FINAL/Final 不受影响

### 文档

- [x] `reference/api-reference.md`：三条输出接口补充“规则出站校验”说明
- [x] `architecture/config-generation.md`：路由生成段落补充校验逻辑、更新 `GetRoute` 签名
- [x] `modules/rule-set.md`：生成方式与“当前限制”补充校验行为
- [x] `reference/singbox-config-schema.md`：`route` 规则说明补充校验行为
- [x] 直接创建于 `requirements/implemented/`（INDEX.md 仅索引需求目录，沿用现有约定不单列文件）

## 验收结果

> 已完成后端实现、单元测试与文档同步；`go build`、`go vet`、`go test ./...` 全部通过。

- [x] 规则引用已存在出站（节点分组 / 代理 / `direct`）时正常输出
- [x] 规则引用不存在出站时被跳过并记录 warning，不产生悬空规则
- [x] Surge 中指向不支持协议（未导出代理）的规则被正确跳过
- [x] `FINAL` / `final` 兜底策略不受影响
- [x] 三条配置输出链路其余行为保持不变
