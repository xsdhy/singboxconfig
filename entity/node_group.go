package entity

// NodeGroup 描述一个 sing-box 节点分组。
// 生成配置时会根据 include / exclude 关键字从订阅节点里筛选成员，
// 再按 groupType 组装为 selector 或 urltest 出站。
type NodeGroup struct {
	Name string `json:"name"` // 管理端展示名称。
	Tag  string `json:"tag"`  // 配置内唯一标识，其他模块按它引用。
	// GroupType 目前主要使用 selector / urltest 两种分组类型。
	GroupType string `json:"groupType"`
	// TestURL 仅在 urltest 分组下使用，作为节点可用性探测地址。
	TestURL string `json:"testURL"`
	// Include 是逗号分隔的包含关键字，只要标签命中任一关键字就会入组。
	Include string `json:"include"`
	// Exclude 是逗号分隔的排除关键字，命中后会从候选节点里剔除。
	Exclude string `json:"exclude"`
}
