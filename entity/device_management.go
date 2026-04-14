package entity

import "time"

// Device 描述一个最终配置的消费终端。
// 每台设备都拥有独立 token、可选的 WireGuard 客户端参数以及 Inbound 绑定关系。
type Device struct {
	Code        string `json:"code"` // 设备唯一编码，也是生成接口里的 device 参数。
	Name        string `json:"name"` // 管理端展示名称。
	Description string `json:"description"`
	Token       string `json:"token"`   // 生成配置时用于鉴权。
	Enabled     bool   `json:"enabled"` // 禁用后无法继续获取配置。
	Sort        int    `json:"sort"`    // 后台列表排序值。
	// WireGuardTag 指向一份 WireGuard 模板，生成时会扩展为 endpoint。
	WireGuardTag string `json:"wireGuardTag"`
	// WireGuardClientAddr 是当前设备在 WireGuard 网络中的地址列表。
	WireGuardClientAddr string `json:"wireGuardClientAddr"`
	// WireGuardClientKey 是当前设备对应的私钥。
	WireGuardClientKey string `json:"wireGuardClientKey"`
}

// Inbound 描述一份可复用的 sing-box 入站模板。
// 具体 JSON 内容存放在 ConfigJSON，生成时会按设备绑定关系注入。
type Inbound struct {
	Tag         string `json:"tag"` // 入站唯一标识，也是设备绑定时使用的键。
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // 如 mixed、tun、direct 等 sing-box inbound 类型。
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
	ConfigJSON  string `json:"configJson"` // 原始 sing-box inbound JSON 配置。
}

// DeviceInbound 表示设备与入站模板的多对多绑定关系。
type DeviceInbound struct {
	DeviceCode string `json:"deviceCode"` // 关联设备编码。
	InboundTag string `json:"inboundTag"` // 关联入站模板 tag。
	Sort       int    `json:"sort"`       // 单设备下多个入站的输出顺序。
}

// WireGuard 描述一份可复用的 WireGuard endpoint 模板。
type WireGuard struct {
	Tag         string `json:"tag"` // 模板唯一标识，设备通过它绑定。
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
	EndpointTag string `json:"endpointTag"` // 输出到 sing-box endpoint.tag。
	MTU         int    `json:"mtu"`
}

// WireGuardPeer 表示某个 WireGuard 模板下的一条对端配置。
type WireGuardPeer struct {
	ID                          int64  `json:"id"`           // 数据库主键。
	WireGuardTag                string `json:"wireGuardTag"` // 所属 WireGuard 模板。
	Sort                        int    `json:"sort"`
	Address                     string `json:"address"`      // 对端地址或域名。
	Port                        int    `json:"port"`         // 对端监听端口。
	PublicKey                   string `json:"publicKey"`    // 对端公钥。
	PreSharedKey                string `json:"preSharedKey"` // 预共享密钥，可为空。
	AllowedIPs                  string `json:"allowedIps"`   // 逗号分隔 CIDR 列表。
	PersistentKeepaliveInterval int    `json:"persistentKeepaliveInterval"`
	Enabled                     bool   `json:"enabled"`
}

// OutboundSource 表示 Outbound 的来源。
type OutboundSource string

const (
	// OutboundSourceManual 表示用户手工维护的 Outbound。
	OutboundSourceManual OutboundSource = "MANUAL"
	// OutboundSourceSubscription 表示从订阅源缓存下来的 Outbound。
	OutboundSourceSubscription OutboundSource = "SUBSCRIPTION"
)

// Outbound 描述系统中统一管理的一条 sing-box 出站配置。
// 手动配置与订阅缓存都落在同一实体，通过 Source 区分来源。
type Outbound struct {
	ID          int64  `json:"id"` // 数据库主键。
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
	// VisibleDevices 用逗号分隔设备编码，控制哪些设备能看到这条出站。
	VisibleDevices string `json:"visibleDevices"`
	// ConfigJSON 是完整的 sing-box outbound JSON。
	ConfigJSON string `json:"configJson"`
	// Source 标识当前记录来自手动维护还是订阅缓存。
	Source OutboundSource `json:"source"`
	// SubscribeName 仅在订阅来源记录中有值，用于标识所属订阅。
	SubscribeName string `json:"subscribeName"`
	// LastFetchTime 记录订阅来源记录最近一次写入缓存的时间。
	LastFetchTime *time.Time `json:"lastFetchTime"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}
