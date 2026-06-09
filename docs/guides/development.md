# 开发指南

本文档面向需要继续开发当前项目的后端或前端开发者。

## 开发模式概览

推荐本地开发方式：

1. 后端在仓库根目录运行 `go run ./cmd/server`
2. 前端在 `web/` 目录运行 `npm run dev`
3. 浏览器访问 `http://localhost:5173`

Vite 会代理 `/api` 到 `http://localhost:7391`。

## 代码结构

后端核心分层：

- `cmd/server`：启动入口、路由注册、嵌入式页面
- `service`：HTTP Handler + 业务编排
- `storage`：存储抽象与各后端实现
- `entity`：核心领域对象与 sing-box 输出结构
- `convert/singbox`：从存储实体到 sing-box JSON 的转换逻辑
- `convert/surge`：从同一批存储实体到 Surge 文本配置的转换逻辑
- `convert/shadowrocket`：从同一批存储实体到 Shadowrocket 文本配置的转换逻辑
- `convert/common`：多输出格式共享的纯函数辅助逻辑
- `protocol`：订阅协议 URL 解码器

前端核心分层：

- `web/src/pages`：各管理页
- `web/src/components`：通用卡片、弹窗、工具栏
- `web/src/api`：Axios API 封装
- `web/src/types`：前端类型定义
- `web/src/utils`：JSON、DNS、导入导出、绑定关系等辅助函数

## 常用开发命令

后端：

```bash
go run ./cmd/server
go test ./...
```

前端：

```bash
cd web
npm install
npm run dev
npm run build
npm test
```

整体构建：

```bash
make build
```

## 前后端联调

### 前端开发服务器

[vite.config.ts](/Users/xsdhy/data/code/go/singboxconfig/web/vite.config.ts) 中配置了：

- `outDir: ../cmd/server`
- `server.proxy['/api'] = 'http://localhost:7391'`

这意味着：

- 开发阶段前端页面由 Vite 提供
- 所有 `/api/*` 请求都会代理到后端
- `/open/*` 不在代理配置中，如需在前端直接调试公开接口，需要手动拼后端地址

### 管理端认证

除 `POST /api/auth/login` 和 `GET /api/admin` 外，管理接口都需要 Bearer Token。联调时如果出现 `401`，优先检查前端本地 token 是否存在且请求头是否带了 `Authorization: Bearer ...`。

## 新功能开发建议

### 新增一个后端资源模块

通常需要同时改动：

1. `entity/`：新增实体
2. `storage/storage.go`：补接口定义
3. `storage/*.go`：实现 Memory / Database / Supabase
4. `service/`：新增 CRUD Handler 与路由
5. `web/src/types`：补前端类型
6. `web/src/api`：补 API 调用
7. `web/src/pages`：补管理页面
8. `docs/modules/` / `docs/reference/`：同步更新文档

### 新增一个会参与配置生成的模块

除上面的步骤外，还要检查：

- 是否需要在 `service/generated.go` 中接入
- 是否需要在 `convert/singbox/` 中增加转换逻辑
- 是否需要在 `convert/surge/` 中同步支持或显式降级
- 是否需要在 `convert/shadowrocket/` 中同步支持或显式降级
- 是否需要修改设备过滤或排序逻辑

## 测试现状

当前仓库已有 Go 测试与部分前端编译测试，但并非全部测试都与当前实现一致。

已知不一致点：

- `service` 的部分测试仍期待旧默认设备 `phone`

因此开发时建议：

- 先运行 `go test ./...`
- 区分“本次改动引入的问题”和“仓库已有预期偏差”

## 代码风格与实现习惯

从当前仓库代码可见的约定：

- 服务层直接使用 Gin `Context`
- 存储层返回 `storage.ErrNotFound` 表示不存在
- 列表接口通常会在服务层按 `sort` 再排序
- JSON 模板类资源会在前端先做格式化/校验，再提交后端
- 生成链路倾向于“跳过非法项并继续生成”，而不是整单失败

## 开发时容易踩坑的点

### 1. `ListSettings` 返回的是对象，不是数组

后端 `GET /api/settings` 当前直接返回 `map[string]string`，而前端类型里把它声明成 `Setting[]`。如果继续改这块，先统一接口契约。

### 2. 规则集更新会移除 `content` 里的换行和空格

[service.go](/Users/xsdhy/data/code/go/singboxconfig/service/service.go) 的 `UpdateRuleSet` 会对 `Content` 做字符串压缩，这会影响可读性，但不影响 JSON 语义。

### 3. 入口代码未真正接通 MySQL

虽然依赖里有驱动，当前运行入口只在 `DATABASE_URL` 分支里使用 PostgreSQL。

### 4. 前端打包产物写入 `cmd/server/`

如果你清理或覆盖了这个目录，可能影响后端嵌入式页面。

## 文档维护要求

按 [INDEX.md](../INDEX.md) 的组织方式，功能变化后至少同步更新：

- 模块文档
- API 文档
- 配置项文档
- 相关前端文档

## 相关文档

- [前端架构](../frontend/architecture.md)
- [页面说明](../frontend/pages.md)
- [API 客户端](../frontend/api-client.md)
- [API 接口列表](../reference/api-reference.md)
