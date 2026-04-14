# 入站管理

## 模块职责

入站管理维护可复用的 sing-box `inbound` 模板，并通过设备绑定关系决定某台设备最终启用哪些入站。

代码入口：

- 实体定义：`entity/device_management.go`
- 管理接口：`service/device_management.go`
- 生成转换：`convert/singbox/inbound.go`

## 数据模型

`entity.Inbound` 字段：

- `tag`：模板唯一标识
- `name`：展示名称
- `description`：说明
- `type`：入站类型，如 `tun`、`http`、`socks`、`mixed`
- `enabled`：是否启用
- `sort`：模板排序
- `configJson`：完整的 sing-box inbound JSON

这个模型的关键点是：系统并不把各类入站字段拆成结构化列，而是保留一份原始 JSON 模板。

## 管理接口

接口路径 `/api/inbounds`：

- `POST /api/inbounds`
- `GET /api/inbounds`
- `GET /api/inbounds/:tag`
- `PUT /api/inbounds/:tag`
- `DELETE /api/inbounds/:tag`

设备绑定接口位于设备模块：

- `PUT /api/devices/:code/inbounds`
- `GET /api/devices/:code/inbounds`

创建时会检查 `tag` 唯一性；更新时要求路径 `tag` 与 body 中 `tag` 一致。

## 生成逻辑

配置生成时，`resolveGenerateInbounds()` 先得到“当前设备应该使用哪些模板”，再由 `convert/singbox.GetInbounds()` 把模板 JSON 转成 `[]entity.SingInbound`。

转换规则：

- 先按 `sort` 升序排序
- 跳过 `enabled=false` 的模板
- 对 `configJson` 做 JSON 反序列化
- 反序列化失败则记录 warning 并跳过该模板

因此单个模板损坏不会让整个设备配置生成失败。

## 默认回退

在空存储兼容模式下，系统内置四个默认入站：

- `tun-default`
- `http-default`
- `socks-default`
- `mixed-default`

并默认把这四个模板全部绑定到当前设备，按 `0 / 10 / 20 / 30` 排序输出。

注意这里的“默认回退”只在设备绑定为空或 Inbound 列表为空时生效，不代表这些模板一定存在于实际存储里。

## 与设备模块的关系

设备与入站是多对多关系，但当前接口设计为“按设备整体覆盖绑定表”：

- 一个设备可绑定多个入站
- 同一个入站模板也可被多个设备复用
- 排序由绑定关系上的 `sort` 决定，而不是只看模板本身的 `sort`

## 推荐理解方式

可以把该模块看作“两层模型”：

1. Inbound 模板：定义可复用的 sing-box JSON 片段
2. DeviceInbound 绑定：决定某台设备最终启用哪些模板、按什么顺序启用

## 当前限制

- `type` 只是管理端元数据，真正生效的是 `configJson` 中的内容
- 后端不对 `configJson` 做详细 schema 校验，只在生成时尝试反序列化
- 不支持模板参数化；不同设备需要差异配置时，通常要建多份模板
- 没有单独的“启用某设备上的某个入站”接口，只有整表覆盖

## 适合更新本文档的场景

- 引入结构化 Inbound 表单
- 增加入站类型校验
- 调整默认入站模板
- 修改设备绑定为增量接口或支持设备级参数覆盖
