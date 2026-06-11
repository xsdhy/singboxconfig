# Surge 配置输出

> 状态：已实现

## 背景

当前系统的全部价值都通过一个出口暴露：`GET /open/generate/:device?token=...` 输出 sing-box JSON 配置。但系统真正稀缺、难以复制的能力其实与输出格式无关，集中在数据管理层：

- 订阅源管理与多协议解析（SS/SSR/Trojan/VMess/VLESS）
- Outbound 缓存与按设备可见性过滤
- 节点分组筛选（`outboundGroupRuleFilter` 的 include/exclude 关键字逻辑）
- 规则集（inline/local 与 remote）
- 设备身份与 token 鉴权

这一整套数据层与 sing-box 没有任何耦合，理论上可以渲染成任意客户端的配置格式。

部分用户使用 Surge 作为客户端，希望直接从本系统获取 Surge 配置，而不必再额外维护一份。因此本需求的目标是：**在不改动数据层的前提下，新增一条 Surge 配置输出链路**。

## 目标

- 新增 `GET /open/surge/:device?token=...` 接口，输出 Surge 配置文本
- 完整复用现有数据层：设备解析、鉴权、订阅拉取与缓存、Outbound 可见性过滤、节点分组筛选、规则集
- 复用现有「局部脏数据可降级、不中断整体生成」的错误处理风格
- SS / Trojan 节点做完整、无损映射
- VMess 做 best-effort 映射（仅 Surge 4+ 适用）
- Surge 不支持的协议（VLESS / Hysteria / Hysteria2 / TUIC）跳过并记录 warning

## 非目标

- 移除或替换现有 sing-box 输出（两套输出并存，互不影响）
- 第一版导出 inbound（TUN/HTTP/SOCKS）（语法迥异，且对 Surge 客户端场景通常非必需，留待后续按需补充）

> WireGuard endpoint 已在后续增量中支持导出（见协议支持矩阵与「设计细节」）。
- 把 sing-box 已存的 `.srs` 二进制远程规则集转换为 Surge 规则格式（remote 规则集仅做 `RULE-SET` 直引，不做格式转换）
- 前端完整的 Surge 链接管理界面（可作为后置增量）
- 在 Surge 端复刻 sing-box 的 DNS 分流细节，仅输出基础 DNS 配置

## 协议支持矩阵

| 协议 | sing-box（现状） | Surge | 本需求处理 |
|------|------|------|------|
| Shadowsocks (ss) | ✅ | ✅ | 完整映射 |
| Trojan | ✅ | ✅ | 完整映射 |
| HTTP / HTTPS | ✅ | ✅ | 完整映射（依据 `tls.enabled` 区分 http/https） |
| SOCKS (socks5) | ✅（`type: socks`，默认 v5） | ✅（关键字 `socks5` / `socks5-tls`） | 完整映射（依据 `tls.enabled` 区分；显式 version≠5 时跳过 + warning） |
| VMess | ✅ | ⚠️ 仅 Surge 4+，功能受限 | best-effort，标注限制 |
| VLESS | ✅ | ❌ | 跳过 + warning |
| Hysteria / Hysteria2 | ✅ | ❌ | 跳过 + warning |
| TUIC | ✅ | ❌ | 跳过 + warning |
| WireGuard | ✅（endpoint） | ✅（语法不同） | 转换为 `wireguard` 代理 + `[WireGuard]` 段 |

> 已确认当前订阅节点以 **SS / Trojan 为主**，因此协议覆盖足以承载实际节点，跳过的协议不构成阻塞。

> 增量说明：在首版基础上后续补充了 **HTTP/HTTPS 节点导出**、**WireGuard endpoint 导出**、**SOCKS5 节点导出**（sing-box 的 `type: socks` 兼容映射为 Surge 的 `socks5` / `socks5-tls`），并把 `[General]` 的 `ipv6` 默认值改为 `false`（关闭 IPv6）。

## 方案结论

整体采用「新增平行输出链路」的方式，数据层零改动。

```
GET /open/generate/:device?token=...   → sing-box JSON（保留，c.JSON）
GET /open/surge/:device?token=...       → Surge 文本（新增，c.String）
```

### 1. service 层入口

新增一个仿照 `service/generated.go` 中 `Generated` 的处理函数：

- 复用 `resolveGenerateDevice`、设备启用判断、token 校验
- 复用 `resolveGenerateOutbounds(ctx, deviceCode)` 获取统一的 `entity.Outbound` 列表
- 复用 `storage.ListRuleSets()` 与按设备过滤逻辑
- 复用 DNS 全局配置读取
- 最后改为调用 `convert/surge` 渲染器，并以 `c.String(http.StatusOK, ...)` 输出，`Content-Type` 为 `text/plain`

### 2. 新增 `convert/surge/` 包

平行于 `convert/singbox/`，负责把统一中间数据渲染成 Surge 分段文本。Surge 是 INI 风格的分段文本格式，而非 JSON，因此需要一套独立的文本序列化器。

| Surge 段 | 数据来源 | 映射逻辑 |
|---------|---------|---------|
| `[General]` | 固定模板 + DNS 配置 | 基础选项与 `dns-server` |
| `[Proxy]` | `entity.Outbound`（SS/Trojan/VMess） | 每个节点一行：`名称 = ss, server, port, ...` |
| `[Proxy Group]` | `NodeGroup` + 复用 `outboundGroupRuleFilter` | `selector → select`、`urltest → url-test`，成员为节点名列表 |
| `[Rule]` | `RuleSet` | remote → `RULE-SET,url,组名`；inline/local → 展开为 `DOMAIN-SUFFIX,...,组名` 等具体规则行 |

### 3. 协议映射

- SS / Trojan：完整字段映射（含 TLS、传输层在 Surge 支持范围内的部分）
- VMess：best-effort 映射，超出 Surge 能力的字段忽略
- VLESS / Hysteria / Hysteria2 / TUIC：跳过该节点并记录 warning，不写入 `[Proxy]`，相应也不出现在分组成员中

### 4. 错误处理

沿用现有生成链路的降级策略：

**致命错误（直接返回错误响应）**
- 设备不存在、禁用、token 错误
- 存储层读取失败

**可降级错误（跳过局部，继续生成）**
- 单个节点协议不被 Surge 支持
- 单个节点关键字段缺失
- 单个本地规则集 `Content` 非法

### 5. 前端（可选，后置）

在设备管理页的 sing-box 链接旁，增加一个 Surge 订阅链接展示与复制入口。第一版可不做，仅提供后端接口。

## 设计细节

### 分组成员引用一致性

`[Proxy Group]` 的成员必须引用 `[Proxy]` 中实际存在的节点名。因此渲染顺序为：

1. 先渲染 `[Proxy]`，得到「成功导出的节点名集合」
2. 再渲染 `[Proxy Group]` 时，成员列表只保留集合内的节点名
3. 若某分组筛选后成员为空，则不输出该分组（与现有 sing-box 逻辑一致）

这样能避免被跳过的协议节点在分组里产生悬空引用。

### 规则集处理

- `RuleSetType == remote`：输出 `RULE-SET,<url>,<outbound>`，不做格式转换
- 其它（inline/local）：将 `Content` 中的规则展开为 Surge 规则行，按规则类型映射（如 `DOMAIN-SUFFIX`、`IP-CIDR` 等），非法 JSON 跳过并记录 warning
- 末尾追加兜底 `FINAL` 规则

### DNS

仅从全局 DNS 配置中提取上游地址输出为 `[General]` 的 `dns-server`，不在 Surge 端复刻 sing-box 的完整 DNS 分流规则。

### HTTP / HTTPS 节点

sing-box 的 `http` outbound 通过 `tls.enabled` 区分明文与加密，对应 Surge 的 `http` 与 `https` 两种关键字。导出时携带 `username` / `password`（按需），并复用 TLS 参数映射（`sni`、`skip-cert-verify`、`alpn`）。

### WireGuard endpoint

WireGuard 在 sing-box 中存放于 `endpoints`（而非 `outbounds`），Surge 输出链路复用 `resolveGenerateEndpoints(device)` 取出 endpoint，再由 `convert/surge` 转换：

- 每个 endpoint 产出一条 `名称 = wireguard, section-name=<名称>` 的 `[Proxy]` 引用行
- 同时输出一段独立的 `[WireGuard <名称>]` 配置：`private-key`、`self-ip` / `self-ip-v6`（按 IPv4 / IPv6 拆分客户端地址）、`mtu`、以及每个 `peer = (public-key=..., allowed-ips=..., endpoint=host:port, preshared-key=..., keepalive=...)`
  - `allowed-ips` 含多个值（逗号分隔）时整体加引号，避免内部逗号被 Surge 当成 peer 字段分隔符
- endpoint tag 注册进代理名集合，可被策略组与规则正常引用；缺少 tag / private-key / 可用 peer 时降级跳过并记录 warning

### IPv6 默认关闭

`[General]` 段固定输出 `ipv6 = false`，默认关闭 IPv6。

## 实现范围

### 后端

- [x] 新增 `GET /open/surge/:device?token=...` 路由与 service 处理函数
- [x] 新增 `convert/surge/` 包：`[General]` / `[Proxy]` / `[Proxy Group]` / `[Rule]` / `[WireGuard]` 渲染
- [x] SS / Trojan / HTTP(HTTPS) 完整映射，VMess best-effort，WireGuard endpoint 转换为 `wireguard` 代理，其余协议跳过 + warning
- [x] `[General]` 默认关闭 IPv6（`ipv6 = false`）
- [x] 复用 `resolveGenerateOutbounds`、`resolveGenerateEndpoints`、节点分组筛选、规则集过滤、设备鉴权
- [x] 单元测试：协议映射、HTTP/HTTPS、WireGuard endpoint、分组成员一致性、不支持协议跳过、规则集展开

### 前端（可选）

- [x] 设备页展示 Surge 订阅链接与复制入口

### 文档

- [x] 实现后将本文件移至 `requirements/implemented/`
- [x] 更新 `reference/api-reference.md`（新增 `/open/surge` 接口）
- [x] 视情况新增 `modules/` 下 Surge 输出说明，或并入配置生成相关文档
- [x] 更新 `INDEX.md` 中受影响的目录条目

## 验收结果

> 已完成后端、前端入口、文档同步与核心转换单元测试。

- [x] `/open/surge/:device?token=...` 能输出 Surge 客户端可直接导入的配置
- [x] SS / Trojan 节点正确映射，可正常连接
- [x] 节点分组正确生成 Surge Proxy Group，成员无悬空引用
- [x] 规则集正确转为 Surge 规则
- [x] 不支持的协议被跳过且不影响整体输出
- [x] 现有 sing-box 输出链路不受影响
