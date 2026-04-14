package transfer

import "singboxconfig/entity"

// ConfigTransferData 定义配置导入导出的统一 JSON 结构。
type ConfigTransferData struct {
	Subscribes     map[string]*entity.Subscribe `json:"subscribes"`
	NodeGroups     map[string]*entity.NodeGroup `json:"node_groups"`
	RuleSets       map[string]*entity.RuleSet   `json:"rule_sets"`
	GlobalSettings map[string]string            `json:"global_settings"`
	Devices        map[string]*entity.Device    `json:"devices"`
	Inbounds       map[string]*entity.Inbound   `json:"inbounds"`
	DeviceInbounds []*entity.DeviceInbound      `json:"device_inbounds"`
	WireGuards     map[string]*entity.WireGuard `json:"wire_guards"`
	WireGuardPeers []*entity.WireGuardPeer      `json:"wire_guard_peers"`
	ExtraOutbounds map[string]*entity.Outbound  `json:"extra_outbounds"`
}

// ImportCategorySummary 表示单类资源的导入结果统计。
type ImportCategorySummary struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// ConfigImportSummary 表示整次导入的摘要。
type ConfigImportSummary struct {
	Subscribes     ImportCategorySummary `json:"subscribes"`
	NodeGroups     ImportCategorySummary `json:"node_groups"`
	RuleSets       ImportCategorySummary `json:"rule_sets"`
	GlobalSettings ImportCategorySummary `json:"global_settings"`
	Devices        ImportCategorySummary `json:"devices"`
	Inbounds       ImportCategorySummary `json:"inbounds"`
	DeviceInbounds ImportCategorySummary `json:"device_inbounds"`
	WireGuards     ImportCategorySummary `json:"wire_guards"`
	WireGuardPeers ImportCategorySummary `json:"wire_guard_peers"`
	ExtraOutbounds ImportCategorySummary `json:"extra_outbounds"`
	Errors         []string              `json:"errors"`
}
