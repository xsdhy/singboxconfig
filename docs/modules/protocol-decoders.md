# 协议解码器

## 模块职责

协议解码器负责把订阅中的节点 URL 转换为统一的 `entity.SingBoxOut`。这是订阅管理和 sing-box / Surge / Shadowrocket 多格式生成之间的桥接层。

代码入口：

- 订阅调度：`service/outbound_cache.go`（`subscriptionOutboundConvertMap`）
- 协议实现：`protocol/ss.go`、`protocol/ssr.go`、`protocol/trojan.go`、`protocol/vmess.go`、`protocol/vless.go`

## 总体架构

生成链路不会直接解析所有节点，而是先读取 URL scheme，再根据 `subscriptionOutboundConvertMap` 分发：

- `ss` -> `DecodeSSURLToSingBox`
- `ssr` -> `DecodeSSRURLToSingBox`
- `trojan` -> `DecodeTrojanUrlToSingBox`
- `vmess` -> `DecodeVmessUrlToSingBox`
- `vless` -> `DecodeVlessUrlToSingBox`

统一流程：

1. 识别协议前缀
2. 解析协议特定字段
3. 转换为 `entity.SingBoxOut`
4. 返回给 `GetOutbounds()` 汇总

## Shadowsocks (SS)

实现文件：`protocol/ss.go`

支持行为：

- 识别 `ss://`
- 支持多种 Base64 解码方式
- 解析 `method:password`
- 解析 `host:port`
- 尝试读取 fragment 作为节点标签

转换结果主要字段：

- `type: "shadowsocks"`
- `server`
- `server_port`
- `method`
- `password`

当前限制：

- 默认 `network` 固定写成 `tcp`
- plugin / plugin_opts 预留但未实际解析

## ShadowsocksR (SSR)

实现文件：`protocol/ssr.go`

支持行为：

- 支持 `ssr://base64(...)`
- 也支持直接格式 `ssr://server:port:...`
- 解析 `protocol`、`method`、`obfs`、`password`
- 解析 `obfsparam`、`protoparam`、`remarks`、`group`

转换结果主要字段：

- `type: "shadowsocksr"`
- `server`
- `server_port`
- `method`
- `password`
- `protocol`
- `protocol_param`
- `obfs`
- `obfs_param`

注意：该解码器已进入订阅生成链路。Surge 输出会因为客户端能力限制跳过 SSR；Shadowrocket 输出会导出 SSR。

## Trojan

实现文件：`protocol/trojan.go`

支持行为：

- 解析标准 URL 结构
- 读取密码、主机、端口
- 解析 `security`、`sni`、`allowInsecure`、`peer`、`type`、`host`
- 使用 URL fragment 作为标签

转换特点：

- 输出 `type: "trojan"`
- 默认启用 TLS
- 若设置了 `peer`，会用它覆盖 `tls.server_name`
- 未指定网络类型时回退为 `tcp`

当前实现里 `allowInsecure` 字段虽然被解析出来，但转换阶段会直接把 `tls.insecure` 设为 `true`，并不是严格依照输入值决定。

## VMess

实现文件：`protocol/vmess.go`

支持行为：

- 解析 `vmess://base64(json)`
- 读取地址、端口、UUID、alterId、网络类型、Host、Path、TLS 等字段
- 支持 `ws`、`http/h2`、`grpc`、`quic`、`kcp`、`tcp`

转换特点：

- 输出 `type: "vmess"`
- 默认 `security=auto`
- `tls=tls` 时生成 TLS 配置
- 根据 `net` 生成 `transport`

标签规则：

- 优先使用 `ps`
- 清洗后为空时回退为 `vmess-node`

## 标签清洗规则

`cleanTag()` 定义在 `protocol/trojan.go`，被多个协议复用。它会：

- 去掉换行
- 去掉首尾空白
- 移除大部分标点、符号和 emoji

优点是生成的 tag 更适合作为 sing-box 配置键；代价是可能损失原始订阅备注中的地区旗帜、分隔符等信息。

## 错误处理策略

解码器普遍采用“单节点失败即跳过”的策略：

- URL 格式错误：返回 error
- 必要字段缺失：返回 error
- 非法 Base64 / JSON：返回 error

上层 `GetOutbounds()` 不会因为单个节点失败而中断整个订阅解析。

## VLESS

实现文件：`protocol/vless.go`

支持行为：

- 解析标准 URL 结构：`vless://<uuid>@<host>:<port>?<params>#<tag>`
- UUID 位于 URL userinfo 部分（@ 符号前）
- 支持三种安全层：`reality`、`tls`、`none`（或空）
- Reality 安全层：解析公钥（`pbk`）、ShortID（`sid`）、uTLS 指纹（`fp`）
- TLS 安全层：解析 SNI（`sni` 优先，其次 `servername`）、uTLS 指纹（`fp`）
- 支持传输协议：`tcp`（无需额外配置）、`ws`、`grpc`、`httpupgrade`
- 使用 URL fragment 作为节点标签，经 `cleanTag` 清洗后输出

转换结果主要字段：

- `type: "vless"`
- `server` / `server_port`
- `uuid`
- `flow`（xtls-rprx-vision 等）
- `tls`（含 `reality` 和 `utls` 子配置）
- `transport`（ws / grpc / httpupgrade 时生成）

注意事项：

- VLESS URL 中的 `mode=multi` 参数为 Xray 特有，与 sing-box 多路复用无关，解析时忽略
- `spx`（Spider X path）为 Reality 服务端参数，客户端出站配置无需设置，解析时忽略
- SNI 优先读取 `sni` 参数，其次读取 `servername` 参数（两者语义相同）

## 扩展新协议的方法

如果要新增一个订阅协议，通常需要四步：

1. 在 `protocol/` 新增解析结构和 `DecodeXxxURL`
2. 实现 `ConvertXxxToSingBox`
3. 提供 `DecodeXxxURLToSingBox`
4. 在 `service/outbound_cache.go` 的 `subscriptionOutboundConvertMap` 里注册 scheme

如果只完成前 1-3 步而没有注册到 `subscriptionOutboundConvertMap`，协议实现不会进入实际生成链路。

## 当前限制

- 缺少统一的解码器接口抽象，目前靠函数表调度
- 不做节点去重
- 各协议字段映射覆盖度有限，只支持项目当前需要的一部分 sing-box 参数

## 适合更新本文档的场景

- 新增订阅协议
- 修改字段映射规则
- 调整标签清洗逻辑
- 把新协议接入 `subscriptionOutboundConvertMap`
