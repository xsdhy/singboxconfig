# API 客户端

当前前端的 API 封装集中在 [api/index.ts](/Users/xsdhy/data/code/go/singboxconfig/web/src/api/index.ts)。

## 基本实现

代码使用：

```ts
const api = axios.create();
```

特点：

- 没有设置 `baseURL`
- 没有请求拦截器
- 没有响应拦截器
- 没有统一错误转换层

因此它依赖：

- 开发环境使用 Vite 代理
- 生产环境前后端同域部署

## 路径参数处理

封装里定义了：

```ts
const encodePathSegment = (value: string | number) => encodeURIComponent(String(value));
```

所有带 `tag`、`name`、`code`、`id` 的接口都会先做路径编码，避免特殊字符破坏 URL。

## 模块划分

当前 API 大致分为以下几组：

- `subscribes`
- `node-groups`
- `rule-sets`
- `settings`
- `devices`
- `inbounds`
- `wire-guards`
- `outbounds`
- `config-transfer`

其中 `subscribes` 组里还额外包含了订阅缓存相关接口：

- `updateSubscribeCacheConfig(name, duration)`
- `refreshSubscribeOutbounds(name)`
- `listSubscribeOutbounds(name, page, limit)`

## 导入导出接口

### 导出

`exportConfig()` 会以 `blob` 形式请求：

```ts
api.get<Blob>('/api/config-transfer/export', { responseType: 'blob' })
```

前端随后从 `Content-Disposition` 里解析文件名。

### 导入

`importConfig(file)` 使用 `FormData` 上传：

```ts
const formData = new FormData();
formData.append('file', file);
```

接口路径：

- `POST /api/config-transfer/import`

## 设备与绑定关系

设备管理被拆成两组接口：

- 设备本身：`/api/devices`
- 设备与 Inbound 绑定：`/api/devices/:code/inbounds`

这种拆分与后端实体一致，也让页面能分别维护基础信息和绑定抽屉。

## 当前缺少的能力

和很多中大型前端项目相比，当前客户端层还没有：

- 自动注入认证头
- 统一超时和重试
- 标准化错误对象
- 接口缓存
- 请求取消

这与项目当前规模一致，但如果后续页面和接口继续增加，建议把这些能力补进一层独立的 HTTP 封装。

## 与后端契约需要注意的地方

### 1. `getSettings` 的类型声明偏差

前端把 `getSettings()` 声明为返回 `Setting[]`，但后端当前实现返回 `map[string]string`。页面能否正常运行取决于调用方是否自行转换。

### 2. 公开接口没有被封装到这里

当前 `api/index.ts` 只封装管理端接口，没有封装：

- `/open/generate/:device`
- `/open/surge/:device`
- `/open/shadowrocket/:device`
- `/open/ruleset/:tag`

设备页展示的 sing-box、Surge 与 Shadowrocket 订阅链接由页面根据设备 `code` 和 `token` 直接拼接，不经过 Axios client。

## 建议的扩展方向

如果继续演进 API 客户端，优先级建议是：

1. 增加统一响应错误处理
2. 明确 `settings` 的返回类型
3. 抽出管理端与公开接口两个 client
4. 增加基础的超时与日志

## 相关文档

- [前端架构](./architecture.md)
- [页面说明](./pages.md)
- [API 接口列表](../reference/api-reference.md)
