package entity

// OutboundProtocol 表示可被生成链路识别的出站协议类型。
// 这些值与 sing-box outbound.type 保持一致，避免在转换层散落裸字符串。
type OutboundProtocol string

const (
	// OutboundProtocolShadowsocks 表示 Shadowsocks 出站，Surge 中映射为 ss。
	OutboundProtocolShadowsocks OutboundProtocol = "shadowsocks"
	// OutboundProtocolShadowsocksR 表示 ShadowsocksR 出站，Shadowrocket 可完整映射其协议与混淆参数。
	OutboundProtocolShadowsocksR OutboundProtocol = "shadowsocksr"
	// OutboundProtocolTrojan 表示 Trojan 出站，Surge 可完整映射其基础连接与 TLS 参数。
	OutboundProtocolTrojan OutboundProtocol = "trojan"
	// OutboundProtocolVMess 表示 VMess 出站，Surge 仅支持 best-effort 映射。
	OutboundProtocolVMess OutboundProtocol = "vmess"
	// OutboundProtocolHTTP 表示 HTTP/HTTPS 出站，Surge 中映射为 http/https。
	OutboundProtocolHTTP OutboundProtocol = "http"
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
	// OutboundProtocolSocks 表示 SOCKS 出站，sing-box 中不带 version 时默认为 SOCKS5，Surge 中映射为 socks5。
	OutboundProtocolSocks OutboundProtocol = "socks"
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

// Software 表示规则集 open 接口与生成链路支持的目标客户端软件。
// 这些取值会同时出现在路由参数 :software、规则集 URL 路径段与各转换器分支中，
// 因此统一定义为枚举，避免在不同层散落裸字符串。
type Software string

const (
	// SoftwareSingbox 表示 sing-box 客户端，规则集输出为 source 格式 JSON。
	SoftwareSingbox Software = "singbox"
	// SoftwareSurge 表示 Surge 客户端，规则集输出为逐行 `类型,值` 文本。
	SoftwareSurge Software = "surge"
	// SoftwareShadowrocket 表示 Shadowrocket 客户端，规则集输出为逐行 `类型,值` 文本。
	SoftwareShadowrocket Software = "shadowrocket"
)

// ParseSoftware 把请求参数解析为受支持的 Software 枚举。
// 第二个返回值为 false 表示传入的是未知软件名，调用方应据此返回 400。
func ParseSoftware(raw string) (Software, bool) {
	switch Software(raw) {
	case SoftwareSingbox:
		return SoftwareSingbox, true
	case SoftwareSurge:
		return SoftwareSurge, true
	case SoftwareShadowrocket:
		return SoftwareShadowrocket, true
	default:
		return "", false
	}
}
