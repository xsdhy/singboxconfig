# 配置导入导出

## 模块职责

配置导入导出模块负责整库配置的备份、迁移和恢复。它把多个业务资源收拢成统一 JSON，并提供“尽量导入、避免覆盖”的导入策略。

代码入口：

- 传输结构：`transfer/config_transfer.go`
- 服务实现：`service/config_transfer.go`
- 路由注册：`cmd/server/main.go`

## 接口

当前已注册接口：

- `GET /api/config-transfer/export`
- `POST /api/config-transfer/import`

## 导出数据范围

`buildConfigTransferData()` 会导出以下资源：

- `subscribes`
- `node_groups`
- `rule_sets`
- `global_settings`
- `devices`
- `inbounds`
- `device_inbounds`
- `wire_guards`
- `wire_guard_peers`
- `extra_outbounds`

其中：

- 主资源多以 `map[key]item` 输出，便于按主键查找
- 关联资源如 `device_inbounds`、`wire_guard_peers` 使用数组输出
- `auth.*` 这组保留全局设置不会进入导出结果

## 导出文件行为

导出时会：

1. 收集当前存储里的全部资源
2. 生成格式化 JSON
3. 以附件方式返回
4. 文件名格式为 `singboxconfig-export-YYYYMMDD-HHMMSS.json`

## 导入策略

`importConfigTransferData()` 采用“分类处理”的策略：

### 主资源

如订阅、节点分组、规则集、设备、Inbound、WireGuard、额外出站：

- 已存在：`skipped`
- 不存在：尝试创建
- 创建失败：`failed`

### 全局设置

全局设置按 key 直接覆盖：

- 不区分已存在还是不存在
- 成功写入即记为 `imported`

但有一个例外：

- `auth.*` 属于认证保留配置
- 导入时会直接跳过，并记入 `skipped`

### 设备 Inbound 绑定

按设备分组处理：

- 若某设备已经存在绑定关系，则该设备整组绑定记为 `skipped`
- 若该设备当前没有绑定关系，则整组写入

### WireGuard Peers

按 `wireGuardTag` 分组处理：

- 若某模板下已经存在 peer，则该模板整组 peer 记为 `skipped`
- 若该模板当前没有 peer，则逐条创建

这是一种“避免把新旧关联关系混在一起”的保守导入策略。

## 导入返回摘要

返回体 `ConfigImportSummary` 会按资源类别统计：

- `imported`
- `skipped`
- `failed`
- `errors`

前端会把这些统计渲染为导入结果摘要。

## 文件校验

上传时有两层基础校验：

- multipart 字段必须是 `file`
- 若文件名带扩展名，则必须是 `.json`

之后会直接用 JSON Decoder 解码为 `transfer.ConfigTransferData`。

## 当前限制

- 不支持版本号或 schema migration
- 不支持覆盖导入、合并导入、预演导入等模式切换
- 不做跨资源引用完整性校验
- 导入是逐类、逐项处理，不提供事务级全局回滚
- 不支持通过普通配置导入导出迁移管理员认证信息

## 适合更新本文档的场景

- 调整导入冲突策略
- 给导出 JSON 增加版本字段
- 引入更严格的引用校验或事务语义
