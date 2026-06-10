# 需求：节点分组按设备维度切换 selector / urltest 策略类型

- 状态：已实现（implemented）
- 提出时间：2026-06-10
- 实现时间：2026-06-10
- 关联模块：[节点分组](../../modules/node-group.md)、[设备管理](../../modules/device.md)、[配置生成流程](../../architecture/config-generation.md)

## 背景

当前 `entity.NodeGroup` 只有一个全局的 `groupType` 字段（`selector` / `urltest`，详见 `entity/node_group.go`、`entity/enums.go`）。生成时由 `convert/singbox/outbound.go` 的 `constructOutboundGroup()` 统一按这个类型渲染为 sing-box 出站，Surge / Shadowrocket 同理。

这意味着**同一个节点分组在所有设备上的策略类型是固定的**：要么全是手动选择（selector），要么全是自动测速（urltest），无法按设备区分。

## 需求描述

希望让 `selector` 和 `urltest` 这两种分组类型**可以按设备决定**：

- 对于网关类、长期无人值守的设备（如旁路由、软路由、电视盒子），希望对应分组渲染为 `urltest`，自动测速选优、无需人工干预。
- 对于电脑、手机等需要灵活手动切换出口的设备，希望同一分组渲染为 `selector`，可以随时在客户端手动选节点。

核心诉求：**同一份节点分组定义，能针对不同设备输出不同的分组类型**，以兼顾“省心自动”与“灵活手动”两类使用场景。

## 现状分析

1. 分组类型来源：`NodeGroup.GroupType`，全局唯一。
2. 生成链路：
   - sing-box：`service/outbound_cache.go:resolveGenerateOutbounds()` → `singbox.GetOutbounds()` → `constructOutboundGroup()`。
   - Surge：`service/generated.go` → `surge.Render()`。
   - Shadowrocket：`service/generated.go` → `shadowrocket.Render()`。
3. 设备上下文：生成链路已经携带 `deviceCode`，并已有成熟的“按设备可见性”模式 `isDeviceVisible(deviceCode, visibleDevices)`（`service/outbound_cache.go:68`），Outbound 与订阅都借助它做设备级过滤。
4. 当前 `constructOutboundGroup()` 拿不到 `deviceCode`，因此无法做设备级的类型判定。

## 方案设计（建议）

复用现有“逗号分隔设备编码 + 默认回退”的约定，给分组增加一份**设备级类型覆盖**，在不破坏现有数据的前提下扩展：

### 数据模型

在 `entity.NodeGroup` 增加可选字段，保留 `GroupType` 作为默认值不变：

- `GroupType`（保留）：默认分组类型，未命中任何设备覆盖规则时使用。
- 新增 `DeviceTypeOverrides`（建议名）：设备到类型的覆盖映射。

存储形式有两种可选，建议优先**方案 A（字符串规则，低迁移成本）**：

- **方案 A：逗号分隔规则字符串**，如 `phone:selector,gateway:urltest`。
  - 优点：沿用现有 `string` 字段风格，JSON 文件兼容性好，数据库只需新增普通字符串列，不需要 JSON / 关联表等复杂结构。
  - 注意：数据库和 Supabase 仍需要补充字段映射与 schema 变更说明；导入导出需要随实体字段一起回归验证。
  - 解析后得到 `map[deviceCode]groupType`。
- **方案 B：结构化字段**（`[]{DeviceCode, GroupType}` 或 `map`）。
  - 表达更清晰，但涉及多存储后端（PostgreSQL / MySQL / Supabase / JSON）的序列化与迁移，成本更高。

### 规则字符串语义（方案 A）

`DeviceTypeOverrides` 字符串格式建议固定为：

```text
deviceCode:groupType,deviceCode:groupType
```

解析规则：

- 以英文逗号 `,` 分隔多条覆盖规则。
- 每条规则以第一个英文冒号 `:` 分隔设备编码与目标类型。
- 设备编码与目标类型两端空白需要 `strings.TrimSpace()`。
- 设备编码为空、目标类型为空、缺少冒号的规则直接忽略。
- 目标类型只接受 `selector` / `select` / `urltest`；非法值忽略并回退默认类型。
- 同一个设备编码出现多次时，后出现的规则覆盖先出现的规则，便于用户在字符串末尾修正。
- 设备编码按现有系统约定大小写敏感；不额外校验设备是否存在，避免删除设备后历史配置无法加载。

### 生成逻辑

1. 在生成链路中把 `deviceCode` 透传给分组拼装函数：
   - `singbox.GetOutbounds(outbounds, groupRules)` → 增加 `deviceCode` 入参；
   - `constructOutboundGroup(groupRules, tags)` → 增加 `deviceCode` 入参。
   - Surge `surge.Render()` 与 Shadowrocket `shadowrocket.Render()` 已持有 `device.Code`，对应分组渲染处同步取用覆盖结果。
2. 新增解析与判定函数，统一所有输出格式：
   ```
   resolveGroupType(group *entity.NodeGroup, deviceCode string) entity.NodeGroupType
   ```
   - 命中设备覆盖 → 返回覆盖类型；
   - 未命中 → 返回 `group.GroupType`（默认）。
3. `constructOutboundGroup()` 内 `switch` 改为基于 `resolveGroupType()` 的返回值决定渲染 `urltest`（带 `url` / `interval` / `tolerance`）还是 `selector`（带 `default`）。
   - sing-box 输出类型必须规范化：`select` 覆盖值仅作为兼容输入，最终输出为 `selector`。
   - 默认 `testURL` 使用局部变量处理，不再在转换函数中回写修改 `groupRule.TestURL`。
4. Surge / Shadowrocket 中 `selector|select → select`、`urltest → url-test` 的映射保持不变，仅把“类型来源”换成 `resolveGroupType()`。
   - 是否追加 `url` / `interval` / `tolerance` 也必须基于同一次解析结果，不能继续直接判断 `group.GroupType`，避免覆盖为 `urltest` 时类型变了但参数缺失。

### 校验与兼容

- 覆盖类型只接受 `selector` / `select` / `urltest`，非法值忽略并回退默认类型（与现状“不强校验、回退默认”的实现风格一致）。
- 旧数据无覆盖字段时，行为与现在完全一致（默认 `GroupType`），保证向后兼容。
- `testURL` 仍为分组全局属性；当某设备被覆盖为 `urltest` 而分组 `testURL` 为空时，沿用默认探测地址 `https://www.gstatic.com/generate_204`。
- 旧数据中如果 `groupType` 使用兼容值 `select`，sing-box 输出仍应按 `selector` 渲染，Surge / Shadowrocket 输出仍按 `select` 渲染。

## 前端改动（建议）

- 节点分组编辑页：在“分组类型”之外增加“设备级类型覆盖”编辑区，每行选择设备 + 目标类型；保存时序列化为方案 A 的规则字符串。
- 设备下拉数据复用现有设备列表接口。
- 文案需说明：未配置覆盖的设备使用默认分组类型。

## 影响范围

| 层 | 文件 / 模块 | 改动 | 状态 |
|----|------------|------|------|
| 实体 | `entity/node_group.go` | 新增覆盖字段 `DeviceTypeOverrides` | ✅ 已完成 |
| 存储 | `storage/models.go` | `DBNodeGroup` 新增 `device_type_overrides` 字符串列，实体映射同步 | ✅ 已完成 |
| 存储 | `storage/supabase.go` | `supabaseNodeGroup` 新增 `device_type_overrides` 字段映射 | ✅ 已完成 |
| 转换 | `convert/common/group.go` | 新增 `ResolveGroupType()` / `ParseDeviceTypeOverrides()` 统一判定 | ✅ 已完成 |
| 转换 | `convert/singbox/outbound.go` | `GetOutbounds` / `constructOutboundGroup` 增加 `deviceCode`，按设备解析类型 | ✅ 已完成 |
| 转换 | `convert/surge/surge.go`、`convert/shadowrocket/shadowrocket.go` | 分组类型来源改为按设备解析 | ✅ 已完成 |
| 服务 | `service/outbound_cache.go` | 透传 `deviceCode` 到分组拼装 | ✅ 已完成 |
| API / 类型 | `docs/reference/api-reference.md`、`web/src/types/index.ts` | 补充 `deviceTypeOverrides` 字段 | ✅ 已完成 |
| 前端 | `web/src/components/NodeGroupModal.tsx` | 设备级类型覆盖编辑 UI，设备列表复用现有接口 | ✅ 已完成 |
| 文档 | `docs/modules/node-group.md` | 补充“设备级类型覆盖”说明 | ✅ 已完成 |

> Supabase schema 变更说明：需要在 `node_groups` 表新增一列 `device_type_overrides text`（可空，默认空字符串），旧行缺列时按空字符串读取即可保持兼容。

## 待确认问题（已定稿）

1. 存储形式：采用方案 A（逗号分隔规则字符串）。
2. 覆盖字段命名：`DeviceTypeOverrides`（JSON `deviceTypeOverrides`，列名 `device_type_overrides`）。
3. 是否支持设备分组/标签级覆盖：本期仅做设备编码级别，后续视情况扩展。

## 验收标准

1. 同一节点分组，在设备 A（覆盖为 `urltest`）生成的配置中：sing-box 渲染为 `type: "urltest"`，Surge / Shadowrocket 渲染为 `url-test`；在设备 B（默认或覆盖为 `selector`）中：sing-box 渲染为 `type: "selector"`，Surge / Shadowrocket 渲染为 `select`。
2. 未配置任何覆盖时，所有设备行为与改动前一致（回归无差异）。
3. 非法覆盖类型被忽略并回退到默认类型，不导致生成失败。
4. 覆盖类型为 `select` 时，sing-box 输出规范化为 `selector`，Surge / Shadowrocket 输出为 `select`。
5. 覆盖为 `urltest` 且 `testURL` 为空时，三种输出都使用默认探测地址；转换过程不修改原始 `NodeGroup.TestURL`。
6. 数据库 / Supabase / JSON 文件存储下，新字段创建、更新、列表、导入导出都能 round-trip，旧数据缺少字段时可正常读取。
7. 三种输出格式（sing-box / Surge / Shadowrocket）行为一致，并补充对应单元测试。

## 实现说明

- 核心判定收敛到 `convert/common.ResolveGroupType()` / `ParseDeviceTypeOverrides()` 两个纯函数，便于单测覆盖（`convert/common/group_test.go`）。
- sing-box `constructOutboundGroup()` 增加 `deviceCode` 入参，使用局部变量处理默认探测地址，不再回写 `groupRule.TestURL`（`convert/singbox/outbound_test.go` 已断言不回写）。
- Surge / Shadowrocket 的 Proxy Group 渲染统一改为基于同一次 `ResolveGroupType()` 结果决定类型关键字与探测参数（各自新增 `TestRenderProxyGroupDeviceTypeOverride`）。
- 前端 `NodeGroupModal` 增加“设备级类型覆盖”编辑区，复用 `GET /api/devices` 设备列表，保存时序列化为规则字符串。

## 实现后

已于 2026-06-10 实现并通过 `go test ./...` 与前端 `tsc --noEmit` 验证，本文件已移动到 `requirements/implemented/`，并同步更新 `docs/modules/node-group.md` 与 `docs/reference/api-reference.md`。
