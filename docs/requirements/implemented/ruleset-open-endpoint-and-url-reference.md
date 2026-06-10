# 需求：规则集独立 open 接口 + 系统 Host 设置 + 生成时规则集改为 URL 引用

- 状态：已实现（implemented）
- 提出时间：2026-06-10
- 完成时间：2026-06-11
- 关联模块：[规则集管理](../../modules/rule-set.md)、[配置生成流程](../../architecture/config-generation.md)、[API 接口列表](../../reference/api-reference.md)、[设备管理](../../modules/device.md)

## 背景

当前 `/open` 下只有“整份配置”接口，按设备输出各软件的**完整配置**：

| 路由 | 处理器 | 输出 |
|------|--------|------|
| `GET /open/generate/:device` | `service.Generated` | sing-box 完整 JSON |
| `GET /open/surge/:device` | `service.SurgeGenerated` | Surge 完整文本 |
| `GET /open/shadowrocket/:device` | `service.ShadowrocketGenerated` | Shadowrocket 完整文本 |
| `GET /open/ruleset/:tag` | `service.GetRuleSetByTag` | 规则集原始 `content`（JSON） |

注：`cmd/server/main.go:110-113` 注册以上路由。`/open/ruleset/:tag` 目前只是把 `RuleSet.Content` 原样反序列化回吐，**不区分软件格式、不校验设备与 token**，本质上是一个历史兼容接口。

规则集（`entity.RuleSet`，见 `entity/rule_set.go`）分两类：

- **remote**：带 `URL` 字段，生成时直接以 URL 引用。
  - sing-box：`convert/singbox/route.go:baseRuleSets()` 输出 `type:"remote"` + `url`。
  - Surge / Shadowrocket：输出一行 `RULE-SET,<url>,<policy>`（`convert/surge/surge.go:506-512`、`convert/shadowrocket/shadowrocket.go:648-654`）。
- **local / inline**：把 `Content` 中的 sing-box source 规则在生成时**展开**。
  - sing-box：`baseRuleSets()` 输出 `type:"inline"`，规则内联进配置（`convert/singbox/route.go:131-141`）。
  - Surge：`expandLocalRuleSet()` 把每条规则拆成 `DOMAIN/DOMAIN-SUFFIX/IP-CIDR/...` 多行（`convert/surge/surge.go:514,544`）。
  - Shadowrocket：同上（`convert/shadowrocket/shadowrocket.go:656,686`）。

也就是说，**local / inline 规则集在三种输出里都是“展开成一条条具体规则”塞进最终配置**，而非作为一个独立的规则集 URL 引用。

本需求合并并取代 `docs/requirements/planned/open-rules-endpoint.md` 中关于多客户端规则集输出接口的路径设计，采用“规则集级 open 接口 + 设备 token 鉴权 + 生成时引用该接口”的完整方案。

## 需求描述

本次包含三个改动：

### 需求 1：新增“规则集级” open 接口

在现有 `/open` 整份配置接口之外，**新增一类按规则集输出的 open 接口**，能针对某个规则集、按指定软件格式输出该规则集的规则内容（不是整份配置）。

- 新接口按设备鉴权，复用完整配置接口的设备解析、启用状态与 token 校验。
- 新接口校验 `RuleSet.AbleDevices`，不可见时按 404 处理，避免泄露规则集存在性。
- 软件名称第一版覆盖 `singbox` / `surge` / `shadowrocket` 三种。
- 返回的是“该规则集在目标软件下的规则集文件内容”，可被对应客户端直接作为远程规则集加载。

### 需求 2：系统设置增加“系统 Host”

在系统设置里新增一项 **系统 Host（system host）**，用于记录本服务对外可访问的基础地址（如 `https://config.example.com`）。该值用于拼接需求 1 接口的**绝对 URL**，供需求 3 在生成配置时引用。

### 需求 3：生成配置时，local / inline 规则集由“展开”改为“URL 引用”

把生成链路里 local / inline 规则集的处理方式，从**展开成一条条规则**改为**输出为一个规则集 URL 引用**。该 URL 指向需求 1 的新接口，并用需求 2 的系统 Host 拼成绝对地址。

- sing-box：local / inline 规则集由 `type:"inline"` 改为 `type:"remote"` + 指向本服务的 `url`。
- Surge / Shadowrocket：local / inline 规则集由多行展开改为单行 `RULE-SET,<url>,<policy>`。
- remote 规则集行为不变（本就是 URL 引用）。

## 方案设计

### 一、规则集 open 接口（需求 1）

#### 路由

新增：

```text
GET /open/rules/:tag/:software/:device?token=...
```

示例：

```text
/open/rules/geosite-cn/singbox/iphone15?token=xxx
/open/rules/geosite-cn/surge/iphone15?token=xxx
/open/rules/geosite-cn/shadowrocket/iphone15?token=xxx
```

路由说明：

- `:tag`：规则集唯一标识（`RuleSet.Tag`）。
- `:software`：目标软件，取值为 `singbox` / `surge` / `shadowrocket`；其余值返回 400。
- `:device`：设备编码，用于复用完整配置接口的设备解析、启用状态和 token 校验。
- `?token=`：设备 token，必须与 `Device.Token` 一致。
- 现有 `GET /open/ruleset/:tag` 保留不动，仅作为历史兼容接口；新功能不再扩展该路径，避免继续放大无鉴权接口的能力。

鉴权与可见性：

1. 通过 `:device` 解析设备，复用 `resolveGenerateDevice(deviceCode)`。
2. 设备不存在返回 404；设备禁用返回 403；token 不匹配返回 401（与现有完整配置接口保持一致）。
3. 通过 `:tag` 查找规则集，未命中返回 404。
4. 校验规则集对当前设备可见：`AbleDevices` 为空表示全部可见，非空时沿用当前 `strings.Contains(ableDevices, deviceCode)` 逻辑；不可见返回 404。

#### 输出格式（按软件）

新增处理器 `service.GetRulesBySoftware`（命名可调整），按 `:software` 输出：

| software | Content-Type | 内容 |
|----------|--------------|------|
| `singbox` | `application/json` | sing-box source 规则集 JSON，规范化为 `{"version":1,"rules":[...]}` |
| `surge` | `text/plain; charset=utf-8` | Surge 规则集文件：每行 `类型,值`（**不含 policy 列**） |
| `shadowrocket` | `text/plain; charset=utf-8` | Shadowrocket 规则集文件：每行 `类型,值`（**不含 policy 列**） |

规则集文件里不带出站策略，策略仍在主配置的 `RULE-SET,<url>,<policy>` 行里指定。

#### local / inline 内容规范化

新增 `convert/ruleset/` 公共包，统一承载规则集解析与渲染能力，避免继续复制 `convert/surge` 与 `convert/shadowrocket` 中的规则展开逻辑。

`RuleSet.Content` 支持以下输入形态：

1. 完整 sing-box source 规则集：`{"version":1,"rules":[...]}`
2. 裸 rules 数组：`[{...}, {...}]`
3. 单条规则对象：`{"domain_suffix":["example.com"]}`

规范化输出规则：

- sing-box：统一输出 `{"version":1,"rules":[...]}`。
- Surge / Shadowrocket：把规范化后的 rules 转成逐行列表。
- `Content` 整体不是合法 JSON 时，接口返回 500；生成整份配置时不能生成指向坏内容的远程 URL，应按兼容降级策略先回退到原展开逻辑，失败后再跳过并记录 warning。

#### Surge / Shadowrocket 字段映射

第一版支持现有展开逻辑已覆盖的字段，并把策略列剥离：

| sing-box source 字段 | 列表行 |
|----------------------|--------|
| `domain` | `DOMAIN,<value>` |
| `domain_suffix` | `DOMAIN-SUFFIX,<value>` |
| `domain_keyword` | `DOMAIN-KEYWORD,<value>` |
| `domain_regex` | `DOMAIN-REGEX,<value>` |
| `ip_cidr` | IPv4 输出 `IP-CIDR,<value>`，IPv6 输出 `IP-CIDR6,<value>` |
| `geoip` | `GEOIP,<UPPERCASE_VALUE>` |

不支持字段第一版跳过并记录 warning，不中断整体输出。若后续确认 Surge / Shadowrocket 对 `no-resolve` 或其它参数有强需求，再单独扩展映射规则并补测试。

#### remote 规则集处理

remote 规则集生成行为不变：完整配置继续引用 `RuleSet.URL`，不经过本服务中转。

新接口第一版仅服务 local / inline 规则集：

- 请求 remote 规则集的 `/open/rules/:tag/:software/:device` 返回 404 或 400（实现时二选一并写入 API 文档，建议 400，错误信息说明 remote 规则集已有原始 URL，不做代理转换）。
- 不下载、不代理、不转换 remote 原始内容，避免引入外部请求、缓存和 SSRF 风险。

### 二、系统 Host 设置（需求 2）

复用现有 key/value 全局设置体系（`storage.GlobalSettingStorage`、`service.*Setting`、前端 `SettingManage.tsx`）：

- 新增固定 key：`system_host`（与 `dns_config` 同级，定义为常量，建议放 `service` 包内）。
- 值为不带尾斜杠的基础地址，如 `https://config.example.com`。
- 读取时统一 `strings.TrimRight(strings.TrimSpace(host), "/")`。
- 后端新增读取助手 `resolveSystemHost()`（参考 `service/generated.go:resolveDNSConfigJSON()`），未配置时返回空串。
- 建议校验 `system_host` 必须是合法绝对 URL，且 scheme 仅允许 `http` / `https`；非法值在生成时视为未配置并记录 warning，或在保存时直接拒绝（二选一并保持前后端一致）。
- 前端在系统设置页提供独立编辑入口（输入框 + 说明文案），保存到 `system_host`；也允许通过通用 key/value 设置项编辑。

> `system_host` 不纳入 `isReservedGlobalSettingKey` 保护，让它和 `dns_config` 一样可被前端正常读写；仅 `auth.` 前缀保持保留。

### 三、生成时 local / inline 规则集改为 URL 引用（需求 3）

新增 URL 拼接助手（建议放 `convert/common` 或 `convert/ruleset`，也可以由 service 层构造后透传）：

```text
rulesetURL(systemHost, tag, software, device, token)
  = trimRight(systemHost, "/")
  + "/open/rules/"
  + pathEscape(tag)
  + "/"
  + software
  + "/"
  + pathEscape(device)
  + "?token="
  + queryEscape(token)
```

要求：

- `tag` 和 `device` 必须使用 `url.PathEscape`。
- `token` 必须使用 query escaping。
- `systemHost` 必须先 trim 尾斜杠。
- 生成出的 URL 是客户端可直接访问的新规则集接口地址。

把 `systemHost` 与设备 token 透传进三条生成链路：

- `singbox.GetRoute(...)` → `baseRuleSets(...)`：增加 `systemHost` / `deviceToken` 入参。local / inline 规则集在 host 可用且内容可解析时输出：
  ```go
  entity.SingRuleSet{
      Type:           "remote",
      Tag:            ruleSet.Tag,
      Format:         "source",
      URL:            rulesetURL(host, ruleSet.Tag, "singbox", device.Code, device.Token),
      DownloadDetour: ruleSet.DownloadDetour,
  }
  ```
- `surge.Render(...)` / `shadowrocket.Render(...)`：增加 `systemHost` / `deviceToken` 入参。`renderRuleSection()` 里 local / inline 分支在 host 可用且内容可解析时输出单行 `RULE-SET,<rulesetURL(...)>,<policy>`。
- service 层在调用各 `Render` / `GetRoute` 前读取 `system_host` 并传入。

#### 兼容与降级

- **`system_host` 未配置或非法时**：无法拼出绝对 URL，回退到现有“展开/inline”行为，保证旧部署零配置仍可用。
- **local / inline `Content` 非法时**：
  - host 为空：保持现有行为，sing-box inline 跳过坏规则集，Surge / Shadowrocket 展开时记录 warning 并跳过。
  - host 非空：不要生成指向坏内容的远程规则集 URL；该规则集应回退到原展开逻辑或跳过，具体策略需保持三条输出一致。建议回退到原展开逻辑，失败后再跳过。
- **remote 规则集**：始终保持原 `URL` 引用，不受影响。
- **设备 token 轮换**：因为生成出的规则集 URL 带设备 token，设备 token 修改后，旧整份配置中的规则集 URL 会失效；这是预期行为，用户需重新拉取整份配置。
- **可访问性依赖**：客户端必须能访问 `system_host`；部署文档需说明反向代理、HTTPS 和公网/内网访问要求。

#### 有效规则集过滤

sing-box 当前会先生成 `route.rule_set`，再根据出站是否存在决定是否追加 `route.rules`。改为远程 URL 后，应避免输出“未被任何有效路由引用、但仍会被客户端下载”的 local 规则集。

实现时建议先计算当前设备真正有效的规则集列表：

1. 通过 `AbleDevices` 过滤。
2. 通过当前设备最终 outbounds 校验 `RuleSet.Outbound` 是否存在。
3. 仅对有效规则集同时输出 `route.rule_set` 与 `route.rules`。

Surge / Shadowrocket 当前已经在 `renderRuleSection()` 中先校验 policy 是否存在，再输出规则行；保持该行为即可。

## 影响范围

| 层 | 文件 / 模块 | 改动 |
|----|------------|------|
| 路由 | `cmd/server/main.go` | 注册 `GET /open/rules/:tag/:software/:device`；保留旧 `GET /open/ruleset/:tag` |
| 服务 | `service/service.go`（或新文件） | 新增 `GetRulesBySoftware` 处理器，复用设备解析、token 鉴权与规则集可见性校验 |
| 服务 | `service/generated.go` | 新增 `system_host` 常量与 `resolveSystemHost()`，透传 host 与 device token 到三条生成链路 |
| 转换 | `convert/ruleset/`（新增） | 规则集内容解析、规范化、各软件规则列表渲染、URL 拼接辅助能力 |
| 转换 | `convert/singbox/route.go` | `GetRoute` / `baseRuleSets` 增加 `systemHost` 与 token 入参；local / inline 在 host 可用时改 remote URL 引用 |
| 转换 | `convert/surge/surge.go` | `Render` / `renderRuleSection` 增加 `systemHost` 与 token 入参；local / inline 在 host 可用时改 `RULE-SET,url,policy` |
| 转换 | `convert/shadowrocket/shadowrocket.go` | 同 Surge |
| 前端 | `web/src/pages/SettingManage.tsx`（或专用编辑） | 系统 Host 编辑入口 |
| 文档 | `docs/reference/api-reference.md` | 新增 `/open/rules/:tag/:software/:device` 接口说明、`system_host` 设置说明 |
| 文档 | `docs/modules/rule-set.md` | 补充“规则集 URL 引用模式”、鉴权模型与系统 Host 依赖说明 |
| 文档 | `docs/reference/configuration.md` | 补充 `system_host` 配置项 |

## 待确认问题

1. remote 规则集请求新接口时返回 400 还是 404？建议 400，并明确“不做代理转换”。
   - **结论（已实现）**：返回 `400`，错误信息说明 remote 规则集已有原始 URL，不做代理转换。
2. `system_host` 非法值是在保存时拒绝，还是生成时降级并记录 warning？建议保存时拒绝，通用 key/value 编辑也做同样校验。
   - **结论（已实现）**：保存时拒绝（`normalizeGlobalSettingValue` 对 `system_host` 校验，非法返回 400）；生成时 `resolveSystemHost` 再做一次防御性降级。
3. local / inline `Content` 非法且 host 已配置时，是先回退展开后跳过，还是直接跳过？建议先回退展开，确保与旧行为尽量一致。
   - **结论（已实现）**：先回退到原展开/内联逻辑，展开失败再跳过并记录 warning，三条输出一致。
4. 是否要在前端规则集管理页提供“复制当前设备规则集 URL”的入口？第一版不是必须，但对验证和使用会更友好。
   - **结论（已实现）**：规则集列表页 local 规则集卡片新增「复制地址」按钮，弹窗内选择设备后按 sing-box / Surge / Shadowrocket 三种软件复制 `/open/rules/:tag/:software/:device?token=` 绝对地址（前缀取系统 Host）。

## 验收标准

1. ✅ `GET /open/rules/:tag/singbox/:device?token=...` 返回该 local / inline 规则集的 sing-box source JSON（规范化为 `{"version":1,"rules":[...]}`）。
2. ✅ `GET /open/rules/:tag/surge/:device?token=...` 与 `/shadowrocket/` 返回对应规则列表文本，每行 `类型,值`，不含 policy。
3. ✅ 新接口非法 software 返回 400；未知 tag 返回 404；设备不存在返回 404；设备禁用返回 403；token 不匹配返回 401；规则集对设备不可见返回 404。
4. ✅ 请求 remote 规则集的新接口按约定返回 400，不下载、不代理、不转换外部 URL。
5. ✅ 系统设置可读写 `system_host`，前端有独立编辑入口，值持久化到全局设置存储（复用既有 key/value 体系，各存储后端 round-trip）。
6. ✅ `system_host` 保存或读取时会去掉尾斜杠；非法 URL 在保存时拒绝（前后端一致校验）。
7. ✅ 配置合法 `system_host` 后：sing-box 生成结果中有效 local / inline 规则集为 `type:"remote"`、`format:"source"`，且 `url` 指向 `.../open/rules/<escaped-tag>/singbox/<escaped-device>?token=<escaped-token>`。
8. ✅ 配置合法 `system_host` 后：Surge / Shadowrocket 中有效 local / inline 规则集为单行 `RULE-SET,.../open/rules/<escaped-tag>/<software>/<escaped-device>?token=<escaped-token>,<policy>`，不再逐条展开。
9. ✅ `tag`、`device`、`token` 中包含空格、斜杠、问号等特殊字符时，生成 URL 正确 escape，接口能正常解析（`BuildRuleSetURL` 单测覆盖）。
10. ✅ 未配置或非法 `system_host` 时，三种输出与改动前保持兼容：local / inline 规则集仍按展开或 inline 渲染，remote 规则集行为不变。
11. ✅ remote 规则集在三种完整配置输出里行为不变。
12. ✅ sing-box 不输出因 `Outbound` 不存在而不会被 `route.rules` 引用的 local / inline 远程规则集条目（先过滤有效规则集，再统一输出 rule_set 与 rules）。
13. ✅ 新接口可见性/枚举解析、URL 拼接、规则规范化、host 为空降级、host 非空 URL 引用、特殊字符 escape、`system_host` 校验均补充单元测试。
14. ✅ `go test ./...` 与前端 `tsc --noEmit` 通过。
