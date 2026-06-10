# 节点分组

## 模块职责

节点分组负责把“订阅解析得到的原始节点标签”组织成策略分组。它本身不保存节点，只保存筛选规则和分组类型；生成时可分别渲染为 sing-box 出站组或 Surge / Shadowrocket Proxy Group。

代码入口：

- 实体定义：`entity/node_group.go`
- 管理接口：`service/service.go`
- 共享筛选：`convert/common/group.go`
- sing-box 生成转换：`convert/singbox/outbound.go`
- Surge 生成转换：`convert/surge/surge.go`
- Shadowrocket 生成转换：`convert/shadowrocket/shadowrocket.go`

## 数据模型

`entity.NodeGroup` 字段：

- `name`：后台展示名称
- `tag`：配置内唯一标识，供路由规则和其他模块引用
- `groupType`：当前主要使用 `selector` 或 `urltest`，作为未命中设备覆盖时的默认类型
- `testURL`：仅 `urltest` 使用，留空时回退到默认探测地址
- `include`：逗号分隔包含关键字
- `exclude`：逗号分隔排除关键字
- `deviceTypeOverrides`：设备级分组类型覆盖规则，详见下文“设备级类型覆盖”

文档中的“分组”并不是静态成员列表，而是“按关键字动态筛选”的规则定义。

## 设备级类型覆盖

`deviceTypeOverrides` 允许“同一份分组定义，针对不同设备输出不同的分组类型”，以兼顾网关类设备的“省心自动测速”与终端设备的“灵活手动切换”两类场景。

规则字符串格式：

```text
deviceCode:groupType,deviceCode:groupType
```

例如 `phone:selector,gateway:urltest` 表示设备 `gateway` 渲染为 `urltest`，设备 `phone` 渲染为 `selector`，其余设备使用默认 `groupType`。

解析规则（实现见 `convert/common.ParseDeviceTypeOverrides` 与 `ResolveGroupType`）：

- 以英文逗号 `,` 分隔多条覆盖规则；
- 每条规则以第一个英文冒号 `:` 分隔设备编码与目标类型，两端空白会被去除；
- 设备编码为空、目标类型为空、缺少冒号的规则直接忽略；
- 目标类型只接受 `selector` / `select` / `urltest`，非法值忽略并回退默认类型；
- `select` 兼容值会被规范化为 `selector`；
- 同一设备编码多次出现时，后出现的规则覆盖先出现的规则；
- 设备编码大小写敏感，不校验设备是否真实存在，避免删除设备后历史配置无法加载。

生成时三种输出格式（sing-box / Surge / Shadowrocket）都通过 `common.ResolveGroupType(group, deviceCode)` 解析最终类型，保证类型关键字与 `url` / `interval` / `tolerance` 等参数来自同一次判定。未配置覆盖字段的旧数据行为与改动前完全一致。

## 管理接口

接口路径为 `/api/node-groups`：

- `POST /api/node-groups`
- `PUT /api/node-groups/:tag`
- `DELETE /api/node-groups/:tag`
- `GET /api/node-groups`

创建时会先检查 `tag` 是否已存在；更新和删除直接按 `tag` 操作。

## 生成逻辑

生成时，`resolveGenerateOutbounds()` 会先得到当前设备可见的统一 Outbound。sing-box 输出由 `convert/singbox.GetOutbounds()` 组装 selector / urltest 出站；Surge 输出由 `convert/surge.Render()` 组装 select / url-test Proxy Group；Shadowrocket 输出由 `convert/shadowrocket.Render()` 组装同结构 Proxy Group。

处理顺序：

1. 收集当前已有出站的全部 `tag`
2. 按每条分组规则执行 `include` / `exclude` 过滤
3. 若过滤后没有成员，则整个分组不输出
4. 根据 `groupType` 生成 selector 或 urltest 出站

## 关键字过滤规则

过滤函数为 `common.FilterOutboundGroupTags()`，行为如下：

- `include` 为空：默认包含所有标签
- `include` 非空：标签命中任一包含关键字即可入组
- `exclude` 非空：命中任一排除关键字即移除
- 关键字使用英文逗号分隔
- 匹配规则是 `strings.Contains`，即子串匹配，不是正则，也不是精确匹配

为了保持生成结果稳定，过滤后的成员会按原始标签顺序输出，而不是按 map 遍历顺序输出。

## 支持的分组类型

### `selector`

生成字段：

- `type: "selector"`
- `outbounds: [...]`
- `default`: 自动设置为第一个成员

适合手动切换节点的场景。

### `urltest`

生成字段：

- `type: "urltest"`
- `outbounds: [...]`
- `url`: 使用 `testURL`，为空时回退到 `https://www.gstatic.com/generate_204`
- `interval: "10m"`
- `tolerance: 50`

适合自动测速和自动选择出口的场景。

Surge / Shadowrocket 输出中会映射为：

- `selector` / `select` -> `select`
- `urltest` -> `url-test`

Surge / Shadowrocket Proxy Group 只引用 `[Proxy]` 中已成功导出的节点，避免不支持协议被跳过后产生悬空成员。

## 与其他模块的关系

- 上游依赖“订阅管理”和“额外出站管理”提供候选标签
- 下游被“规则集管理”和“DNS 配置”引用，常见用法是把 `tag` 填到 `outbound` 或 `detour`

分组本身不会校验引用目标是否存在，因此后台可以保存一个暂时没有成员、或未来才会被引用的分组。

## 实现边界

当前实现的几个边界需要特别说明：

- 不支持正则、权重、优先级表达式等高级筛选能力
- 不支持分组嵌套；筛选输入是“已有出站 tag 列表”
- `groupType` 与设备覆盖类型只接受 `selector` / `select` / `urltest`，非法值会被忽略并回退到默认类型（`select` 规范化为 `selector`）
- 不做成员去重之外的质量校验；重复标签只按首次遇到顺序保留

## 适合更新本文档的场景

- 新增分组类型
- 更改关键字匹配算法
- 引入静态成员、嵌套分组或权重策略
- 调整默认 `urltest` 参数
