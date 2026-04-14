# 前端架构

当前前端是一个不使用 React Router 的单页管理台，入口在 [App.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/App.tsx)。

## 技术栈

- React 18
- TypeScript
- Vite
- Arco Design React
- Axios
- Monaco Editor
- `vite-plugin-singlefile`

## 整体结构

前端目录主要由四层组成：

- `pages/`：页面级容器，负责数据加载、保存、删除、弹窗状态
- `components/`：可复用 UI 组件与弹窗
- `api/`：所有 HTTP 调用封装
- `utils/`：纯函数工具，如 JSON 规范化、DNS 默认模板、导入摘要格式化

## 路由与导航

当前没有 URL 路由系统。

页面切换方式：

- 左侧菜单项定义在 [navigation.ts](/Users/xsdhy/data/code/go/singboxconfig/web/src/utils/navigation.ts)
- `App.tsx` 用 `activeKey` 本地状态切换页面
- `renderPage()` 用 `switch` 决定渲染哪个页面组件

这意味着：

- 页面刷新不会保留菜单位置到 URL
- 没有浏览器级前进/后退路由语义
- 管理台更接近一个“控制面板应用”而不是多路由站点

## 页面组织方式

各页面基本遵循同一种模式：

1. `useEffect` 初始加载列表或基础数据
2. 顶部统一使用 `PageToolbar`
3. 空状态统一使用 `DataState`
4. 列表展示采用卡片而非表格
5. 新增/编辑通过 `Modal` 或 `Drawer`
6. 删除统一通过 `Modal.confirm`

典型页面分两类：

- 简单 CRUD：订阅、节点分组、规则集、设置
- JSON 模板型：Inbound、额外出站、DNS

## 状态管理

当前项目没有引入 Redux、Zustand、React Query 一类外部状态库。

状态管理方式是：

- 页面内局部 `useState`
- 列表数据由页面自己加载和刷新
- 子组件通过 props 接收数据和回调

优点：

- 结构直接
- 维护成本低

代价：

- 页面之间没有共享缓存
- 导入配置后不会自动广播各页面刷新，需要用户手动切换或刷新

## API 调用层

[api/index.ts](/Users/xsdhy/data/code/go/singboxconfig/web/src/api/index.ts) 中创建了一个裸 `axios.create()` 实例，并按资源导出方法。

当前特点：

- 没有统一 `baseURL`
- 没有请求/响应拦截器
- 依赖浏览器当前域名和 Vite 代理
- 路径参数会先走 `encodeURIComponent`

## Monaco Editor 集成

以下页面使用 Monaco 编辑 JSON：

- DNS
- Inbound 管理
- 额外出站管理
- 规则集管理中的本地规则集编辑

使用方式比较统一：

- `editorRef.current?.getValue()` 读取当前文本
- 保存前通过 `normalizeJsonText` 或 `JSON.parse` 做校验
- 回显时通过 `prettyJsonText` 规范缩进

## UI 布局

应用主布局来自 `Layout + Sider + Content`：

- 左侧：菜单导航
- 右侧：当前页面说明、导入导出按钮、正文区域

页面顶部说明文字和统计信息由 `PageToolbar` 统一渲染。

## 与后端耦合点

前端和后端目前是强耦合的同仓结构，主要体现在：

- Vite 构建输出直接写到 `cmd/server/`
- 类型命名和 JSON 字段基本一一对应后端实体
- 菜单项和后端资源模块高度对齐
- DNS 页面直接读写 `settings` 中的 `dns_config`

## 当前架构限制

### 1. 无路由持久化

菜单状态完全在内存中，刷新页面会回到默认页。

### 2. 错误处理较轻

大多数请求失败只显示一条通用 `Message.error`，没有统一错误抽象。

### 3. 接口契约有少量历史偏差

例如 `getSettings()` 的返回类型在前端写成 `Setting[]`，但后端实际返回 `map[string]string`。

### 4. 认证已迁移到前端登录态

前端现在提供独立登录页，并在本地存储 Bearer Token。`web/src/api/index.ts` 会统一把 token 注入到后续 `/api/*` 请求中，账号设置支持修改用户名和密码。

## 相关文档

- [页面说明](./pages.md)
- [API 客户端](./api-client.md)
- [项目概述](../architecture/overview.md)
