# 页面说明

本文档按当前 `web/src/pages/` 中的实际页面组织说明管理台功能。

## 页面入口

当前菜单定义在 [navigation.ts](/Users/xsdhy/data/code/go/singboxconfig/web/src/utils/navigation.ts)，共有 9 个页面：

- 订阅管理
- 节点分组
- 规则集管理
- 全局设置
- 设备管理
- Inbound 管理
- WireGuard 管理
- 额外出站
- DNS 设置

## 订阅管理

文件：

- [SubscribeManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/SubscribeManage.tsx)

功能：

- 加载订阅列表
- 新增、编辑、删除订阅
- 录入订阅名称、URL、User-Agent、启用状态

特点：

- 这是生成节点列表的起点
- 页面本身不执行“测试订阅”或“手动拉取”，生成链路在请求配置时才会拉取订阅

## 节点分组

文件：

- [NodeGroupManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/NodeGroupManage.tsx)

功能：

- 新增、编辑、删除分组
- 配置 `selector` 或 `urltest`
- 录入 `include` / `exclude` 关键字和测速 URL

特点：

- 页面维护的是“筛选规则”，不是手工节点列表

## 规则集管理

文件：

- [RuleSetManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/RuleSetManage.tsx)
- [RuleSetModal.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/components/RuleSetModal.tsx)

功能：

- 管理本地和远程规则集
- 维护 `outbound`、`downloadDetour`、`sort`
- 本地规则集通过 Monaco 编辑 JSON

特点：

- 页面会先并行加载规则集和节点分组，用于下拉选择出站
- 本地规则集要求 JSON 合法

## 全局设置

文件：

- [SettingManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/SettingManage.tsx)

功能：

- 管理普通 key/value 全局配置
- 新增、编辑、删除设置项

注意：

- DNS 页面虽然也使用全局设置存储，但不通过这个页面编辑，而是单独走 JSON 编辑器页面

## 设备管理

文件：

- [DeviceManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/DeviceManage.tsx)

功能：

- 管理设备基础信息
- 设置设备 token、启用状态、排序
- 绑定 WireGuard 模板
- 维护 WireGuard 客户端地址和私钥
- 通过抽屉配置设备可用的 Inbound 列表

特点：

- 页面初始化时会并行加载 `devices`、`inbounds`、`wire-guards`
- 设备与 Inbound 的绑定是“全量替换提交”

## Inbound 管理

文件：

- [InboundManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/InboundManage.tsx)

功能：

- 管理可复用的 Inbound 模板
- 使用 Monaco 编辑原始 JSON
- 配置 `enabled`、`sort`、`type`、说明等字段

特点：

- 实际生成时会先按设备绑定筛选，再按 `sort` 输出

## WireGuard 管理

文件：

- [WireGuardManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/WireGuardManage.tsx)

功能：

- 管理 WireGuard 模板
- 管理模板下的 Peer 列表
- 支持新增、编辑、删除模板和 Peer

特点：

- 页面把“模板”和“Peer”拆成两层维护
- 设备侧只绑定模板 tag，客户端地址和私钥仍保存在设备上

## 额外出站

文件：

- [ExtraOutboundManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/ExtraOutboundManage.tsx)

功能：

- 管理订阅之外的静态出站模板
- 用 Monaco 编辑原始 outbound JSON
- 配置 `visibleDevices`

特点：

- `visibleDevices` 留空表示所有设备可见
- 多个设备 code 使用逗号分隔

## DNS 设置

文件：

- [DnsManage.tsx](/Users/xsdhy/data/code/go/singboxconfig/web/src/pages/DnsManage.tsx)

功能：

- 单独编辑 `GlobalSettings["dns_config"]`
- 提供默认 DNS 模板
- 保存前校验 JSON

特点：

- 恢复默认只会覆盖编辑器内容，不会立即提交后端
- 页面保存时会优先 `PUT`，失败后再尝试 `POST`

## 页面间的协作关系

生成链路相关页面的典型依赖关系如下：

- 订阅管理 -> 节点来源
- 节点分组 -> 出站组构建
- 规则集管理 -> 路由分流
- Inbound 管理 + 设备管理 -> `inbounds`
- WireGuard 管理 + 设备管理 -> `endpoints`
- 额外出站 -> `outbounds`
- DNS 设置 -> `dns`

## 当前缺失页面

按现有仓库实现，管理台里还没有：

- 独立的“配置预览”页面
- “测试订阅可用性”页面
- “默认数据初始化”页面
- 用户登录页

## 相关文档

- [前端架构](./architecture.md)
- [API 客户端](./api-client.md)
- [API 接口列表](../reference/api-reference.md)
