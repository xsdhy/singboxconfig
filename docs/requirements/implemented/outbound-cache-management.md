# 需求文档：Outbound 订阅缓存管理系统

**文档版本**: 1.1
**创建日期**: 2026-04-12
**状态**: 已实现
**优先级**: 高

---

## 1. 需求概述

### 1.1 背景

当前系统的订阅管理中，从订阅源拉取的 Outbound（出站节点）在生成 sing-box 配置时被直接输出，没有进行保存和缓存。这导致每次生成配置都需要重新解析订阅源，存在以下问题：

1. **性能问题**：每次生成配置都需要远程拉取，增加了生成时间和网络负载
2. **可维护性问题**：没有查看、编辑、管理从订阅中获取的 Outbound 的能力
3. **稳定性问题**：订阅源不可用时，无法降级到缓存数据
4. **用户体验问题**：无法对订阅获取的 Outbound 进行筛选、搜索等操作

### 1.2 业务价值

- 提升配置生成性能（减少重复的订阅拉取）
- 提供统一的 Outbound 管理界面（手动和订阅来源统一管理）
- 支持对 Outbound 的灵活筛选、搜索、编辑操作
- 通过缓存机制提高系统稳定性和容错能力

### 1.3 需求范围

本需求包含以下三个主要方面：

1. **数据模型重构**：ExtraOutbound → Outbound（添加来源字段）
2. **功能升级**：额外出站管理 → Outbound 统一管理（支持筛选、搜索、批量操作）
3. **缓存机制**：实现智能缓存策略（支持手动更新、自动过期更新）

---

## 2. 功能需求

### 2.1 Outbound 统一管理

#### 2.1.1 数据来源整合

将所有 Outbound 统一到单一表中，通过来源字段区分：

| 来源 | 说明 | 操作权限 |
|------|------|--------|
| `MANUAL` | 手动添加的出站节点 | 完全编辑、删除 |
| `SUBSCRIPTION` | 来自订阅源的出站节点 | 查看、启用/禁用、删除；不可编辑 JSON。注意：订阅刷新会全量覆盖这类记录，用户对订阅来源节点的手动删除或启用状态修改不保证持久保留 |

#### 2.1.2 管理界面功能

**列表页面**：
- 按来源筛选（手动 / 订阅）
- 按订阅源筛选（仅订阅来源显示）
- 按启用状态筛选
- 按标签、名称搜索
- 分页显示（每页 20 条）
- 批量启用/禁用
- 单个删除

**详情/编辑页面**：
- 查看完整信息（包括 configJson）
- 手动来源：支持编辑所有字段
- 订阅来源：仅显示，不可编辑（可手动删除）
- 显示来源、订阅源、最后更新时间等元数据

**订阅来源更新**：
- "立即获取" 按钮：手动触发从订阅源拉取最新 Outbound
- 显示拉取状态和结果（新增/更新/删除数量）

### 2.2 订阅缓存机制

#### 2.2.1 缓存字段添加到 Subscribe 表

```sql
ALTER TABLE subscribe ADD COLUMN (
  outbound_last_fetch_time TIMESTAMP NULL COMMENT 'Outbound 最后一次拉取时间',
  outbound_cache_duration INT DEFAULT 0 COMMENT 'Outbound 缓存时间（分钟），0 表示不缓存'
);
```

**字段说明**：
- `outbound_last_fetch_time`：上次从此订阅源成功拉取 Outbound 的时间
- `outbound_cache_duration`：缓存有效期（单位：分钟）
  - `0`：每次都重新拉取（无缓存）
  - `> 0`：缓存 N 分钟，超过后重新拉取
  - 默认值：`0`（用户可配置）

#### 2.2.2 缓存过期判断逻辑

```
if outbound_cache_duration == 0:
    # 无缓存模式，每次都拉取
    FETCH_FROM_SUBSCRIPTION()
elif now() - outbound_last_fetch_time > outbound_cache_duration * 60:
    # 缓存过期，需要重新拉取
    FETCH_FROM_SUBSCRIPTION()
    UPDATE outbound_last_fetch_time = now()
else:
    # 使用缓存中的 Outbound
    USE_CACHED_OUTBOUND()
```

### 2.3 生成流程改进

#### 2.3.1 配置生成时的缓存策略

当生成设备配置时，对每个订阅源：

1. **检查缓存状态**
   - 获取订阅源的 `outbound_cache_duration` 配置
   - 检查 `outbound_last_fetch_time` 和当前时间

2. **缓存决策**
   - 若在有效期内，使用 Outbound 表中的缓存数据
   - 若已过期或无缓存，调用订阅源拉取新数据

3. **缓存更新**
   - 若拉取成功，对该订阅源的 Outbound 执行
     - 删除旧数据
     - 新增当前数据
   - 更新 `outbound_last_fetch_time = now()`
   - 不要求事务级原子性，采用最大努力更新策略

4. **容错处理**
   - 若拉取失败但有缓存，使用旧缓存数据（降级策略）
   - 若无缓存也无法拉取，返回空列表并记录告警

#### 2.3.2 后台更新接口

**订阅管理页面**：支持手动触发更新

```
POST /api/subscribes/{name}/refresh-outbound

Response:
{
  "status": "success",
  "data": {
    "added": 5,
    "updated": 2,
    "deleted": 1,
    "last_fetch_time": "2026-04-12T10:30:45Z"
  }
}
```

---

## 3. 数据模型设计

### 3.1 表结构变更

#### 3.1.1 当前 ExtraOutbound 表结构

```sql
CREATE TABLE extra_outbound (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tag VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255),
  description TEXT,
  type VARCHAR(50),
  enabled BOOLEAN DEFAULT true,
  sort INT DEFAULT 0,
  visible_devices TEXT,
  config_json LONGTEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

#### 3.1.2 新 Outbound 表结构

```sql
CREATE TABLE outbound (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tag VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255),
  description TEXT,
  type VARCHAR(50),           -- 出站协议类型
  enabled BOOLEAN DEFAULT true,
  sort INT DEFAULT 0,
  visible_devices TEXT,       -- 设备可见性（逗号分隔）
                              -- 对于 MANUAL：来自 outbound 本身配置
                              -- 对于 SUBSCRIPTION：冗余自订阅源的 visible_devices
  config_json LONGTEXT NOT NULL,

  -- 新增字段：来源标识
  source VARCHAR(20) NOT NULL DEFAULT 'MANUAL', -- MANUAL | SUBSCRIPTION

  -- 若来自订阅，记录订阅源信息
  subscribe_name VARCHAR(255), -- 来源订阅 name，MANUAL 时为 NULL

  -- 缓存元数据
  last_fetch_time TIMESTAMP NULL COMMENT '来自订阅时的最后拉取时间',

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  -- 索引
  INDEX idx_source (source),
  INDEX idx_subscribe_name (subscribe_name),
  INDEX idx_enabled (enabled),
  INDEX idx_tag (tag)
);
```

**tag 唯一性说明**：
- `tag` 保持全局唯一，这是 sing-box 配置的约束，不做放宽
- 当手动 Outbound 与订阅 Outbound，或不同订阅之间出现相同 `tag` 时，按最新一次写入结果覆盖旧记录
- 覆盖属于预期行为，不额外保留同 tag 的多份记录
- 刷新结果中的新增、更新、删除统计应以最终写入后的覆盖结果为准

#### 3.1.3 Subscribe 表添加缓存配置和设备可见性

```sql
ALTER TABLE subscribe ADD COLUMN (
  -- 设备可见性
  visible_devices TEXT COMMENT '可使用此订阅的设备列表（逗号分隔），为空表示所有设备可用',

  -- 缓存配置
  outbound_last_fetch_time TIMESTAMP NULL COMMENT 'Outbound 最后一次拉取时间',
  outbound_cache_duration INT DEFAULT 0 COMMENT 'Outbound 缓存时间（分钟），0 表示不缓存',

  -- 拉取状态记录
  outbound_last_fetch_status VARCHAR(20) COMMENT 'PENDING|SUCCESS|FAILED',
  outbound_last_fetch_error TEXT COMMENT '最后一次拉取的错误信息'
);
```

**字段说明**：
- `visible_devices`：此订阅源可被哪些设备使用（逗号分隔的设备码）
  - 为空：所有设备都可使用
  - 非空：仅列出的设备可使用
- `outbound_last_fetch_time`：上次从此订阅源成功拉取 Outbound 的时间
- `outbound_cache_duration`：缓存有效期（单位：分钟）
  - `0`：每次都重新拉取（无缓存）
  - `> 0`：缓存 N 分钟，超过后重新拉取
  - 默认值：`0`（用户可配置）

### 3.2 Go 实体定义

#### 3.2.1 新 Outbound 实体

```go
type Outbound struct {
    ID              int64      `gorm:"primaryKey"`
    Tag             string     `gorm:"uniqueIndex;not null"`
    Name            string
    Description     string
    Type            string
    Enabled         bool       `gorm:"default:true"`
    Sort            int        `gorm:"default:0"`
    VisibleDevices  string     // 逗号分隔的设备码
    ConfigJson      string     `gorm:"type:longtext"`

    // 来源信息
    Source          string     `gorm:"default:MANUAL;index"`    // MANUAL | SUBSCRIPTION
    SubscribeName   string     `gorm:"index"` // 订阅源 name
    LastFetchTime   *time.Time // 来自订阅时的拉取时间

    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type OutboundSource string

const (
    OutboundSourceManual       OutboundSource = "MANUAL"
    OutboundSourceSubscription OutboundSource = "SUBSCRIPTION"
)
```

#### 3.2.2 Subscribe 表新增字段

```go
type Subscribe struct {
    // ... 现有字段 ...

    // 设备可见性
    VisibleDevices          string     // 可使用此订阅的设备列表（逗号分隔），为空表示所有设备可用

    // Outbound 缓存配置
    OutboundLastFetchTime   *time.Time // 最后一次拉取时间
    OutboundCacheDuration   int        // 缓存时间（分钟）
    OutboundLastFetchStatus string     // PENDING|SUCCESS|FAILED
    OutboundLastFetchError  string     // 拉取错误信息
}
```

---

## 4. 技术方案设计

### 4.1 核心设计原则

#### 4.1.1 缓存策略

- **延迟计算**：配置生成时才判断是否需要更新，避免后台定时任务复杂性
- **主动更新**：支持订阅管理页面手动触发更新
- **自动降级**：拉取失败时使用旧缓存数据（灰度策略）
- **非强制刷新**：缓存过期后才重新拉取，不是实时性要求
- **生成路径内刷新**：当前方案接受在配置生成路径中触发远程拉取与缓存更新，不单独引入后台刷新链路

#### 4.1.2 数据一致性

- **最大努力更新**：更新缓存时不强依赖事务能力，允许在多存储后端下采用分步写入
- **全量替换**：更新缓存时删除旧数据，插入新数据（硬删除）
- **版本控制**：记录每次拉取时间，支持追溯
- **覆盖优先**：若写入的 `tag` 已存在，以最新数据覆盖旧记录，这是显式业务规则
- **刷新覆盖用户操作**：订阅来源 Outbound 的删除、启用/禁用等用户操作，在下一次订阅刷新后允许被覆盖

#### 4.1.3 性能考虑

- **索引优化**：在 `source`、`subscribe_name`、`enabled` 上建立索引
- **批量操作**：删除旧数据时使用批量 DELETE，避免逐条操作
- **最大努力落库**：删除旧数据、新增缓存、更新订阅状态按顺序执行，不要求同一事务原子提交

### 4.2 实现流程

#### 4.2.1 配置生成时的缓存逻辑

```
生成设备配置(deviceCode):
  outbounds = []

  FOR each subscription in subscriptions:
    IF 需要更新缓存(subscription):
      outbound_list = 拉取订阅(subscription)  // 远程拉取
      IF 拉取成功:
        // 最大努力更新，不要求事务
        DELETE FROM outbound
          WHERE subscribe_name = subscription.name
          AND source = 'SUBSCRIPTION'

        FOR each outbound in outbound_list:
          UPSERT INTO outbound (
            source=SUBSCRIPTION,
            subscribe_name=subscription.name,
            last_fetch_time=now()
          )

        UPDATE subscribe
          SET outbound_last_fetch_time=now(),
              outbound_last_fetch_status='SUCCESS'
          WHERE name=subscription.name
    ELSE:
      // 使用缓存
      outbound_list = SELECT * FROM outbound
        WHERE subscribe_name = subscription.name
        AND source = 'SUBSCRIPTION'

  RETURN outbounds

需要更新缓存(subscription):
  IF subscription.outbound_cache_duration == 0:
    RETURN true  // 无缓存，每次都拉取

  IF subscription.outbound_last_fetch_time == NULL:
    RETURN true  // 从未拉取过

  elapsed = now() - subscription.outbound_last_fetch_time
  RETURN elapsed > subscription.outbound_cache_duration * 60 seconds
```

#### 4.2.2 手动更新接口

```
RefreshSubscriptionOutbound(subscribeName):
  subscription = 获取订阅(subscribeName)

  AWAIT: outbound_list = 拉取订阅(subscription)  // 异步，可能耗时

  IF 拉取成功:
    // 最大努力执行，失败时按步骤记录状态
    DELETE FROM outbound
      WHERE subscribe_name = subscribeName
      AND source = 'SUBSCRIPTION'

    FOR each outbound in outbound_list:
      UPSERT ...

    UPDATE subscribe
      SET outbound_last_fetch_time = now(),
          outbound_cache_duration = user_configured_value,
          outbound_last_fetch_status = 'SUCCESS',
          outbound_last_fetch_error = NULL
      WHERE name = subscribeName
    RETURN {added: count, updated: 0, deleted: 0}
  ELSE:
    UPDATE subscribe
      SET outbound_last_fetch_status = 'FAILED',
          outbound_last_fetch_error = error_message
    RETURN ERROR
```

### 4.3 数据迁移策略

#### 4.3.1 关键声明

⚠️ **重要**: 由于项目未上线，无历史数据需要保留，实现代码**不需要考虑向后兼容**。

- ❌ 不需要迁移脚本处理存量数据
- ❌ 不需要保留历史字段或兼容旧 API
- ❌ 不需要灰度发布或过渡方案
- ❌ 不需要在代码中留下与旧数据模型相关的逻辑
- ✅ 直接删除旧表，不保留迁移脚本
- ✅ 新代码中删除所有 ExtraOutbound 相关代码
- ✅ API 直接从 `/api/extra-outbounds` 切换到 `/api/outbounds`，不需要双支持

#### 4.3.2 实现步骤

1. **删除旧 Outbound 表**
   ```sql
   DROP TABLE IF EXISTS extra_outbound;
   ```

2. **创建新 Outbound 表**（直接创建最终设计的表结构，见第 3 节）

3. **更新应用代码**
   - 删除 `entity.ExtraOutbound` 定义，新增 `entity.Outbound`
   - 删除 `service.ExtraOutboundService` 所有实现
   - 新增 `service.OutboundService` 及缓存逻辑
   - 删除 `/api/extra-outbounds` 所有接口
   - 新增 `/api/outbounds` 接口
   - 删除相关前端页面和组件
   - 新增 Outbound 管理前端页面
   - 更新 `service/generated.go` 中的 `resolveGenerateExtraOutbounds()` → `resolveGenerateOutbounds()`
   - 更新 `convert/singbox/outbound.go` 处理缓存逻辑

4. **集成测试**
   - 验证配置生成流程正常
   - 验证设备能获取正确的 Outbound

---

## 4.4 设备可见性统一处理

### 4.4.1 设计目标

让用户对 Outbound 的设备可见性有统一的控制，避免订阅源和手动 Outbound 的可见性规则不一致。当前方案接受 `visible_devices` 冗余到每条订阅缓存 Outbound 的一致性代价。

### 4.4.2 处理流程

**订阅源的 Outbound**：
1. 从 Subscribe 表读取 `visible_devices` 配置
2. 拉取 Outbound 后，**冗余复制** Subscribe 的 `visible_devices` 到 Outbound 表
3. 查询时：按 Outbound 表的 `visible_devices` 过滤

**手动 Outbound**：
1. 用户在 Outbound 编辑页面配置 `visible_devices`
2. 直接保存到 Outbound 表
3. 查询时：同样按 Outbound 表的 `visible_devices` 过滤

### 4.4.3 统一过滤逻辑

```go
// 检查设备对某个 Outbound 是否可见
func isOutboundVisibleToDevice(outbound *entity.Outbound, deviceCode string) bool {
  if outbound.VisibleDevices == "" {
    // 为空表示所有设备可见
    return true
  }

  devices := strings.Split(outbound.VisibleDevices, ",")
  for _, device := range devices {
    if strings.TrimSpace(device) == deviceCode {
      return true
    }
  }
  return false
}

// 查询设备对应的有效 Outbound
func (s *Service) GetOutboundsByDevice(deviceCode string) ([]*entity.Outbound, error) {
  // 从数据库读取所有启用的 Outbound
  outbounds, err := s.storage.ListOutbounds()
  if err != nil {
    return nil, err
  }

  // 过滤：启用 + 对该设备可见
  result := make([]*entity.Outbound, 0)
  for _, out := range outbounds {
    if out.Enabled && isOutboundVisibleToDevice(out, deviceCode) {
      result = append(result, out)
    }
  }
  return result, nil
}
```

### 4.4.4 缓存更新时的冗余

当从订阅源拉取 Outbound 时：

```go
func (s *Service) RefreshSubscriptionOutbound(ctx context.Context, subscribeName string) error {
  sub, err := s.storage.GetSubscribe(subscribeName)
  // ...

  // 拉取原始 Outbound 列表
  outbounds, err := s.fetchOutboundsFromSubscription(ctx, sub)
  // ...

  // 为每个 Outbound 冗余 Subscribe 的 visible_devices
  for _, out := range outbounds {
    out.VisibleDevices = sub.VisibleDevices  // 冗余字段
    out.SubscribeName = sub.Name
    out.Source = "SUBSCRIPTION"
    out.LastFetchTime = time.Now()
  }

  // 最大努力：先删后写，不要求事务
  err = s.storage.DeleteOutboundsBySubscribe(subscribeName)
  if err != nil {
    return err
  }
  err = s.storage.CreateOrUpdateOutbounds(outbounds)
  // ...
}
```

---

## 5. API 接口设计

### 5.1 Outbound 管理接口

#### 5.1.1 列表查询

```
GET /api/outbounds?page=1&limit=20&source=MANUAL&enabled=true&search=keyword&subscribe_name=sub-1

Response:
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "tag": "us-proxy-1",
        "name": "US Proxy 1",
        "type": "http",
        "enabled": true,
        "source": "MANUAL",
        "created_at": "2026-04-12T10:00:00Z",
        "updated_at": "2026-04-12T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20
  }
}
```

#### 5.1.2 创建（仅手动）

```
POST /api/outbounds

Request:
{
  "tag": "us-proxy-1",
  "name": "US Proxy 1",
  "description": "US direct proxy",
  "type": "http",
  "enabled": true,
  "sort": 0,
  "visible_devices": "device-1,device-2",
  "config_json": { ... }
}
```

#### 5.1.3 编辑（仅手动来源）

```
PUT /api/outbounds/{id}

Constraint: 仅允许编辑 source=MANUAL 的记录
```

#### 5.1.4 删除

```
DELETE /api/outbounds/{id}

行为: 硬删除（不考虑软删除，因为这是管理操作）
```

#### 5.1.5 批量启用/禁用

```
PATCH /api/outbounds/batch-enable

Request:
{
  "ids": [1, 2, 3],
  "enabled": true
}
```

### 5.2 订阅相关接口变更

#### 5.2.1 更新订阅缓存配置

```
PUT /api/subscribes/{name}/cache-config

Request:
{
  "outbound_cache_duration": 60  // 分钟
}
```

#### 5.2.2 手动刷新 Outbound

```
POST /api/subscribes/{name}/refresh-outbound

Response:
{
  "code": 200,
  "data": {
    "status": "success",
    "added": 5,
    "deleted": 2,
    "last_fetch_time": "2026-04-12T10:30:45Z"
  }
}
```

#### 5.2.3 查询订阅的 Outbound 列表

```
GET /api/subscribes/{name}/outbounds?page=1&limit=20

Response:
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "tag": "node-1",
        "name": "Node 1",
        "type": "ss",
        "source": "SUBSCRIPTION",
        "subscribe_name": "my-subscribe",
        "last_fetch_time": "2026-04-12T10:00:00Z"
      }
    ],
    "total": 20,
    "subscribe_cache_info": {
      "last_fetch_time": "2026-04-12T10:00:00Z",
      "cache_duration": 60,
      "is_expired": false
    }
  }
}
```

---

## 6. 前端界面设计

### 6.1 Outbound 管理页面

#### 6.1.1 页面布局

```
┌─────────────────────────────────────────────────────────────────┐
│                      Outbound 管理                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│ 筛选: [来源 ▼] [状态 ▼] [订阅 ▼] 搜索: [________]  [搜索]       │
│                                                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│ ┌──┬──────────┬──────────┬──────┬──────────┬──────────┬────────┐ │
│ │✓ │ 标签     │ 名称     │ 类型 │ 来源     │ 状态     │ 操作   │ │
│ ├──┼──────────┼──────────┼──────┼──────────┼──────────┼────────┤ │
│ │  │ us-1     │ USProxy │ http │ 手动     │ 启用 ✓   │ 编辑删除│ │
│ │  │ ss-node  │ SS Node  │ ss   │ 订阅-1   │ 启用 ✓   │ 删除   │ │
│ │  │ block    │ 屏蔽     │ -    │ 手动     │ 禁用 ✗   │ 编辑删除│ │
│ │  │ ...      │ ...      │ ...  │ ...      │ ...      │ ...    │ │
│ └──┴──────────┴──────────┴──────┴──────────┴──────────┴────────┘ │
│                                                                   │
│ [+ 新增 Outbound]  [批量启用]  [批量禁用]   第 1 页 (共 3 页)   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

#### 6.1.2 操作功能

- **新增 Outbound**：仅支持新增 MANUAL 类型
- **编辑 Outbound**：仅支持编辑 MANUAL 类型（源字段不可改）
- **删除 Outbound**：两种来源都可删除
- **批量操作**：支持批量启用/禁用
- **搜索**：按标签、名称搜索
- **筛选**：按来源、订阅源、状态筛选

### 6.2 订阅管理页面改进

#### 6.2.1 新增缓存配置面板

在订阅编辑页面添加 "Outbound 缓存配置" 段落：

```
┌─────────────────────────────────────────────────┐
│ Outbound 缓存配置                               │
├─────────────────────────────────────────────────┤
│                                                 │
│ 缓存时长: [____] 分钟  (0 = 每次都拉取)          │
│                                                 │
│ 最后拉取: 2026-04-12 10:00:00                   │
│ 缓存状态: 有效 (剩余时间: 45 分钟)              │
│                                                 │
│ [立即拉取]  [查看 Outbound]                     │
│                                                 │
└─────────────────────────────────────────────────┘
```

#### 6.2.2 Outbound 列表抽屉

点击 "查看 Outbound" 后弹出侧抽屉，显示该订阅源的所有 Outbound：

```
┌─────────────────────────────────────┐
│ Outbound 列表 (订阅源: xxx)         │
├─────────────────────────────────────┤
│ 共 10 个节点，最后更新: xxx         │
│                                     │
│ • node-1 (ss) - 启用                │
│ • node-2 (vmess) - 启用             │
│ • node-3 (trojan) - 禁用            │
│ • ...                               │
│                                     │
│ [立即刷新]  [关闭]                  │
└─────────────────────────────────────┘
```

---

## 7. 实施步骤

### 关键代码改动点

#### service/generated.go 改动

**当前流程**：
```go
// 第 65 行：直接读取订阅列表
subscribes, err := s.storage.ListSubscribes()

// 第 83 行：调用 GetOutbounds，其中每个订阅都重新拉取
outbounds, err := singbox.GetOutbounds(c.Request.Context(), subscribes, groupRules, extraOutbounds)
```

**需要改动**：
1. 在 `resolveGenerateOutbounds()` 新方法中实现缓存检查逻辑
2. 传递 `deviceCode` 给 GetOutbounds，用于设备可见性过滤
3. 调用新的 `RefreshSubscriptionOutboundIfNeeded()` 方法处理缓存更新

**改动代码示例**：
```go
// 新增方法：处理订阅缓存和设备过滤
func (s *Service) resolveGenerateOutbounds(ctx context.Context, deviceCode string) ([]entity.SingBoxOut, error) {
  subscribes, err := s.storage.ListSubscribes()

  for _, sub := range subscribes {
    // 检查设备可见性
    if !isDeviceVisible(deviceCode, sub.VisibleDevices) {
      continue
    }

    // 检查缓存是否过期，若需要则更新
    if needsRefresh(sub) {
      err := s.RefreshSubscriptionOutbound(ctx, sub.Name)
      if err != nil {
        // 记录错误，继续使用旧缓存
        logrus.WithError(err).Warn("Failed to refresh outbound")
      }
    }
  }

  // 从 outbound 表读取所有有效数据
  return s.GetAllValidOutbounds(deviceCode)
}
```

#### convert/singbox/outbound.go 改动

**当前函数签名**：
```go
func GetOutbounds(ctx context.Context, subscribes []*entity.Subscribe,
  groupRules []*entity.NodeGroup, extraOutbounds []entity.SingBoxOut) ([]entity.SingBoxOut, error)
```

**需要改动**：
1. 改为从 Outbound 表读取缓存数据，而非从订阅源实时拉取
2. 支持设备可见性过滤（对订阅源和 Outbound 都过滤）
3. 简化为 GetOutbounds(outbounds []entity.SingBoxOut, groupRules []*entity.NodeGroup)

**改动代码示例**：
```go
// 旧逻辑：从订阅源实时拉取
for _, subscribe := range subscribes {
  subscriptionBody, err := HttpGetByte(ctx, subscribe.URL, subscribe.UserAgent)
  // 解析并追加
}

// 新逻辑：直接使用缓存
func GetOutbounds(outbounds []entity.SingBoxOut, groupRules []*entity.NodeGroup) ([]entity.SingBoxOut, error) {
  // outbounds 已经是过滤后的缓存数据
  // 只需构建组、追加固定出站
  groupOutbounds := constructOutboundGroup(groupRules, getTagsFromOutbounds(outbounds))
  outbounds = append(outbounds, groupOutbounds...)
  // ...
}
```

#### storage 层改动

**新增接口**：
```go
// 按设备和来源查询 Outbound
ListOutboundsBySource(source string) ([]*entity.Outbound, error)
ListOutboundsBySubscribe(subscribeName string) ([]*entity.Outbound, error)
GetOutboundsByDevice(deviceCode string) ([]*entity.Outbound, error)  // 合并过滤逻辑

// 缓存管理
DeleteOutboundsBySubscribe(subscribeName string) error
CreateOrUpdateOutbounds(outbounds []*entity.Outbound) error
```

---

### Phase 1: 数据库和实体层（第 1 周）

- [x] 完成数据库结构初始化变更
  - 创建 Outbound 表
  - 修改 Subscribe 表
  - 直接删除旧 `extra_outbounds` 表，不保留迁移逻辑

- [x] 更新 Go 实体
  - 定义 Outbound entity
  - 定义 OutboundSource enum
  - 更新 Subscribe entity
  - 删除 ExtraOutbound entity

- [x] 更新 GORM 模型关系
  - Subscribe 与 Outbound 的一对多关系

### Phase 2: 业务逻辑层（第 2 周）

**storage 层改动**：
- [x] 添加 Outbound 表 Repository 接口
  - `ListOutbounds(filters...)` - 按来源、订阅源、启用状态查询
  - `GetOutboundsByDevice(deviceCode)` - 按设备可见性过滤
  - `CreateOrUpdateOutbounds(items)` - 批量创建或按 tag 覆盖
  - `DeleteOutboundsBySubscribe(subscribeName)` - 按订阅源删除
  - `UpdateOutboundCacheTime(subscribeName, timestamp)` - 更新缓存时间

**service 层改动**：
- [x] 实现 Outbound 管理服务
  - 增删改查逻辑（区分 MANUAL 和 SUBSCRIPTION 权限）
  - 批量操作

- [x] 实现缓存更新核心逻辑
  - `RefreshSubscriptionOutbound(subscribeName)` - 从订阅源拉取并更新缓存
  - `needsRefresh(subscribe)` - 判断缓存是否过期
  - `isDeviceVisible(deviceCode, visibleDevices)` - 检查设备可见性
  - 最大努力处理：删除旧 + 插入新/覆盖 + 更新订阅状态
  - 错误处理：拉取失败时降级使用旧缓存

- [x] 修改 `resolveGenerateOutbounds()` 方法
  - 遍历订阅，检查缓存过期状态
  - 调用 RefreshSubscriptionOutbound（若需要）
  - 从 Outbound 表查询有效数据（按设备和启用状态过滤）
  - 合并手动 Outbound

**convert 层改动**：
- [x] 重构 `convert/singbox/outbound.go`
  - 修改 `GetOutbounds()` 函数签名
    - 输入改为：已缓存的 Outbound 列表 + 节点分组规则
    - 删除：订阅列表、HTTP 拉取逻辑
  - 保留：构建分组、追加固定出站的逻辑
  - 新增设备可见性过滤

- [x] 清理过时代码
  - 删除 `HttpGetByte()` 等拉取逻辑（移至 service 层）
  - 删除 `parseBodyContent()` 等解析逻辑（移至 service 层）

### Phase 3: API 接口层（第 3 周）

- [x] 实现 Outbound API endpoints
  - GET /api/outbounds（列表）
  - POST /api/outbounds（创建）
  - PUT /api/outbounds/{id}（编辑）
  - DELETE /api/outbounds/{id}（删除）
  - PATCH /api/outbounds/batch-enable（批量操作）

- [x] 实现 Subscribe 缓存 API
  - POST /api/subscribes/{name}/refresh-outbound（手动刷新）
  - PUT /api/subscribes/{name}/cache-config（更新缓存配置）
  - GET /api/subscribes/{name}/outbounds（查询列表）

### Phase 4: 前端界面（第 4 周）

- [x] Outbound 管理页面
  - 列表组件（分页、搜索、筛选）
  - 新增/编辑对话框
  - 批量操作

- [x] 订阅管理页面改进
  - 缓存配置面板
  - 手动刷新按钮
  - Outbound 列表抽屉

### Phase 5: 文档（第 5 周）


- [x] 更新文档
  - [x] 模块文档 (modules/outbound.md)
  - [x] API 文档
  - [x] 配置生成流程文档
  - [x] 将此文档从 planned/ 移至 implemented/

---

## 8. 风险和考虑

### 8.1 性能风险

**风险**：大量 Outbound 数据导致查询变慢

**缓解措施**：
- 在 source、subscribe_name、enabled 字段建立索引
- 实现分页查询
- 添加缓存层（Redis）

### 8.2 数据一致性风险

**风险**：并发更新导致数据不一致

**缓解措施**：
- 接受最大努力写入下的短暂不一致窗口
- 刷新完成后以最终落库结果为准
- 增加刷新日志与错误状态记录，便于排查

### 8.3 迁移风险

**风险**：数据迁移期间服务不可用或数据丢失

**缓解措施**：
- 由于项目未上线，直接切换到新结构，不设计灰度迁移
- 实现前在本地和测试环境验证初始化、建表和导入导出流程
- 保持迁移步骤简单，避免保留双模型

### 8.4 缓存过期风险

**风险**：使用过期的缓存数据生成配置

**缓解措施**：
- 清晰的过期判断逻辑
- 拉取失败时有降级策略
- 缓存可手动刷新
- 记录拉取状态便于调试

### 8.5 向后兼容性

**风险**：现有代码与新表结构不兼容

**缓解措施**：
- 本需求不要求向后兼容，直接删除旧模型和旧接口
- 通过统一替换调用点，避免代码中长期共存两套实现

---

## 9. 后续优化方向

### 9.1 短期优化

1. **后台定时任务**：支持定期自动刷新（而不仅是手动刷新）
2. **缓存统计**：显示缓存命中率、拉取成功率等指标
3. **版本管理**：支持查看历史版本和回滚

### 9.2 中期优化

1. **智能缓存策略**：根据拉取频率自动调整缓存时长
2. **增量更新**：支持增量 diff，而不是全量替换
3. **节点健康检查**：定期检测 Outbound 可用性

### 9.3 长期优化

1. **分布式缓存**：使用 Redis 实现分布式缓存
2. **CDN 加速**：缓存订阅源在 CDN
3. **机器学习**：智能预测缓存过期时间

---

## 10. 参考资源

- [项目概述](../architecture/overview.md)
- [配置生成流程](../architecture/config-generation.md)
- [Outbound 管理](../modules/outbound.md)
- [订阅管理](../modules/subscribe.md)
- [存储抽象层](../architecture/storage-layer.md)

---

**文档版本历史**：
- v1.0 (2026-04-12): 初稿完成
- v1.1 (2026-04-13): 完成 Phase 1-5 实施并同步文档到 implemented
