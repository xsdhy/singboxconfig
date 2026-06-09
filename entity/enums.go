package entity

// OutboundProtocol 表示可被生成链路识别的出站协议类型。
// 这些值与 sing-box outbound.type 保持一致，避免在转换层散落裸字符串。
type OutboundProtocol string

const (
	// OutboundProtocolShadowsocks 表示 Shadowsocks 出站，Surge 中映射为 ss。
	OutboundProtocolShadowsocks OutboundProtocol = "shadowsocks"
	// OutboundProtocolTrojan 表示 Trojan 出站，Surge 可完整映射其基础连接与 TLS 参数。
	OutboundProtocolTrojan OutboundProtocol = "trojan"
	// OutboundProtocolVMess 表示 VMess 出站，Surge 仅支持 best-effort 映射。
	OutboundProtocolVMess OutboundProtocol = "vmess"
	// OutboundProtocolVLESS 表示 VLESS 出站，Surge 第一版不导出。
	OutboundProtocolVLESS OutboundProtocol = "vless"
	// OutboundProtocolHysteria 表示 Hysteria 出站，Surge 第一版不导出。
	OutboundProtocolHysteria OutboundProtocol = "hysteria"
	// OutboundProtocolHysteria2 表示 Hysteria2 出站，Surge 第一版不导出。
	OutboundProtocolHysteria2 OutboundProtocol = "hysteria2"
	// OutboundProtocolTUIC 表示 TUIC 出站，Surge 第一版不导出。
	OutboundProtocolTUIC OutboundProtocol = "tuic"
	// OutboundProtocolDirect 表示直连出站，用于 Surge 内置 DIRECT 策略。
	OutboundProtocolDirect OutboundProtocol = "direct"
)

// NodeGroupType 表示节点分组的策略类型。
// 数据库存储仍然使用字符串字段，转换层统一通过这些常量判断。
type NodeGroupType string

const (
	// NodeGroupTypeSelector 表示手动选择策略组。
	NodeGroupTypeSelector NodeGroupType = "selector"
	// NodeGroupTypeSelect 兼容部分导入数据中使用的 Surge 风格 select 名称。
	NodeGroupTypeSelect NodeGroupType = "select"
	// NodeGroupTypeURLTest 表示按连通性测试自动选择策略组。
	NodeGroupTypeURLTest NodeGroupType = "urltest"
)

// RuleSetType 表示规则集来源类型。
// remote 会以远程规则集引用输出，local 和 inline 会尝试展开 Content。
type RuleSetType string

const (
	// RuleSetTypeRemote 表示远程规则集。
	RuleSetTypeRemote RuleSetType = "remote"
	// RuleSetTypeLocal 表示本地规则集。
	RuleSetTypeLocal RuleSetType = "local"
	// RuleSetTypeInline 表示内联规则集。
	RuleSetTypeInline RuleSetType = "inline"
)
