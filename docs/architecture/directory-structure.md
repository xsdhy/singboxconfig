# 目录结构

## 顶层结构

```text
singboxconfig/
├── cmd/
│   └── server/                 # 服务入口、Gin 路由、嵌入式前端产物
├── convert/
│   └── singbox/                # 将业务实体转换为 sing-box 配置
├── docs/                       # 当前文档中心
├── entity/                     # 领域实体与 sing-box 输出结构
├── protocol/                   # 订阅协议 URL 解码器
├── service/                    # HTTP 服务层与业务编排
├── storage/                    # 存储抽象接口与多后端实现
├── transfer/                   # 导入导出数据结构
├── web/                        # React 管理前端
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

整体上，目录划分围绕“入口、领域、转换、存储、前端”五个层次展开，边界比较清楚。

## 后端目录说明

### `cmd/server/`

服务启动入口，包含：

- `main.go`：读取环境变量、选择存储实现、创建 Gin 路由
- `index.html`：前端构建产物，由 Vite 输出并被 `//go:embed` 嵌入
- `main_test.go`：路由注册层面的测试

这是部署视角的顶层入口，不承载具体业务逻辑。

### `service/`

服务层负责 HTTP 请求编排，是后端的核心应用层。主要文件：

- `service.go`：订阅、节点分组、规则集、全局设置等基础 CRUD
- `generated.go`：开放配置生成接口，是最关键的运行时主流程
- `device_management.go`：设备、Inbound、WireGuard、额外出站等管理接口
- `config_transfer.go`：导入导出、默认数据初始化、导入摘要统计

这一层的职责是：

- 做 JSON 绑定和路径参数校验
- 调用 `storage` 获取或写入实体
- 调用 `convert/singbox` 组装最终配置

### `storage/`

持久化抽象层，分为三部分：

- `storage.go`：`Storage` 总接口及各子接口
- `memory.go`：基于单 JSON 文件的内存存储
- `database.go`：基于 GORM 的数据库实现
- `supabase.go`：基于 Supabase REST API 的 HTTP 存储实现
- `models.go`：GORM 数据模型与实体映射

这一层的目标是让服务层不感知底层持久化方式。

### `entity/`

领域实体层，定义系统内稳定的数据结构：

- 业务实体：`Subscribe`、`NodeGroup`、`RuleSet`、`Device`、`Inbound`、`WireGuard` 等
- sing-box 输出结构：`SingBoxConfig`、`SingRoute`、`SingDNS`、`SingBoxOut` 等
- 设备常量：如 `phone`、`tv`

这里不是数据库模型层，而是跨服务层、存储层、转换层共享的统一结构层。

### `convert/singbox/`

把后台维护的业务实体转换为 sing-box 最终配置。主要文件：

- `outbound.go`：订阅抓取、协议解码、分组出站构造
- `route.go`：规则集与路由规则组装
- `dns.go`：DNS JSON 解析与默认值回退
- `inbound.go`：Inbound 模板 JSON 反序列化
- `endpoint.go`：WireGuard endpoint 生成
- `extra_outbound.go`：额外出站过滤与注入
- `experimental.go`：按设备输出 experimental 配置
- `device_management_defaults.go`：空存储兼容默认值

这是项目里最接近”配置编译器”的部分。

### `convert/clashx/` (预留)

该目录目前为空，作为未来扩展预留。设计意图是将业务实体转换为 Clash 配置格式。

当前项目仅支持 sing-box 配置输出，Clash 格式转换尚在规划阶段。

### `protocol/`

订阅协议解析器目录。职责单一：

- 输入：单条订阅 URL
- 输出：`entity.SingBoxOut`

当前包含 SS、SSR、Trojan、VMess 的解析与测试。

### `transfer/`

导入导出数据结构定义。它并不直接操作 HTTP 或存储，而是作为：

- `service/config_transfer.go` 的输入输出模型
- 全量配置迁移文件的结构约定

## 前端目录说明

### `web/src/`

前端源码主目录。

#### `web/src/pages/`

按业务页面拆分的管理页：

- `SubscribeManage.tsx`
- `NodeGroupManage.tsx`
- `RuleSetManage.tsx`
- `SettingManage.tsx`
- `DeviceManage.tsx`
- `InboundManage.tsx`
- `WireGuardManage.tsx`
- `ExtraOutboundManage.tsx`
- `DnsManage.tsx`

页面之间不是路由关系，而是由 `App.tsx` 中的侧边菜单切换。

#### `web/src/components/`

复用组件目录，主要是表格、弹窗、工具栏与状态展示组件，例如：

- `SubscribeTable.tsx` / `SubscribeModal.tsx`
- `NodeGroupTable.tsx` / `NodeGroupModal.tsx`
- `RuleSetTable.tsx` / `RuleSetModal.tsx`
- `PageToolbar.tsx`
- `DataState.tsx`

#### `web/src/api/`

接口访问封装层，目前集中在 `index.ts` 一个文件中，负责统一所有 REST API 调用。

#### `web/src/utils/`

前端工具函数，包括：

- 导航菜单定义
- 导入导出结果格式化
- DNS 文本辅助处理
- 设备管理辅助逻辑
- JSON 格式化与校验
- 删除确认逻辑

#### `web/src/types/`

TypeScript 类型定义，与后端实体结构保持基本对应。

### 其它前端关键文件

- `web/src/main.tsx`：前端入口
- `web/src/App.tsx`：应用壳层、侧边菜单、导入导出按钮、页面切换逻辑
- `web/src/App.css`：整体界面样式
- `web/vite.config.ts`：构建输出、开发代理、单文件打包配置
- `web/tests/`：前端工具函数测试

## 文档与历史设计目录

### `docs/`

当前文档中心，按：

- `architecture/`：架构设计文档
- `modules/`：核心模块说明
- `frontend/`：前端相关文档
- `reference/`：参考资料与接口说明
- `guides/`：快速开始与部署指南
- `requirements/`：需求跟踪，分为已实现、计划中、Bug修复三类

组织内容，面向人类和 AI 阅读。

### `docs/requirements/`

需求跟踪系统，按实施状态分类：

- `implemented/`：已实现的功能需求文档
- `planned/`：计划中的功能需求
- `bugfixes/`：Bug 排查与修复过程记录

每个需求单独一个文件，便于追踪演进过程。

## 测试分布

测试文件基本与实现代码同目录放置：

- `convert/singbox/*_test.go`
- `service/*_test.go`
- `storage/*_test.go`
- `protocol/*_test.go`
- `cmd/server/main_test.go`
- `web/tests/*.test.mjs`

这种布局适合小中型项目，便于在修改某一层时就近补测试。

## 目录设计的优点

- 领域实体与存储模型分离，避免 ORM 污染业务结构
- 转换逻辑独立成 `convert/singbox`，生成链路容易定位
- 前端与后端清晰隔离，但通过构建产物实现一体部署
- 历史文档与当前文档分仓放置，不互相污染

## 当前可以继续优化的地方

- `service/` 目前承担了全部 HTTP handler，未来继续增长时可以按模块拆分更多文件
- `web/src/api/index.ts` 已经偏大，后续可按模块拆为多个 API 文件
- `cmd/server/` 当前同时存放源码和前端构建产物，部署方便，但会混合“手写代码”和“生成文件”
