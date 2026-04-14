# 技术栈

## 技术栈总览

项目采用 Go 后端 + React 前端 + 可切换存储后端的组合。核心原则是：

- 后端尽量保持单二进制部署
- 前端提供完整管理能力，但不引入过重状态管理框架
- 数据模型在服务层和存储层之间保持稳定
- 配置生成逻辑与持久化逻辑解耦

## 后端技术选型

### Go 1.25

`go.mod` 指定 `go 1.25.0`。Go 作为主语言的优势在于：

- 单二进制部署简单
- 标准库足够支撑 HTTP、JSON、并发与文件处理
- 对配置生成这类结构化数据拼装工作非常合适

### Gin

后端 HTTP 框架使用 `github.com/gin-gonic/gin v1.10.0`。

在本项目中，Gin 主要承担：

- 路由注册
- JSON 参数绑定
- 统一返回 JSON 响应
- Bearer Token 认证中间件

当前项目没有引入复杂中间件体系，也没有自定义路由层抽象，保持了较低复杂度。

### GORM

ORM 使用 `gorm.io/gorm v1.31.1`，数据库驱动依赖包含：

- `gorm.io/driver/postgres`
- `gorm.io/driver/mysql`

当前实际情况需要区分：

- 抽象层上，`DatabaseStorage` 是通用 GORM 实现，理论上可接 PostgreSQL / MySQL
- 启动层上，`cmd/server/main.go` 当前只在 `DATABASE_URL` 分支中调用 `postgres.Open(...)`

也就是说，数据库存储抽象是多驱动友好的，但默认启动入口当前只接通了 PostgreSQL。

### Logrus

日志使用 `github.com/sirupsen/logrus v1.9.3`，主要出现在：

- 订阅抓取与解码日志
- DNS / Inbound / Extra Outbound 非法 JSON 的降级告警

项目中的日志用途偏运行诊断，而不是复杂的审计或结构化观测。

## 存储技术选型

项目采用存储抽象设计，支持两种后端实现，可根据部署场景灵活选择：

### 存储后端对比

| 存储类型 | 推荐场景 | 优点 | 缺点 | 持久化策略 |
|---------|---------|------|------|----------|
| **DatabaseStorage** | 生产环境、多实例共享、高可用需求 | 实时持久化、多实例支持、SQL 可查审 | 需要外部数据库、连接池管理复杂 | 每次操作实时提交到数据库 |
| **SupabaseStorage** | 云部署、无数据库运维、Serverless 场景 | 无需管理基础设施、REST API 简单、托管 PostgreSQL | 网络延迟、API 配额限制、调试困难 | 通过 REST API 实时持久化 |

### 选择建议

- **开发阶段**：使用 `DatabaseStorage` + PostgreSQL/MySQL
- **自建部署**：使用 `DatabaseStorage` + PostgreSQL/MySQL，获得最好的性能和可靠性
- **云托管**：使用 `SupabaseStorage`，简化运维复杂度
- **生产多副本**：必须使用 `DatabaseStorage` 或 `SupabaseStorage`，避免数据不一致

### DatabaseStorage

### DatabaseStorage

`storage/database.go` 通过 GORM 实现关系型数据库存储，启动时自动迁移表结构。

适合：

- 持久化要求更高的生产场景
- 需要多实例共享数据的部署
- 希望通过 SQL 工具直接审查数据的场景

### SupabaseStorage

`storage/supabase.go` 直接通过 Supabase PostgREST HTTP API 实现 `Storage` 接口。

设计意图是：

- 不依赖项目内直连数据库驱动和连接池
- 利用托管 Postgres + REST API 降低部署复杂度
- 保持与本地实体模型基本一致的字段结构

代价是：

- 需要手写 HTTP 请求和查询参数
- 调试复杂度高于本地 ORM
- 网络抖动会直接影响后台 CRUD 和生成接口

## 配置生成相关技术

### 自定义 sing-box 结构体

`entity/singbox.go` 定义了一套接近 sing-box JSON 结构的 Go struct，而不是完全依赖 map[string]any。这样做的收益是：

- 输出结构更稳定
- 字段意图更清晰
- 测试可以直接做结构比较

### 自研协议解码器

`protocol/` 目录实现了：

- SS
- SSR
- Trojan
- VMess

当前生成链路在 `convert/singbox/outbound.go` 中启用了：

- `ss`
- `trojan`
- `vmess`

`ssr` 代码和测试存在，但尚未加入 `convertMap`。

## 前端技术选型

### React 18 + TypeScript 5

前端位于 `web/`，采用 React 18.3.1 和 TypeScript 5.5。

当前前端特征：

- 单页应用
- 无 React Router
- 以页面组件 + 表格/弹窗组件为主
- 本地状态使用 `useState`、`useMemo` 即可支撑

这与项目后台控制台的场景匹配，避免了过度设计。

### Vite

构建工具使用 Vite 5.4，配置位于 `web/vite.config.ts`。

关键点：

- 开发时将 `/api` 代理到 `http://localhost:7391`
- 构建输出目录是 `../cmd/server`
- 使用 `vite-plugin-singlefile` 将前端产物收敛为单 HTML

这让后端可以通过 `//go:embed index.html` 直接嵌入前端产物。

### Arco Design

UI 组件库使用 `@arco-design/web-react`。项目中的主要用途是：

- 表格
- 表单
- Modal
- Message
- Layout 与 Menu

选型原因很务实：后台管理界面组件成熟，能快速搭建 CRUD 页面。

### Axios

HTTP 客户端使用 Axios。`web/src/api/index.ts` 统一封装所有接口：

- 统一路径规则
- 路径参数统一 `encodeURIComponent`
- 导入导出接口单独处理 `blob` 和 `multipart/form-data`

## 构建与部署工具链

### Makefile

`Makefile` 的默认链路是：

1. 进入 `web/`
2. `npm install`
3. `npm run build`
4. `go build -o singboxconfig cmd/server/*.go`

即先构建前端，再构建后端，最终得到包含管理界面的可执行文件。

### Dockerfile

`Dockerfile` 基于 `golang:1.25.0-alpine`，使用多阶段构建：

1. 前端构建阶段：使用 Node.js 构建前端资源
2. 后端构建阶段：使用 Go 编译后端二进制（CGO_ENABLED=0，纯静态编译）
3. 运行时镜像：基于 alpine，仅包含必要的运行时依赖

## 为什么是这套组合

- Go + Gin：足够稳定，部署简单，适合配置型服务
- GORM：减少样板 CRUD 代码，便于统一模型迁移
- Memory / DB / Supabase 并存：兼顾本地运行、传统数据库部署、云托管后端
- React + Arco + Vite：适合内部管理台，开发速度高
- 单文件前端嵌入：降低部署复杂度，避免额外静态资源服务

## 当前技术债

- 数据库驱动在抽象层和启动层之间还没完全对齐
- 单管理员认证仍较轻量，尚未扩展到多用户和 RBAC
- 前端未引入更细粒度的数据缓存或错误恢复机制
- 订阅拉取仍是请求时同步进行，缺少缓存与异步刷新机制

这些问题不影响当前可用性，但决定了系统在规模扩大后的演进方向。
