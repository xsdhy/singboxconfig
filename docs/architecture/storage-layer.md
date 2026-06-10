# 存储抽象层

## 设计目标

存储层的目标不是做复杂领域建模，而是向上提供统一、稳定的 CRUD 能力，让服务层只依赖业务实体和接口，不依赖具体存储介质。

项目需要这层抽象的原因：

- 生产部署需要关系型数据库
- 云环境下希望直接接入 Supabase REST API

因此项目把所有持久化能力统一收敛到 `storage.Storage`。

## 快速选择指南

| 场景 | 推荐方案 | 环境变量设置 |
|------|---------|----------|
| 本地开发测试 | DatabaseStorage | `DATABASE_URL=postgresql://...` |
| Docker 单机部署 | DatabaseStorage | `DATABASE_URL=postgresql://...` |
| 生产自建（PostgreSQL） | DatabaseStorage | `DATABASE_URL=postgresql://...` |
| 生产自建（MySQL） | DatabaseStorage | `DATABASE_URL=mysql://...` |
| 云部署（Supabase） | SupabaseStorage | `SUPABASE_URL=...` + `SUPABASE_KEY=...` |
| 多实例高可用 | DatabaseStorage + PostgreSQL | `DATABASE_URL=postgresql://...` |

**选择原则**：
- **开发/测试**：DatabaseStorage + PostgreSQL/MySQL
- **单机生产**：DatabaseStorage + PostgreSQL/MySQL
- **云原生**：SupabaseStorage
- **多副本**：DatabaseStorage 或 SupabaseStorage

## 接口结构

`storage/storage.go` 没有定义一个巨大的平面接口，而是先拆成多个子接口，再组合为 `Storage`：

- `SubscribeStorage`
- `NodeGroupStorage`
- `RuleSetStorage`
- `GlobalSettingStorage`
- `DeviceStorage`
- `InboundStorage`
- `DeviceInboundStorage`
- `WireGuardStorage`
- `WireGuardPeerStorage`
- `ExtraOutboundStorage`
- `OutboundStorage`

最终：

```go
type Storage interface {
    SubscribeStorage
    NodeGroupStorage
    RuleSetStorage
    GlobalSettingStorage
    DeviceStorage
    InboundStorage
    DeviceInboundStorage
    WireGuardStorage
    WireGuardPeerStorage
    ExtraOutboundStorage
    OutboundStorage
}
```

这样做的收益是：

- 接口语义按资源类型自然分组
- 测试时可以只关注某一类存储行为
- 后续新增资源类型时扩展成本低

统一的“未找到”语义由 `storage.ErrNotFound` 表达，服务层据此决定返回 404、400 或忽略默认值回退。

## 统一数据模型

存储层对外统一使用 `entity/` 下的实体，而不是暴露数据库模型或 HTTP DTO。这样有几个直接效果：

- `service/` 与具体存储方式解耦
- `convert/singbox/`、`convert/surge/` 与 `convert/shadowrocket/` 可以直接消费服务层取出的实体
- 导入导出可以直接复用实体结构

例如：

- `storage.GetDevice` 返回 `*entity.Device`
- `storage.ListRuleSets` 返回 `[]*entity.RuleSet`
- `storage.ListWireGuardPeers` 返回 `[]*entity.WireGuardPeer`

## DatabaseStorage

### 实现方式

`storage/database.go` 基于 GORM，对每种资源都提供实体与数据库模型之间的双向映射。

初始化时调用 `AutoMigrate` 创建或更新表结构，涉及：

- `DBSubscribe`
- `DBNodeGroup`
- `DBRuleSet`
- `DBGlobalSetting`
- `DBDevice`
- `DBInbound`
- `DBDeviceInbound`
- `DBWireGuard`
- `DBWireGuardPeer`
- `DBExtraOutbound`
- `DBOutbound`

数据库模型定义集中在 `storage/models.go`。

### 映射原则

设计上采取“显式映射”而不是直接把 GORM tag 写在实体上：

- `DB*` 结构负责数据库字段定义
- `ToEntity()` 把数据库记录转为业务实体
- `FromEntity()` 把业务实体写回数据库模型

这种做法避免了：

- 业务实体被 ORM 注解污染
- 存储层变更直接影响服务层和转换层
- JSON 字段命名与数据库列命名互相牵制

### 行为特点

- `Create*` 直接插入
- `Update*` 基于主键条件更新，并检查 `RowsAffected`
- `Delete*` 删除后也检查 `RowsAffected`
- `SetGlobalSetting` 使用 `Save`，具有 upsert 语义
- 删除设备时会在事务中先清理 `device_inbounds`，再删设备主体

### 当前实现现状

虽然 `go.mod` 已引入 PostgreSQL / MySQL 驱动，但启动入口当前只在 `DATABASE_URL` 分支中接入 PostgreSQL。换句话说：

- `DatabaseStorage` 本身是通用 GORM 存储层
- `cmd/server/main.go` 目前只把 PostgreSQL 接到了实际运行路径

## SupabaseStorage

### 实现方式

`storage/supabase.go` 通过 Supabase PostgREST API 访问数据，而不是走 GORM。

核心特点：

- `baseURL = SUPABASE_URL + "/rest/v1"`
- 使用 `http.Client`
- 为每类资源定义 `supabase*` 数据模型
- JSON tag 使用 snake_case 以匹配 Postgres 列名

例如：

- `supabaseDevice` 对应 `entity.Device`
- `supabaseInbound` 对应 `entity.Inbound`
- `supabaseWireGuardPeer` 对应 `entity.WireGuardPeer`

### 为什么单独实现

Supabase 的访问模式与 GORM 完全不同：

- 查询通过 URL 参数表达
- 写入通过 HTTP body
- 认证依赖 API Key

如果强行复用 GORM 模型反而会增加复杂度，因此项目单独做了一套 HTTP 存储实现。

### 优点

- 不需要应用直接管理数据库连接
- 适合云托管部署
- 与 Supabase 平台能力兼容

### 局限

- 请求序列化和错误处理更繁琐
- 调试体验弱于本地数据库
- 延迟和网络失败会直接暴露到应用层
- 不会像 GORM `AutoMigrate` 那样自动建列：新增字段时需要手动在 Supabase 执行 DDL

### Schema 变更须知

`DatabaseStorage` 在启动时通过 `AutoMigrate` 自动新增列，而 `SupabaseStorage` 走 REST API 不做迁移，新增字段必须手动在 Supabase 加列。例如「节点分组设备级类型覆盖」需要：

```sql
ALTER TABLE node_groups
  ADD COLUMN IF NOT EXISTS device_type_overrides text NOT NULL DEFAULT '';
```

旧行会回填空字符串，保持向后兼容（空字符串 = 无覆盖，所有设备使用默认 `group_type`）。

## 存储后端切换机制

启动时由 `cmd/server/main.go` 根据环境变量决定使用哪种实现：

1. `SUPABASE_URL` 和 `SUPABASE_KEY` 同时存在 -> `NewSupabaseStorage`
2. `DATABASE_URL` 存在 -> `NewDatabaseStorage`
3. 否则 -> 启动失败，要求配置数据库

这种切换机制的特点是：

- 服务层完全不需要知道底层实现
- 运行环境可以无代码切换持久化方案
- 强制使用持久化存储，确保数据安全

## 关联数据处理策略

项目中有两类资源需要特别注意：

### `DeviceInbound`

它描述设备与入站模板的绑定关系，是典型的关联表数据。

实现策略：

- 服务层通过 `SetDeviceInbounds(deviceCode, bindings)` 做整组覆盖
- 数据库存储中按设备维度清空再重建
- 生成配置时再按 `Sort` 排序并关联到具体 `Inbound`

### `WireGuardPeer`

它依附于 `WireGuard` 模板。

实现策略：

- 使用独立资源存储
- 读取时按 `wireGuardTag` 聚合
- 输出 endpoint 时按 `Sort` 和 `ID` 稳定排序

## 存储层与服务层的职责边界

当前分工比较明确：

- 存储层负责 CRUD、数据映射、缺失语义、简单级联删除
- 服务层负责参数校验、路径一致性校验、排序、业务性跳过逻辑

例如：

- “设备不存在时返回什么 HTTP 状态”属于服务层
- “删除设备前先删除 device_inbounds”属于存储层
- “RuleSet 列表要按 sort 排序”属于服务层

这让存储层保持相对通用，不被 HTTP 语义污染。

## 与导入导出的关系

`service/config_transfer.go` 会遍历 `Storage` 中的所有资源，构造 `transfer.ConfigTransferData`。因此存储层实际上承担了“系统事实来源”的角色：

- 导出时，从所有资源表拉全量数据
- 导入时，按实体逐类写入
- 资源是否已存在，依赖 `Get*` + `ErrNotFound` 判断

如果某种存储实现与其它实现的语义不一致，导入导出和生成流程都会受影响，因此测试覆盖很关键。

## 设计优点

- 上层业务与存储介质解耦
- 单元测试可以复用同一套业务逻辑
- 便于按环境选择最合适的持久化方案
- 新资源类型可以按接口扩展，不需要重写整个存储层

## 当前需要注意的问题

- 启动入口暂未真正暴露 MySQL 的切换方式
- Supabase 存储的行为依赖远程 API，故障模型与本地数据库不同
- 某些”唯一性”校验当前由服务层先查再写，严格并发下仍应依赖底层存储约束兜底
