package entity

import "time"

// Subscribe 表示一个远程节点订阅源。
type Subscribe struct {
	Name string `json:"name"` // 订阅名称，兼作后台管理主键。
	URL  string `json:"url"`  // 订阅拉取地址。
	// UserAgent 为空时，生成流程会回退到默认桌面浏览器 UA。
	UserAgent string `json:"userAgent"`
	// Status 为 true 时，该订阅才会参与节点抓取和配置生成。
	Status bool `json:"status"`
	// VisibleDevices 用逗号分隔设备编码，控制哪些设备可使用当前订阅。
	VisibleDevices string `json:"visibleDevices"`
	// OutboundLastFetchTime 记录最近一次成功刷新订阅 Outbound 缓存的时间。
	OutboundLastFetchTime *time.Time `json:"outboundLastFetchTime"`
	// OutboundCacheDuration 表示 Outbound 缓存时长，单位为分钟，0 表示不缓存。
	OutboundCacheDuration int `json:"outboundCacheDuration"`
	// OutboundLastFetchStatus 记录最近一次刷新状态，例如 SUCCESS 或 FAILED。
	OutboundLastFetchStatus string `json:"outboundLastFetchStatus"`
	// OutboundLastFetchError 记录最近一次刷新失败原因，成功时通常为空。
	OutboundLastFetchError string `json:"outboundLastFetchError"`
}
