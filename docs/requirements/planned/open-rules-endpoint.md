# 多客户端规则集输出接口（/open/rules）

> 状态：计划中

## 背景

目前 `/open/*` 公开接口只有「完整配置」一类输出：

```
GET /open/generate/:device?token=...        → sing-box 完整 JSON 配置
GET /open/surge/:device?token=...           → Surge 完整文本配置
GET /open/shadowrocket/:device?token=...    → Shadowrocket 完整文本配置
GET /open/ruleset/:tag                      → 规则集原始 JSON（仅 sing-box source 格式，无鉴权）
```

规则集（`entity.RuleSet`）本身是客户端无关的数据，但当前的对外暴露方式存在两个问题：

1. **格式单一**：`/open/ruleset/:tag` 只会原样返回 `Content` 中的 sing-box source JSON，Surge / Shadowrocket / Clash 等客户端无法直接以 `RULE-SET` / `rule-providers` 方式引用它。
2. **只能内联**：Surge 与 Shadowrocket 配置链路中，local/inline 规则集是在渲染时整体展开进配置文本的（见 `convert/surge/surge.go` 的 `expandLocalRuleSet`）。规则集一旦更新，客户端必须重新拉取整份配置才能生效，也无法利用客户端自身对远程规则集的独立更新与缓存能力。

不同软件对「规则集」有各自的格式定义：sing-box 是 source JSON / binary（.srs），Surge 与 Shadowrocket 是逐行的 `DOMAIN-SUFFIX,xxx` 列表（.list），Clash 是 rule-provider YAML。本需求要做的，是把系统中已有的 Rules 数据，**按目标软件的规则集定义封装成对应格式**，通过一类新的 open 接口对外输出。

## 目标

- 新增一类规则集输出接口：

  ```
  GET /open/rules/:tag/:software/:device?token=...
  ```

  - `:tag` —— 规则集唯一标识（`RuleSet.Tag`）
  - `:software` —— 目标软件名称，第一版支持 `singbox` / `surge` / `shadowrocket`
  - `:device` —— 设备编码，用于 token 鉴权与 `AbleDevices` 可见性校验

- 同一份规则集数据，按 `:software` 渲染为对应客户端可直接引用的规则集格式
- 复用现有设备鉴权链路（`resolveGenerateDevice` + token 校验）与规则集按设备过滤逻辑（`AbleDevices`）
- 为后续「Surge / Shadowrocket 配置中以 RULE-SET 引用本服务规则集 URL、替代内联展开」打下基础

## 非目标

- 不移除或修改现有 `/open/ruleset/:tag` 接口（保留兼容，后续视使用情况再决定是否废弃）
- 不改动现有三条完整配置输出链路的行为（Surge / Shadowrocket 配置中的规则内联展开方式本期不变，切换为 RULE-SET 引用作为后续增量）
- 不做 sing-box `.srs` 二进制格式的编译输出（sing-box 端输出 source JSON 即可，`format: source` 可被直接引用）
- remote 类型规则集不做「下载源文件再转格式」的代理转换（见下文处理策略）
- 第一版不支持 Clash 格式（接口设计上预留 `:software` 扩展位即可）

## 接口设计

### 路由

```
GET /open/rules/:tag/:software/:device?token=...
```

示例：

```
/open/rules/geosite-cn/surge/iphone15?token=xxx        → Surge .list 文本
/open/rules/geosite-cn/shadowrocket/iphone15?token=xxx → Shadowrocket .list 文本
/open/rules/geosite-cn/singbox/iphone15?token=xxx      → sing-box source JSON
```

### 鉴权与可见性

1. 通过 `:device` 解析设备（复用 `resolveGenerateDevice`），设备不存在或被禁用返回 404 / 403
2. 校验 `?token=` 与设备 Token 一致，不一致返回 403
3. 校验规则集对该设备可见（`AbleDevices` 为空或包含该设备编码，复用 `isRuleSetVisibleForDevice` 同等逻辑），不可见按 404 处理（不泄露规则集存在性）

> 现有 `/open/ruleset/:tag` 无任何鉴权，新接口借 `:device` + token 补上这一短板，这也是路径中包含设备名称的原因。

### 各软件输出格式

| `:software` | 输出格式 | Content-Type | 说明 |
|------|------|------|------|
| `singbox` | sing-box source JSON（`{"version":1,"rules":[...]}`） | `application/json` | 与现有 `/open/ruleset/:tag` 输出对齐，可被 sing-box remote rule_set（format=source）直接引用 |
| `surge` | 逐行规则列表（.list） | `text/plain` | `DOMAIN-SUFFIX,example.com`、`IP-CIDR,1.1.1.0/24,no-resolve` 等，可被 `RULE-SET,<url>,<policy>` 引用 |
| `shadowrocket` | 逐行规则列表（.list） | `text/plain` | 与 Surge 同构的列表格式 |

注意：Surge / Shadowrocket 的 `RULE-SET` 列表文件**不含策略字段**（策略写在主配置的 `RULE-SET,<url>,<policy>` 行里），因此输出行只到规则本身，不带 `RuleSet.Outbound`。

### 规则类型映射（sing-box source → 列表行）

| sing-box 规则字段 | Surge / Shadowrocket 行 |
|------|------|
| `domain` | `DOMAIN,<v>` |
| `domain_suffix` | `DOMAIN-SUFFIX,<v>` |
| `domain_keyword` | `DOMAIN-KEYWORD,<v>` |
| `ip_cidr` | `IP-CIDR,<v>` / `IP-CIDR6,<v>`（按 IPv6 区分，附 `no-resolve`） |
| `process_name` | `PROCESS-NAME,<v>`（Shadowrocket 不支持时跳过 + warning） |
| 其它不支持字段 | 跳过该条 + 记录 warning，不中断整体输出 |

该映射与 `convert/surge` 中内联展开已实现的逻辑一致，应抽取共享而非复制（见下文）。

### 规则集类型处理策略

| `RuleSetType` | 处理 |
|------|------|
| `local` / `inline` | 解析 `Content`（沿用 `parseLocalRules` 的两种容错形态：完整 source JSON 或裸 rules 数组），转换为目标格式输出 |
| `remote` | `singbox`：302 重定向到 `RuleSet.URL`；`surge` / `shadowrocket`：第一版返回 400 并附说明（源文件是 .srs 二进制，无法转换），后续可评估代理转换 |

## 实现方案

### 1. 转换逻辑下沉为公共包

`convert/surge` 与 `convert/shadowrocket` 中已各自实现「sing-box 规则 → 列表行」的展开逻辑（`parseLocalRules` / `expandLocalRuleSet` 及其 shadowrocket 对应物）。本需求需要在配置渲染之外独立复用这段逻辑，建议：

- 新增 `convert/ruleset/` 包，承载：
  - sing-box source `Content` 的解析（容错两种 JSON 形态）
  - 规则字段 → Surge / Shadowrocket 列表行的映射
  - sing-box source JSON 的规范化输出
- `convert/surge`、`convert/shadowrocket` 的内联展开改为调用该包（行为不变，仅消除重复）

若评估重构回归风险偏高，第一版也可仅新增 `convert/ruleset/` 并从 surge 包迁移纯函数，shadowrocket 的接入作为后续清理项。

### 2. service 层

在 service 层新增 `GetRulesBySoftware`（命名可调整）处理函数：

1. 解析 `:device` 并做 token 鉴权（复用 `resolveGenerateDevice`）
2. 按 `:tag` 查找规则集，校验设备可见性
3. 按 `:software` 分派到 `convert/ruleset/` 对应渲染函数；未知软件名返回 400
4. `c.String` / `c.JSON` 输出，附带正确的 Content-Type

并在 `cmd/server/main.go` 注册路由：

```go
r.GET("/open/rules/:tag/:software/:device", service.GetRulesBySoftware)
```

### 3. 错误处理

沿用配置生成链路的「致命 / 可降级」分层：

**致命错误（直接返回错误响应）**
- 设备不存在、禁用、token 错误 → 403 / 404
- 规则集不存在或对设备不可见 → 404
- 未知 `:software` → 400
- remote 规则集请求 surge/shadowrocket 格式 → 400
- `Content` 整体不是合法 JSON → 500

**可降级（跳过局部，继续输出）**
- 单条规则包含目标软件不支持的字段 → 跳过该条 + warning 日志

## 实现范围

### 后端

- [ ] 新增 `convert/ruleset/` 公共转换包（解析 + 各软件格式渲染）
- [ ] `convert/surge` / `convert/shadowrocket` 内联展开切换到公共包（行为不变）
- [ ] 新增 `GET /open/rules/:tag/:software/:device` 路由与 service 处理函数（鉴权、可见性、分派）
- [ ] 单元测试：各软件格式渲染、规则类型映射、不支持字段跳过、remote 类型分支、鉴权与可见性

### 前端（可选，后置）

- [ ] 规则集管理页提供各软件格式的引用 URL 展示与复制入口（需选择设备以拼出完整 URL）

### 文档

- [ ] 实现后将本文件移至 `requirements/implemented/`
- [ ] 更新 `reference/api-reference.md`（新增 `/open/rules` 接口说明）
- [ ] 更新 `modules/rule-set.md`（补充多客户端格式输出能力）
- [ ] 视情况更新 `INDEX.md` 核心功能描述

## 验收标准

- [ ] `/open/rules/:tag/singbox/:device?token=...` 输出可被 sing-box remote rule_set（format=source）直接引用的 JSON
- [ ] `/open/rules/:tag/surge/:device?token=...` 输出可被 Surge `RULE-SET,<url>,<policy>` 直接引用的列表文本
- [ ] `/open/rules/:tag/shadowrocket/:device?token=...` 输出可被 Shadowrocket `RULE-SET` 直接引用的列表文本
- [ ] token 错误、设备禁用、规则集对设备不可见时均无法获取规则内容
- [ ] remote 类型规则集按策略处理（singbox 重定向、surge/shadowrocket 明确报错）
- [ ] 不支持的规则字段被跳过且不影响整体输出
- [ ] 现有 `/open/ruleset/:tag` 与三条完整配置输出链路行为不变
