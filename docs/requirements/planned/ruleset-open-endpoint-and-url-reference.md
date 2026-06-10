# 需求：规则集独立 open 接口 + 系统 Host 设置 + 生成时规则集改为 URL 引用

- 状态：计划中（planned）
- 提出时间：2026-06-10
- 关联模块：[规则集管理](../../modules/rule-set.md)、[配置生成流程](../../architecture/config-generation.md)、[API 接口列表](../../reference/api-reference.md)、[设备管理](../../modules/device.md)

## 背景

当前 `/open` 下只有“整份配置”接口，按设备输出各软件的**完整配置**：

| 路由 | 处理器 | 输出 |
|------|--------|------|
| `GET /open/generate/:device` | `service.Generated` | sing-box 完整 JSON |
| `GET /open/surge/:device` | `service.SurgeGenerated` | Surge 完整文本 |
| `GET /open/shadowrocket/:device` | `service.ShadowrocketGenerated` | Shadowrocket 完整文本 |
| `GET /open/ruleset/:tag` | `service.GetRuleSetByTag` | 规则集原始 `content`（JSON） |

注：`cmd/server/main.go:110-113` 注册以上路由。`/open/ruleset/:tag` 目前只是把 `RuleSet.Content` 原样反序列化回吐，**不区分软件格式**，本质上是给 sing-box 用的 source JSON。

规则集（`entity.RuleSet`，见 `entity/rule_set.go`）分两类：

- **remote**：带 `URL` 字段，生成时直接以 URL 引用。
  - sing-box：`convert/singbox/route.go:baseRuleSets()` 输出 `type:"remote"` + `url`。
  - Surge / Shadowrocket：输出一行 `RULE-SET,<url>,<policy>`（`convert/surge/surge.go:506-512`、`convert/shadowrocket/shadowrocket.go:648-654`）。
- **local**：把 `Content`（sing-box source 规则数组）在生成时**展开**：
  - sing-box：`baseRuleSets()` 输出 `type:"inline"`，规则内联进配置（`convert/singbox/route.go:131-141`）。
  - Surge：`expandLocalRuleSet()` 把每条规则拆成 `DOMAIN/DOMAIN-SUFFIX/IP-CIDR/...` 多行（`convert/surge/surge.go:514,544`）。
  - Shadowrocket：同上（`convert/shadowrocket/shadowrocket.go:656,686`）。

也就是说，**local 规则集在三种输出里都是“展开成一条条具体规则”塞进最终配置**，而非作为一个独立的规则集 URL 引用。

## 需求描述

本次三个改动：

### 需求 1：新增“规则集级” open 接口

在现有 `/open` 整份配置接口之外，**新增一类按规则集输出的 open 接口**，能针对某个规则集、按指定软件格式输出该规则集的规则内容（不是整份配置）。

- 路径形如 `/open/规则集名称/软件名称`，顺序由实现侧确定（见下方方案）。
- 软件名称至少覆盖 `singbox` / `surge` / `shadowrocket` 三种。
- 返回的是“该规则集在目标软件下的规则集文件内容”，可被对应客户端直接作为远程规则集加载。

### 需求 2：系统设置增加“系统 Host”

在系统设置里新增一项 **系统 Host（system host）**，用于记录本服务对外可访问的基础地址（如 `https://config.example.com`）。该值用于拼接需求 1 接口的**绝对 URL**，供需求 3 在生成配置时引用。

### 需求 3：生成配置时，local 规则集由“展开”改为“URL 引用”

把生成链路里 local 规则集的处理方式，从**展开成一条条规则**改为**输出为一个规则集（完整 URL 地址）**，该 URL 指向需求 1 的新接口（用需求 2 的系统 Host 拼成绝对地址）：

- sing-box：local 规则集由 `type:"inline"` 改为 `type:"remote"` + 指向本服务的 `url`。
- Surge / Shadowrocket：local 规则集由多行展开改为单行 `RULE-SET,<url>,<policy>`。

remote 规则集行为不变（本就是 URL 引用）。

## 方案设计（建议）

### 一、规则集 open 接口（需求 1）

#### 路由

新增：

```text
GET /open/ruleset/:tag/:software
```

- 选 `tag` 在前、`software` 在后，是为了与现有 `GET /open/ruleset/:tag` 共用 `:tag` 这一层参数名，避免 gin 在同层注册不同参数名（如 `:software` 与 `:tag`）时 panic 的冲突问题。
- 现有 `GET /open/ruleset/:tag`（无软件后缀）保留不动，行为兼容（等价于 `software = singbox`）。
- `:software` 取值：`singbox` / `surge` / `shadowrocket`；其余值返回 400。
- 与现有 `/open/*` 一致，本接口**不需要设备 token**（规则集内容本身不含敏感凭据），但需要 `:tag` 命中已存在的规则集，否则 404。

> 备选命名 `/open/ruleset/:software/:tag` 因与现有 `:tag` 参数同层命名冲突，不采用。

#### 输出格式（按软件）

新增处理器 `service.GetRuleSetForSoftware`，按 `:software` 输出：

| software | Content-Type | 内容 |
|----------|--------------|------|
| `singbox` | `application/json` | sing-box source 规则集 JSON，形如 `{"version":1,"rules":[...]}`，即规则集 `Content` 规范化后的结果 |
| `surge` | `text/plain` | Surge 规则集文件：每行 `类型,值`（**不含 policy 列**），如 `DOMAIN-SUFFIX,google.com` |
| `shadowrocket` | `text/plain` | Shadowrocket 规则集文件：同 Surge 的列表格式 |

实现上复用已有展开逻辑，但**剥离 policy 列**（规则集文件里不带出站策略，策略在主配置的 `RULE-SET` 行指定）：

- Surge：抽出 `convert/surge` 现有 `expandLocalRuleSet()` 的“规则 → 文本行”能力，新增一个“只生成 `类型,值`、不拼 policy”的导出函数。
- Shadowrocket：同理，复用 `convert/shadowrocket` 的规则解析与行格式。
- sing-box：直接规范化回吐 `Content`（与现有 `GetRuleSetByTag` 类似）。
- 对 `RuleSetType == remote` 的规则集：该接口可选择 302 跳转到其原始 `URL`，或直接返回 404（建议先返回 404，仅对 local 规则集提供本接口，避免代理转发外部内容）。

### 二、系统 Host 设置（需求 2）

复用现有 key/value 全局设置体系（`storage.GlobalSettingStorage`、`service.*Setting`、前端 `SettingManage.tsx`）：

- 新增固定 key：`system_host`（与 `dns_config` 同级，定义为常量，建议放 `service` 包内）。
- 值为不带尾斜杠的基础地址，如 `https://config.example.com`，读取时统一 `strings.TrimRight(host, "/")`。
- 后端新增读取助手 `resolveSystemHost()`（参考 `service/generated.go:resolveDNSConfigJSON()`），未配置时返回空串。
- 前端在系统设置页提供独立编辑入口（输入框 + 说明文案），保存到 `system_host`；也允许通过通用 key/value 设置项编辑。

> `system_host` 是否纳入 `isReservedGlobalSettingKey`（`service/auth.go:131`）保护：建议**不纳入保留前缀**，让它和 `dns_config` 一样可被前端正常读写；仅 `auth.` 前缀保持保留。

### 三、生成时 local 规则集改为 URL 引用（需求 3）

新增 URL 拼接助手（建议放 `convert/common` 或 service 层透传）：

```text
rulesetURL(systemHost, tag, software) = systemHost + "/open/ruleset/" + tag + "/" + software
```

把 `systemHost` 透传进三条生成链路（与 `deviceCode` 透传方式一致）：

- `singbox.GetRoute(...)` → `baseRuleSets(...)`：增加 `systemHost` 入参；local 规则集分支由 `type:"inline"` 改为：
  ```go
  entity.SingRuleSet{Type:"remote", Tag:ruleSet.Tag, Format:"source", URL: rulesetURL(host, tag, "singbox"), DownloadDetour: ruleSet.DownloadDetour}
  ```
- `surge.Render(...)` / `shadowrocket.Render(...)`：`renderRuleSection()` 里 local 分支由 `expandLocalRuleSet()` 改为输出单行 `RULE-SET,<rulesetURL(host,tag,software)>,<policy>`。
- service 层在调用各 `Render` / `GetRoute` 前读取 `system_host` 并传入。

#### 兼容与降级

- **`system_host` 未配置时**：无法拼出绝对 URL，**回退到现有“展开/inline”行为**，保证旧部署零配置仍可用（实现上：host 为空 → 走原 local 展开分支）。
- remote 规则集：始终保持原 `URL` 引用，不受影响。
- 该改动让最终配置体积更小、规则集可被客户端独立缓存/更新，但引入“客户端需能访问 system_host”的运行期依赖，需在文档中说明。

## 影响范围

| 层 | 文件 / 模块 | 改动 |
|----|------------|------|
| 路由 | `cmd/server/main.go` | 注册 `GET /open/ruleset/:tag/:software` |
| 服务 | `service/service.go`（或新文件） | 新增 `GetRuleSetForSoftware` 处理器 |
| 服务 | `service/generated.go` | 新增 `system_host` 常量与 `resolveSystemHost()`，透传 host 到三条生成链路 |
| 转换 | `convert/common/`（新增助手） | `rulesetURL()` 拼接 + 规则 → 列表行（剥离 policy）的导出能力 |
| 转换 | `convert/singbox/route.go` | `GetRoute`/`baseRuleSets` 增加 `systemHost`，local 改 `remote`+url |
| 转换 | `convert/surge/surge.go` | `Render`/`renderRuleSection` local 改 `RULE-SET,url,policy`，并复用导出器 |
| 转换 | `convert/shadowrocket/shadowrocket.go` | 同 Surge |
| 前端 | `web/src/pages/SettingManage.tsx`（或专用编辑） | 系统 Host 编辑入口 |
| 文档 | `docs/reference/api-reference.md` | 新增 `/open/ruleset/:tag/:software` 接口说明、`system_host` 设置说明 |
| 文档 | `docs/modules/rule-set.md` | 补充“规则集 URL 引用模式”与系统 Host 依赖说明 |
| 文档 | `docs/reference/configuration.md` | 补充 `system_host` 配置项 |

## 待确认问题

1. URL 顺序：采用 `/open/ruleset/:tag/:software`（tag 在前，避免参数名同层冲突），可否接受？
2. `system_host` 未配置时回退到“展开”行为，是否符合预期（而不是直接报错）？
3. remote 规则集是否也要走本服务的 `/open/ruleset` 中转（默认不走，保持外部 URL 原样）？
4. 软件名称取值是否就用 `singbox` / `surge` / `shadowrocket`（与现有 `/open/*` 命名一致）？

## 验收标准

1. `GET /open/ruleset/:tag/singbox` 返回该规则集的 sing-box source JSON；`/surge`、`/shadowrocket` 返回对应的规则列表文本（每行 `类型,值`，不含 policy）；非法 software 返回 400，未知 tag 返回 404。
2. 系统设置可读写 `system_host`，前端有独立编辑入口，值持久化到全局设置存储（数据库 / Supabase / JSON 文件均 round-trip）。
3. 配置 `system_host` 后：sing-box 生成结果中 local 规则集为 `type:"remote"` 且 `url` 指向 `…/open/ruleset/<tag>/singbox`；Surge / Shadowrocket 中 local 规则集为单行 `RULE-SET,…/open/ruleset/<tag>/<software>,<policy>`，不再有逐条展开。
4. 未配置 `system_host` 时，三种输出与改动前完全一致（local 规则集仍按展开/inline 渲染），回归无差异。
5. remote 规则集在三种输出里行为不变。
6. 新接口与生成改动均补充单元测试（含 host 为空的降级分支）。
7. `go test ./...` 与前端 `tsc --noEmit` 通过。
