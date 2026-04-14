# 规则集管理

## 模块职责

规则集管理负责维护 sing-box `route.rule_set` 和基于规则集的路由规则。它把“规则内容”与“命中后走哪个出站”绑定在一起。

代码入口：

- 实体定义：`entity/rule_set.go`
- 管理接口：`service/service.go`
- 生成转换：`convert/singbox/route.go`

## 数据模型

`entity.RuleSet` 字段：

- `name`：后台展示名称
- `tag`：规则集唯一标识
- `ruleSetType`：通常为 `remote` 或 `local`
- `format`：规则集格式，远程常用 `binary`，本地常用 `source`
- `content`：本地规则集 JSON 文本
- `url`：远程规则集地址
- `outbound`：命中该规则集后的默认出站
- `downloadDetour`：远程规则集下载时使用的出站
- `ableDevices`：逗号分隔设备编码，用来限制可见设备
- `sort`：输出顺序，越小越靠前

## 管理接口

接口路径为 `/api/rule-sets`：

- `POST /api/rule-sets`
- `PUT /api/rule-sets/:tag`
- `DELETE /api/rule-sets/:tag`
- `GET /api/rule-sets`
- `GET /open/ruleset/:tag`

说明：

- 创建时会检查 `tag` 唯一性
- 更新时会把 `content` 中的换行和空格去掉，便于以紧凑 JSON 持久化
- `GET /open/ruleset/:tag` 会把 `content` 反序列化后直接输出，主要面向本地规则集内容查看

## 生成到 sing-box 的方式

`convert/singbox.GetRoute()` 会把规则集拆成两层结构：

1. `route.rule_set`
2. `route.rules` 中按规则集引用生成的路由规则

### `route.rule_set`

由 `baseRuleSets()` 生成：

- `ruleSetType == "remote"`：输出 `type=remote`、`tag`、`format`、`url`、`download_detour`
- 其他情况：尝试把 `content` 解析成 JSON，成功后输出 `type=inline`

如果本地 `content` 不是合法 JSON，该规则集会被静默跳过，不会阻断整个配置生成。

### `route.rules`

由 `GetRoute()` 生成：

- 先写入一组基础规则
- 再按 `sort` 升序遍历规则集
- 每条规则集追加一条 `{"rule_set":[tag], "outbound": outbound}`

## 设备可见性控制

`ableDevices` 用于限制哪些设备会加载该规则集，当前逻辑是：

- 为空：所有设备可见
- 非空：使用 `strings.Contains(ableDevices, deviceCode)` 判断

这里是简单子串匹配，不是严格的“逗号分隔精确匹配”。例如 `ableDevices="pad,phone"` 时，`deviceCode="ph"` 也可能误命中。这是当前实现的一个已知边界。

## 基础路由规则

除规则集派生规则外，系统还会固定输出几条基础规则：

- `protocol=dns -> action=hijack-dns`
- `clash_mode=direct -> outbound=direct`
- `clash_mode=global -> outbound=select`
- `protocol=quic -> action=reject`

同时路由根对象默认值为：

- `final: "general"`
- `find_process: false`
- `auto_detect_interface: true`

这里的 `final="general"` 假设系统中存在名为 `general` 的出站或节点分组；代码当前不做存在性校验。

## 与其他模块的关系

- 常把 `outbound` 指向“节点分组”的 `tag`
- `downloadDetour` 也通常引用某个已存在出站
- DNS 默认规则中会引用 `cnip`、`cnsite` 两个规则集 tag，因此常见部署会创建这两个规则集

## 当前限制

- 不支持规则集引用合法性校验
- `ableDevices` 采用子串匹配，存在误匹配空间
- 本地规则集保存为字符串，不做结构化字段校验
- `GET /open/ruleset/:tag` 仅适合本地规则集；远程规则集没有单独下载代理接口

## 适合更新本文档的场景

- 调整 `ableDevices` 匹配逻辑
- 新增规则集类型或格式
- 修改基础路由规则
- 引入规则预校验、依赖检查或更严格的引用约束
