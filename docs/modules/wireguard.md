# WireGuard 管理

## 模块职责

WireGuard 模块负责维护可复用的 WireGuard endpoint 模板及其 peers，并在设备生成时与设备侧私钥、地址信息合成为 sing-box `endpoints`。

代码入口：

- 实体定义：`entity/device_management.go`
- 管理接口：`service/device_management.go`
- 生成转换：`convert/singbox/endpoint.go`

## 数据模型

### WireGuard

`entity.WireGuard`：

- `tag`：模板唯一标识
- `name`：展示名称
- `description`：说明
- `enabled`：是否启用
- `sort`：后台排序
- `endpointTag`：最终输出到 sing-box endpoint 的 `tag`
- `mtu`：WireGuard MTU

### WireGuardPeer

`entity.WireGuardPeer`：

- `id`：主键
- `wireGuardTag`：所属模板
- `sort`：排序
- `address`：对端地址
- `port`：对端端口
- `publicKey`：对端公钥
- `preSharedKey`：预共享密钥
- `allowedIps`：逗号分隔 CIDR 列表
- `persistentKeepaliveInterval`：保活间隔
- `enabled`：是否启用

### Device 中的 WireGuard 字段

WireGuard 模板本身不包含客户端私钥和客户端地址，这部分放在设备对象里：

- `device.wireGuardTag`
- `device.wireGuardClientAddr`
- `device.wireGuardClientKey`

## 管理接口

模板接口 `/api/wire-guards`：

- `POST /api/wire-guards`
- `GET /api/wire-guards`
- `GET /api/wire-guards/:tag`
- `PUT /api/wire-guards/:tag`
- `DELETE /api/wire-guards/:tag`

Peer 接口：

- `POST /api/wire-guards/:tag/peers`
- `GET /api/wire-guards/:tag/peers`
- `PUT /api/wire-guards/:tag/peers/:id`
- `DELETE /api/wire-guards/:tag/peers/:id`

服务层会校验：

- 创建 peer 前模板必须存在
- 更新 peer 时路径 `tag` 与 body `wireGuardTag` 不能冲突
- 更新 peer 时路径 `id` 与 body `id` 不能冲突

## 生成逻辑

设备生成时，`resolveGenerateEndpoints()` 的流程是：

1. 若设备没有 `wireGuardTag`，直接不生成 endpoint
2. 读取对应模板
3. 读取该模板下所有 peers
4. 调用 `GetWireGuardEndpoints()` 组合成 sing-box endpoint

`GetWireGuardEndpoints()` 还会做以下判断：

- 模板不存在或设备为空：不生成
- 模板未启用：不生成
- 设备没有客户端地址：不生成
- 设备没有客户端私钥：不生成

满足条件时，最终输出结构为：

- `type: "wireguard"`
- `tag: wg.endpointTag`
- `mtu: wg.mtu`
- `address: [device.wireGuardClientAddr]`
- `private_key: device.wireGuardClientKey`
- `peers: [...]`

## Peer 转换规则

peer 列表会按 `sort` 升序、再按 `id` 稳定排序。每个启用的 peer 会被转换为：

- `address`
- `port`
- `public_key`
- `pre_shared_key`
- `allowed_ips`
- `persistent_keepalive_interval`

其中 `allowedIps` 会按英文逗号拆分并去掉空白。

## 客户端地址处理

`wireGuardClientAddr` 会经过 `normalizeWireGuardClientAddr()`：

- 已带 CIDR：原样保留
- 未带 CIDR：自动补 `/32`

这让后台既可以填写 `10.0.0.2`，也可以填写 `10.0.0.2/32`。

## 默认回退

虽然生成链路保留了默认 WireGuard 回退入口：

- `GetDefaultWireGuard()`
- `GetDefaultWireGuardPeers()`

但当前实现这两个函数都返回 `nil`。也就是说，空存储模式下 WireGuard 默认模板实际上未接通。

## 当前限制

- 设备侧私钥明文保存
- 一个设备当前只支持绑定一个 `wireGuardTag`
- 不支持设备级 peer 覆盖或 peer 可见性控制
- 默认 WireGuard 兼容模板目前未提供实际内容

## 适合更新本文档的场景

- 新增默认 WireGuard 模板
- 改为支持一个设备多个 WireGuard endpoint
- 引入更严格的密钥/地址校验
- 增加设备级覆盖参数
