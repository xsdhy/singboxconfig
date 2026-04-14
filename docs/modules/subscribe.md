# 订阅管理

## 模块职责

订阅管理负责维护远程节点订阅源，并在配置生成时把订阅内容转换为 sing-box `outbounds`。它覆盖两部分能力：

- 管理端 CRUD：增删改查订阅源
- 生成端解析：拉取远程文本、拆分节点 URL、调用协议解码器转换为 sing-box 出站

代码入口：

- 实体定义：`entity/subscribe.go`
- 服务接口：`service/service.go`
- 生成转换：`convert/singbox/outbound.go`
- 协议解析：`protocol/*.go`

## 数据模型

后端实体 `entity.Subscribe`：

- `name`：订阅名称，也是后台主键
- `url`：远程订阅地址
- `userAgent`：可选请求头，留空时回退到默认桌面浏览器 UA
- `status`：是否启用；只有启用的订阅才会参与生成

这个模型本身不存储节点明细。项目采用“生成时实时拉取并解析”的方式，节点不会持久化到存储层。

## 管理接口

管理接口挂在 Bearer Token 保护下的 `/api/subscribes`：

- `POST /api/subscribes`：创建订阅
- `PUT /api/subscribes/:name`：更新订阅
- `DELETE /api/subscribes/:name`：删除订阅
- `GET /api/subscribes`：列出订阅

服务层基本是薄封装，直接读写 `storage.SubscribeStorage`。这一层没有额外校验 URL 可达性，也不会在保存时预解析订阅内容。

## 配置生成中的处理流程

生成接口 `/open/generate/:device?token=...` 会调用 `convert/singbox.GetOutbounds()`，订阅处理流程如下：

1. 读取全部订阅
2. 跳过 `status=false` 的订阅
3. 对每个订阅发起 HTTP GET
4. 若订阅响应可整体 Base64 解码，则先解码；否则按原文处理
5. 按换行拆分每个节点 URL
6. 根据 `scheme://` 识别协议并交给对应解码器
7. 把成功解析的节点追加到最终 `outbounds`

失败行为是“单订阅降级、继续整体生成”：

- 拉取失败：记录日志，跳过该订阅
- 内容解析失败：记录日志，跳过该订阅
- 单个节点不支持或解码失败：跳过该节点

这意味着只要还有其他订阅或额外出站可用，整个设备配置仍可继续生成。

## 请求行为

HTTP 拉取逻辑位于 `convert/singbox/outbound.go`：

- 超时时间：30 秒
- 默认 UA：桌面 Chrome UA
- 非 200 状态码视为失败

`userAgent` 仅影响订阅拉取请求，不会进入最终 sing-box 配置。

## 与其他模块的关系

- 与“协议解码器”模块强耦合：节点 URL 最终由 `protocol` 包转换
- 与“节点分组”模块耦合：解析出来的节点标签会继续参与分组筛选
- 与“额外出站”模块并列：两者共同组成候选 `outbounds`

## 当前支持范围与限制

代码现状需要明确区分“解码器存在”和“生成链路已接通”：

- 已接入生成链路：`ss`、`trojan`、`vmess`
- 已实现解码器但当前未接入生成链路：`ssr`

也就是说，`protocol/ssr.go` 可以单独解析 SSR URL，但 `convertMap` 当前注释掉了 `ssr`，因此 SSR 节点不会出现在生成结果中。

另外还有几个实现边界：

- 不做订阅结果缓存，每次生成都会重新拉取
- 不做重复节点去重，多个订阅可产生同名标签
- 不做订阅健康检查或最后拉取状态持久化
- 不支持订阅级代理、重试策略、ETag/If-None-Match 等优化

## 适合更新本文档的场景

以下变更需要同步更新本文档：

- 新增或删除支持的订阅协议
- 修改订阅拉取策略、超时或 UA 逻辑
- 增加节点缓存、去重或预解析能力
- 调整订阅接口路径或校验规则
