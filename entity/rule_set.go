package entity

// RuleSet 描述一个规则集及其对应的默认出站策略。
// 它既支持本地内联规则，也支持 sing-box remote 规则集。
type RuleSet struct {
	Name string `json:"name"` // 管理端展示名称。
	Tag  string `json:"tag"`  // 规则集唯一标识，对应 sing-box rule_set.tag。
	// RuleSetType 取值通常为 remote 或 local。
	RuleSetType string `json:"ruleSetType"`
	// Format 对应 sing-box 规则集格式，remote 常用 binary，local 常用 source。
	Format string `json:"format"`
	// Content 存放 local 规则集的 JSON 文本。
	Content string `json:"content"`
	// URL 是 remote 规则集的下载地址。
	URL string `json:"url"`
	// Outbound 指定命中该规则集后的默认出站，值通常是节点分组 tag。
	Outbound string `json:"outbound"`
	// DownloadDetour 指定 remote 规则集下载时走哪个出站。
	DownloadDetour string `json:"downloadDetour"`
	// AbleDevices 用逗号分隔设备编码，用来控制规则是否对某设备生效。
	AbleDevices string `json:"ableDevices"`
	// Sort 控制路由规则输出顺序，值越小越靠前。
	Sort int `json:"sort"`
}
