# 设备管理

## 模块职责

设备管理定义“谁可以获取配置”以及“该设备应加载哪些入站和 WireGuard 参数”。它是开放生成接口的认证与个性化中心。

代码入口：

- 实体定义：`entity/device_management.go`
- 管理接口：`service/device_management.go`
- 生成入口：`service/generated.go`
- 默认回退：`convert/singbox/device_management_defaults.go`

## 核心对象

### Device

`entity.Device` 字段：

- `code`：设备唯一编码，也是 `/open/generate/:device` 的路径参数
- `name`：展示名称
- `description`：说明
- `token`：生成接口鉴权令牌
- `enabled`：禁用后拒绝生成
- `sort`：后台排序
- `wireGuardTag`：关联的 WireGuard 模板 tag
- `wireGuardClientAddr`：该设备在 WireGuard 网络中的地址
- `wireGuardClientKey`：该设备的 WireGuard 私钥

### DeviceInbound

`entity.DeviceInbound` 表示设备和 Inbound 模板的多对多关系：

- `deviceCode`
- `inboundTag`
- `sort`

## 管理接口

设备接口：

- `POST /api/devices`
- `GET /api/devices`
- `GET /api/devices/:code`
- `PUT /api/devices/:code`
- `DELETE /api/devices/:code`

设备与 Inbound 绑定接口：

- `PUT /api/devices/:code/inbounds`
- `GET /api/devices/:code/inbounds`

关键行为：

- 创建设备时会检查 `code` 唯一性
- 更新时要求路径 `:code` 与 body 中 `code` 一致
- 设置绑定关系时，服务层会校验设备存在、绑定列表不含 `null`、每个 `inboundTag` 确实存在
- `SetDeviceInbounds` 是全量覆盖，不是增量追加

## 开放生成接口中的作用

生成接口：

- `GET /open/generate/:device?token=...`：输出 sing-box JSON
- `GET /open/surge/:device?token=...`：输出 Surge 配置文本
- `GET /open/shadowrocket/:device?token=...`：输出 Shadowrocket 配置文本

处理顺序：

1. 根据 `device` 查找目标设备
2. 若设备不存在，返回 `404`
3. 若设备被禁用，返回 `403`
4. 若 `token` 不匹配，返回 `401`
5. 认证通过后继续组装对应格式需要的数据

Surge / Shadowrocket 输出与 sing-box 输出共享设备鉴权和 Outbound 解析链路，但第一版不导出 Inbound 与 WireGuard endpoint。

因此设备对象既承担身份识别，也承担配置个性化参数承载。

## 空存储回退逻辑

项目保留了一层旧行为兼容：

- 当存储里的设备列表完全为空时，`resolveGenerateDevice()` 才会回退到内置默认设备
- 当前默认设备只有一个：`code=default`，`token=996007`

这点很重要：只要存储里已有任意设备，默认设备就不会再参与匹配。

## 与 Inbound 的关系

设备本身不直接保存 Inbound JSON，只保存绑定关系。生成时 `resolveGenerateInbounds()` 会：

1. 读取该设备的绑定列表
2. 若绑定为空，则尝试使用默认绑定
3. 读取全部 Inbound 模板
4. 若模板为空，则尝试使用默认 Inbound
5. 按绑定 `sort` 排序后输出

这意味着“设备配置”由设备对象、绑定表和 Inbound 模板三部分共同决定。

## 与 WireGuard 的关系

如果设备设置了 `wireGuardTag`，生成时还会尝试：

1. 读取对应 WireGuard 模板
2. 读取该模板下的 peers
3. 结合设备自己的 `wireGuardClientAddr` 和 `wireGuardClientKey`
4. 生成 sing-box `endpoints`

只有模板、设备客户端地址、设备私钥同时满足条件时，最终才会输出 WireGuard endpoint。

## 当前实现边界

- 设备 token 为明文存储，没有哈希
- 没有 token 轮换、过期时间、最后访问时间等管理能力
- 开放生成接口只做 query token 校验，没有签名或更高级认证
- 默认设备不是 `phone`、`pad` 等多设备集合，当前只有 `default`
- 设备删除时，文档层面应默认理解为存储层也会清理关联数据；具体清理策略取决于存储实现

## 适合更新本文档的场景

- 调整生成接口鉴权方式
- 增加设备字段
- 更改默认回退设备
- 引入设备级路由、设备级 DNS 或更细粒度可见性控制
