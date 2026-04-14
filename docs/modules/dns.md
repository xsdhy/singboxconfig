# DNS 配置

## 模块职责

DNS 模块负责维护 sing-box 根配置中的 `dns` 段。它没有单独的数据库表，而是把整段 DNS JSON 保存在全局设置 `dns_config` 中。

代码入口：

- sing-box DNS 结构：`entity/singbox.go`
- 管理页面：`web/src/pages/DnsManage.tsx`
- 全局设置接口：`service/service.go`
- 生成转换：`convert/singbox/dns.go`

## 存储方式

DNS 配置通过 `GlobalSettings` 保存：

- key：`dns_config`
- value：完整 JSON 字符串

前端 DNS 页面本质上是一个针对 `dns_config` 的专用 JSON 编辑器，并不是单独的后端 DNS 资源。

涉及接口：

- `GET /api/settings/key/dns_config`
- `POST /api/settings`
- `PUT /api/settings/:key`

当前页面保存时会先尝试更新，若不存在则回退到创建。

## 数据结构

后端使用以下结构承载 DNS 配置：

- `entity.SingDNS`
- `entity.SingDNSServer`
- `entity.SingDNSRule`

核心字段：

- `servers`
- `rules`
- `final`

这些字段尽量贴近 sing-box 官方 JSON，便于直接序列化输出。

## 默认 DNS

当 `dns_config` 不存在或内容非法时，生成链路会回退到 `GetDefaultDNS()`。默认配置包含：

- `dns_proxy`
- `dns_direct`
- `dns_block`
- `dns_resolver`

默认规则会引用：

- `cnip`
- `cnsite`

因此如果实际部署中没有这两个规则集，sing-box 侧是否报错取决于运行时对缺失 rule-set 的容忍度。项目代码本身不做额外校验。

## 生成逻辑

生成时 `service.Generated()` 会：

1. 读取 `GlobalSettings["dns_config"]`
2. 若 key 不存在，则把空字符串交给下游
3. 调用 `singbox.ResolveDNS(configJSON)`

`ResolveDNS()` 规则：

- 为空：返回默认 DNS
- JSON 非法：记录 warning，返回默认 DNS
- JSON 合法：按保存值直接使用

这意味着 DNS 配置损坏不会阻断整个设备配置生成，只会触发回退。

## 前端管理行为

`web/src/pages/DnsManage.tsx` 的特点：

- 默认加载 `dns_config`
- 若读取失败，编辑器展示内置默认模板
- “恢复默认”只覆盖编辑区，不会立即写回后端
- 保存前会先做 JSON 规范化

这让 DNS 页面更像“专用设置编辑器”，而不是 CRUD 列表页。

## 与其他模块的关系

- DNS 规则里常引用规则集 tag，如 `cnip`、`cnsite`
- DNS server 的 `detour` 常引用节点分组或固定出站，如 `general`、`direct`

这些引用当前不做存在性校验，因此文档和运维配置需要额外注意一致性。

## 当前限制

- 没有独立的 DNS REST 资源模型
- 不做结构化字段级校验，保存层只是字符串
- 默认 DNS 与前端默认模板需要人工保持一致
- 规则集引用和 detour 引用都不做预检查

## 适合更新本文档的场景

- DNS 改为独立资源表
- 调整默认 DNS 模板
- 增加后端 schema 校验
- 引入设备级 DNS 覆盖
