# sing-box 配置结构

公开生成接口 `GET /open/generate/:device` 返回的根结构定义在 [entity/singbox.go](/Users/xsdhy/data/code/go/singboxconfig/entity/singbox.go)。

## 顶层结构

当前返回 JSON 的顶层字段是：

```json
{
  "dns": {},
  "endpoints": [],
  "route": {},
  "experimental": {},
  "inbounds": [],
  "outbounds": []
}
```

对应 Go 结构：

- `dns`
- `endpoints`
- `route`
- `experimental`
- `inbounds`
- `outbounds`

## `dns`

类型：

- `entity.SingDNS`

字段：

- `servers`
- `rules`
- `final`

来源：

- 优先读取 `GlobalSettings["dns_config"]`
- 为空或 JSON 非法时回退到内置默认值

## `endpoints`

类型：

- `[]entity.SingEndpointWireguard`

当前用途：

- 只用于 WireGuard endpoint 输出

生成条件：

- 设备绑定了 `wireGuardTag`
- 找到了对应 WireGuard 模板
- 设备有 `wireGuardClientAddr`
- 设备有 `wireGuardClientKey`

字段：

- `type`
- `tag`
- `mtu`
- `address`
- `private_key`
- `peers`

## `route`

类型：

- `entity.SingRoute`

字段：

- `rules`
- `rule_set`
- `final`
- `find_process`
- `auto_detect_interface`

当前固定值：

- `final = "general"`
- `find_process = false`
- `auto_detect_interface = true`

规则由两部分组成：

- 基础规则：DNS 劫持、`clash_mode`、QUIC 拒绝
- 按 `sort` 排序后的规则集引用；其 `outbound` 不在最终出站列表（含节点分组出站与 `direct`）中时跳过该条，`final` 兜底不参与校验

## `experimental`

类型：

- `*entity.SingExperimental`

注意：

- 该字段带 `omitempty`
- 返回 `nil` 时不会出现在 JSON 中

当前逻辑：

- 对 `phone` 返回 `nil`
- 对 `tv` 使用固定 `192.168.10.66:9090`
- 其他设备默认使用 `127.0.0.1:9090`

这部分仍带有历史设备常量逻辑，不完全来自后台配置。

## `inbounds`

类型：

- `[]entity.SingInbound`

来源：

1. 先取设备绑定的 Inbound tag 列表
2. 再从 Inbound 模板表中查找对应模板
3. 按 `sort` 排序
4. 反序列化每个 `configJson`
5. 跳过禁用或非法 JSON 模板

当前结构只声明了项目会输出的常见字段：

- `type`
- `tag`
- `listen`
- `listen_port`
- `address`
- `inet4_address`
- `auto_route`
- `stack`
- `sniff`

## `outbounds`

类型：

- `[]entity.SingBoxOut`

来源由三部分合并：

1. 订阅解析得到的节点
2. 额外出站模板
3. 节点分组生成的 `selector` / `urltest`
4. 固定追加 `{ "type": "direct", "tag": "direct" }`

当前 `SingBoxOut` 是一个大而宽的兼容结构，覆盖多种协议需要的字段，例如：

- `type`
- `tag`
- `server`
- `server_port`
- `uuid`
- `password`
- `tls`
- `transport`
- `outbounds`
- `url`
- `interval`
- `tolerance`

## 生成链路中的过滤规则

### 规则集设备过滤

`route.rule_set` 和 `route.rules` 都会按 `RuleSet.AbleDevices` 过滤。

当前实现使用 `strings.Contains`，不是严格的逗号分隔精确匹配。

### 额外出站设备过滤

`visibleDevices` 为空时全部设备可见；否则按逗号分隔后做精确匹配。

### Inbound 过滤

只有被设备绑定且模板启用的 Inbound 才会进入结果。

## 与官方 sing-box 文档的关系

本项目输出结构尽量贴近 sing-box JSON，但不是完整镜像：

- 只实现了当前项目生成链路需要的字段
- 某些字段名称直接沿用 sing-box 规范
- 某些结构是项目内部为了兼容多协议做的聚合表示

因此本文件适合用于理解“本项目会生成什么”，不适合作为 sing-box 官方完整 schema 替代品。

## 相关文档

- [配置生成流程](../architecture/config-generation.md)
- [设备管理](../modules/device.md)
- [WireGuard 管理](../modules/wireguard.md)
