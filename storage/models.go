// Package storage 提供存储层的数据库模型定义。
// 本文件定义了所有数据库表对应的 GORM 模型结构体，以及与业务实体层（entity 包）之间的转换方法。
// 每个 DB* 结构体都包含 GORM 标签定义、ToEntity() 和 FromEntity() 方法，用于实现存储层与业务层的解耦。
package storage

import (
	"singboxconfig/entity"
	"time"
)

// DBSubscribe 订阅表数据库模型
// 对应 subscribes 表，用于存储远程节点订阅源的配置信息。
// 每个订阅源包含 URL、状态、设备可见性控制以及 Outbound 缓存相关字段。
type DBSubscribe struct {
	// Name 订阅名称，作为主键唯一标识一个订阅源
	Name string `gorm:"primaryKey;size:255" json:"name"`

	// URL 订阅拉取地址，支持 SS/SSR/Trojan/VMess 等协议的订阅链接
	URL string `gorm:"type:text" json:"url"`

	// UserAgent 请求订阅时使用的 User-Agent，为空时使用默认桌面浏览器 UA
	UserAgent string `gorm:"size:500" json:"userAgent"`

	// Status 订阅启用状态，为 true 时该订阅才会参与节点抓取和配置生成
	Status bool `gorm:"default:false" json:"status"`

	// VisibleDevices 可见设备列表，逗号分隔的设备编码，控制哪些设备可使用当前订阅的节点
	VisibleDevices string `gorm:"type:text" json:"visibleDevices"`

	// OutboundLastFetchTime 最近一次成功刷新订阅 Outbound 缓存的时间戳
	OutboundLastFetchTime *time.Time `json:"outboundLastFetchTime"`

	// OutboundCacheDuration Outbound 缓存时长，单位为分钟，0 表示不启用缓存
	OutboundCacheDuration int `gorm:"default:0" json:"outboundCacheDuration"`

	// OutboundLastFetchStatus 最近一次刷新状态，例如 "SUCCESS" 或 "FAILED"
	OutboundLastFetchStatus string `gorm:"size:20" json:"outboundLastFetchStatus"`

	// OutboundLastFetchError 最近一次刷新失败的错误信息，成功时为空
	OutboundLastFetchError string `gorm:"type:text" json:"outboundLastFetchError"`

	// Outbounds 关联的 Outbound 记录列表（一对多关系），通过 SubscribeName 外键关联
	Outbounds []DBOutbound `gorm:"foreignKey:SubscribeName;references:Name" json:"-"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBSubscribe 对应的数据库表名
func (DBSubscribe) TableName() string {
	return "subscribes"
}

// ToEntity 将数据库模型转换为业务实体
// 用于从存储层向业务层传递数据时的类型转换
func (db *DBSubscribe) ToEntity() *entity.Subscribe {
	return &entity.Subscribe{
		Name:                    db.Name,
		URL:                     db.URL,
		UserAgent:               db.UserAgent,
		Status:                  db.Status,
		VisibleDevices:          db.VisibleDevices,
		OutboundLastFetchTime:   db.OutboundLastFetchTime,
		OutboundCacheDuration:   db.OutboundCacheDuration,
		OutboundLastFetchStatus: db.OutboundLastFetchStatus,
		OutboundLastFetchError:  db.OutboundLastFetchError,
	}
}

// FromEntity 从业务实体创建数据库模型
// 用于从业务层向存储层传递数据时的类型转换
func (db *DBSubscribe) FromEntity(e *entity.Subscribe) {
	db.Name = e.Name
	db.URL = e.URL
	db.UserAgent = e.UserAgent
	db.Status = e.Status
	db.VisibleDevices = e.VisibleDevices
	db.OutboundLastFetchTime = e.OutboundLastFetchTime
	db.OutboundCacheDuration = e.OutboundCacheDuration
	db.OutboundLastFetchStatus = e.OutboundLastFetchStatus
	db.OutboundLastFetchError = e.OutboundLastFetchError
}

// DBNodeGroup 节点分组表数据库模型
// 对应 node_groups 表，用于存储节点分组规则。
// 节点分组通过 include/exclude 关键字从订阅节点中筛选成员，
// 并根据 groupType 生成 selector（手动选择）或 urltest（自动测速）类型的 sing-box 出站组。
type DBNodeGroup struct {
	// Tag 分组唯一标识，作为主键，也是其他模块引用该分组时使用的键
	Tag string `gorm:"primaryKey;size:255" json:"tag"`

	// Name 分组展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// GroupType 分组类型，主要使用 "selector"（手动选择）或 "urltest"（自动测速）
	GroupType string `gorm:"size:50" json:"groupType"`

	// TestURL 节点可用性探测地址，仅在 urltest 类型分组中使用
	TestURL string `gorm:"size:500" json:"testURL"`

	// Include 包含关键字，逗号分隔，节点标签命中任一关键字即会被纳入该分组
	Include string `gorm:"type:text" json:"include"`

	// Exclude 排除关键字，逗号分隔，节点标签命中后会从候选节点中剔除
	Exclude string `gorm:"type:text" json:"exclude"`

	// DeviceTypeOverrides 设备级分组类型覆盖规则，逗号分隔的 "设备编码:分组类型" 列表，
	// 例如 "phone:selector,gateway:urltest"。为空时所有设备都使用 GroupType 默认类型。
	DeviceTypeOverrides string `gorm:"type:text" json:"deviceTypeOverrides"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBNodeGroup 对应的数据库表名
func (DBNodeGroup) TableName() string {
	return "node_groups"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBNodeGroup) ToEntity() *entity.NodeGroup {
	return &entity.NodeGroup{
		Name:                db.Name,
		Tag:                 db.Tag,
		GroupType:           db.GroupType,
		TestURL:             db.TestURL,
		Include:             db.Include,
		Exclude:             db.Exclude,
		DeviceTypeOverrides: db.DeviceTypeOverrides,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBNodeGroup) FromEntity(e *entity.NodeGroup) {
	db.Tag = e.Tag
	db.Name = e.Name
	db.GroupType = e.GroupType
	db.TestURL = e.TestURL
	db.Include = e.Include
	db.Exclude = e.Exclude
	db.DeviceTypeOverrides = e.DeviceTypeOverrides
}

// DBRuleSet 规则集表数据库模型
// 对应 rule_sets 表，用于存储路由规则集配置。
// 规则集既支持本地内联规则（local），也支持远程规则集（remote），
// 用于控制流量分流策略和路由决策。
type DBRuleSet struct {
	// Tag 规则集唯一标识，作为主键，对应 sing-box 配置中的 rule_set.tag
	Tag string `gorm:"primaryKey;size:255" json:"tag"`

	// Name 规则集展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// RuleSetType 规则集类型，通常为 "remote"（远程规则集）或 "local"（本地规则集）
	RuleSetType string `gorm:"size:50" json:"ruleSetType"`

	// Format 规则集格式，对应 sing-box 规则集格式
	// remote 类型常用 "binary"，local 类型常用 "source"
	Format string `gorm:"size:50" json:"format"`

	// Content 本地规则集的 JSON 文本内容，仅在 local 类型规则集中使用
	Content string `gorm:"type:text" json:"content"`

	// URL 远程规则集的下载地址，仅在 remote 类型规则集中使用
	URL string `gorm:"size:500" json:"url"`

	// Outbound 命中该规则集后的默认出站标识，通常是节点分组的 tag
	Outbound string `gorm:"size:255" json:"outbound"`

	// DownloadDetour 远程规则集下载时使用的出站标识，指定通过哪个代理下载规则集
	DownloadDetour string `gorm:"size:255" json:"downloadDetour"`

	// AbleDevices 可用设备列表，逗号分隔的设备编码，控制该规则集对哪些设备生效
	AbleDevices string `gorm:"type:text" json:"ableDevices"`

	// Sort 路由规则输出顺序，值越小越靠前，影响规则匹配的优先级
	Sort int `gorm:"default:0" json:"sort"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBRuleSet 对应的数据库表名
func (DBRuleSet) TableName() string {
	return "rule_sets"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBRuleSet) ToEntity() *entity.RuleSet {
	return &entity.RuleSet{
		Name:           db.Name,
		Tag:            db.Tag,
		RuleSetType:    db.RuleSetType,
		Format:         db.Format,
		Content:        db.Content,
		URL:            db.URL,
		Outbound:       db.Outbound,
		DownloadDetour: db.DownloadDetour,
		AbleDevices:    db.AbleDevices,
		Sort:           db.Sort,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBRuleSet) FromEntity(e *entity.RuleSet) {
	db.Tag = e.Tag
	db.Name = e.Name
	db.RuleSetType = e.RuleSetType
	db.Format = e.Format
	db.Content = e.Content
	db.URL = e.URL
	db.Outbound = e.Outbound
	db.DownloadDetour = e.DownloadDetour
	db.AbleDevices = e.AbleDevices
	db.Sort = e.Sort
}

// DBGlobalSetting 全局设置表数据库模型
// 对应 global_settings 表，用于存储系统级的全局配置项。
// 采用 Key-Value 结构，支持存储 DNS 配置等全局设置的 JSON 文本。
type DBGlobalSetting struct {
	// Key 配置项键名，作为主键唯一标识一个配置项
	Key string `gorm:"primaryKey;size:255" json:"key"`

	// Value 配置项值，通常存储 JSON 格式的配置文本
	Value string `gorm:"type:text" json:"value"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBGlobalSetting 对应的数据库表名
func (DBGlobalSetting) TableName() string {
	return "global_settings"
}

// DBDevice 设备表数据库模型
// 对应 devices 表，用于存储配置消费终端的信息。
// 每台设备拥有独立的 token 用于鉴权，可配置 WireGuard 客户端参数，
// 并通过 DeviceInbound 关联表绑定入站模板。
type DBDevice struct {
	// Code 设备唯一编码，作为主键，也是生成接口中的 device 参数
	Code string `gorm:"primaryKey;size:255" json:"code"`

	// Name 设备展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// Description 设备描述信息
	Description string `gorm:"type:text" json:"description"`

	// Token 设备鉴权令牌，生成配置时用于验证设备身份
	Token string `gorm:"size:255" json:"token"`

	// Enabled 设备启用状态，禁用后无法继续获取配置
	Enabled bool `gorm:"default:true" json:"enabled"`

	// Sort 设备列表排序值，用于管理界面中的显示顺序
	Sort int `gorm:"default:0" json:"sort"`

	// WireGuardTag 关联的 WireGuard 模板标识，生成配置时会扩展为 endpoint
	WireGuardTag string `gorm:"size:255" json:"wireGuardTag"`

	// WireGuardClientAddr 设备在 WireGuard 网络中的地址列表（CIDR 格式）
	WireGuardClientAddr string `gorm:"size:255" json:"wireGuardClientAddr"`

	// WireGuardClientKey 设备对应的 WireGuard 私钥
	WireGuardClientKey string `gorm:"type:text" json:"wireGuardClientKey"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBDevice 对应的数据库表名
func (DBDevice) TableName() string {
	return "devices"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBDevice) ToEntity() *entity.Device {
	return &entity.Device{
		Code:                db.Code,
		Name:                db.Name,
		Description:         db.Description,
		Token:               db.Token,
		Enabled:             db.Enabled,
		Sort:                db.Sort,
		WireGuardTag:        db.WireGuardTag,
		WireGuardClientAddr: db.WireGuardClientAddr,
		WireGuardClientKey:  db.WireGuardClientKey,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBDevice) FromEntity(e *entity.Device) {
	db.Code = e.Code
	db.Name = e.Name
	db.Description = e.Description
	db.Token = e.Token
	db.Enabled = e.Enabled
	db.Sort = e.Sort
	db.WireGuardTag = e.WireGuardTag
	db.WireGuardClientAddr = e.WireGuardClientAddr
	db.WireGuardClientKey = e.WireGuardClientKey
}

// DBInbound 入站表数据库模型
// 对应 inbounds 表，用于存储可复用的 sing-box 入站模板。
// 每个入站模板包含完整的 sing-box inbound JSON 配置，
// 通过 DeviceInbound 关联表与设备建立多对多绑定关系。
type DBInbound struct {
	// Tag 入站唯一标识，作为主键，也是设备绑定时使用的键
	Tag string `gorm:"primaryKey;size:255" json:"tag"`

	// Name 入站展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// Description 入站描述信息
	Description string `gorm:"type:text" json:"description"`

	// Type 入站类型，对应 sing-box inbound 类型，如 "mixed"、"tun"、"direct" 等
	Type string `gorm:"size:50" json:"type"`

	// Enabled 入站启用状态
	Enabled bool `gorm:"default:true" json:"enabled"`

	// Sort 入站列表排序值，用于管理界面中的显示顺序
	Sort int `gorm:"default:0" json:"sort"`

	// ConfigJSON 原始 sing-box inbound JSON 配置文本
	ConfigJSON string `gorm:"type:text" json:"configJson"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBInbound 对应的数据库表名
func (DBInbound) TableName() string {
	return "inbounds"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBInbound) ToEntity() *entity.Inbound {
	return &entity.Inbound{
		Tag:         db.Tag,
		Name:        db.Name,
		Description: db.Description,
		Type:        db.Type,
		Enabled:     db.Enabled,
		Sort:        db.Sort,
		ConfigJSON:  db.ConfigJSON,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBInbound) FromEntity(e *entity.Inbound) {
	db.Tag = e.Tag
	db.Name = e.Name
	db.Description = e.Description
	db.Type = e.Type
	db.Enabled = e.Enabled
	db.Sort = e.Sort
	db.ConfigJSON = e.ConfigJSON
}

// DBDeviceInbound 设备入站关联表数据库模型
// 对应 device_inbounds 表，用于存储设备与入站模板的多对多绑定关系。
// 通过此表可以为每台设备灵活配置不同的入站组合。
type DBDeviceInbound struct {
	// DeviceCode 关联的设备编码，与 devices 表的 code 字段对应，作为联合主键之一
	DeviceCode string `gorm:"primaryKey;size:255" json:"deviceCode"`

	// InboundTag 关联的入站模板标识，与 inbounds 表的 tag 字段对应，作为联合主键之一
	InboundTag string `gorm:"primaryKey;size:255" json:"inboundTag"`

	// Sort 单设备下多个入站的输出顺序，值越小越靠前
	Sort int `gorm:"default:0" json:"sort"`
}

// TableName 指定 DBDeviceInbound 对应的数据库表名
func (DBDeviceInbound) TableName() string {
	return "device_inbounds"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBDeviceInbound) ToEntity() *entity.DeviceInbound {
	return &entity.DeviceInbound{
		DeviceCode: db.DeviceCode,
		InboundTag: db.InboundTag,
		Sort:       db.Sort,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBDeviceInbound) FromEntity(e *entity.DeviceInbound) {
	db.DeviceCode = e.DeviceCode
	db.InboundTag = e.InboundTag
	db.Sort = e.Sort
}

// DBWireGuard WireGuard 表数据库模型
// 对应 wire_guards 表，用于存储可复用的 WireGuard endpoint 模板。
// 每个模板可以包含多个 Peer 配置，设备通过 WireGuardTag 字段绑定模板。
type DBWireGuard struct {
	// Tag 模板唯一标识，作为主键，设备通过此标识绑定 WireGuard 模板
	Tag string `gorm:"primaryKey;size:255" json:"tag"`

	// Name 模板展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// Description 模板描述信息
	Description string `gorm:"type:text" json:"description"`

	// Enabled 模板启用状态
	Enabled bool `gorm:"default:true" json:"enabled"`

	// Sort 模板列表排序值，用于管理界面中的显示顺序
	Sort int `gorm:"default:0" json:"sort"`

	// EndpointTag 输出到 sing-box 配置中的 endpoint.tag 字段值
	EndpointTag string `gorm:"size:255" json:"endpointTag"`

	// MTU WireGuard 接口的 MTU 值，0 表示使用默认值
	MTU int `gorm:"default:0" json:"mtu"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBWireGuard 对应的数据库表名
func (DBWireGuard) TableName() string {
	return "wire_guards"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBWireGuard) ToEntity() *entity.WireGuard {
	return &entity.WireGuard{
		Tag:         db.Tag,
		Name:        db.Name,
		Description: db.Description,
		Enabled:     db.Enabled,
		Sort:        db.Sort,
		EndpointTag: db.EndpointTag,
		MTU:         db.MTU,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBWireGuard) FromEntity(e *entity.WireGuard) {
	db.Tag = e.Tag
	db.Name = e.Name
	db.Description = e.Description
	db.Enabled = e.Enabled
	db.Sort = e.Sort
	db.EndpointTag = e.EndpointTag
	db.MTU = e.MTU
}

// DBWireGuardPeer WireGuard 对等节点表数据库模型
// 对应 wire_guard_peers 表，用于存储 WireGuard 模板下的对端配置。
// 每个 Peer 记录包含对端地址、端口、公钥、预共享密钥等 WireGuard 连接参数。
type DBWireGuardPeer struct {
	// ID 数据库自增主键
	ID int64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// WireGuardTag 所属 WireGuard 模板标识，与 wire_guards 表的 tag 字段关联，建立索引
	WireGuardTag string `gorm:"index;size:255" json:"wireGuardTag"`

	// Sort Peer 列表排序值，用于控制多个 Peer 的输出顺序
	Sort int `gorm:"default:0" json:"sort"`

	// Address 对端地址或域名
	Address string `gorm:"size:255" json:"address"`

	// Port 对端监听端口
	Port int `gorm:"default:0" json:"port"`

	// PublicKey 对端公钥
	PublicKey string `gorm:"type:text" json:"publicKey"`

	// PreSharedKey 预共享密钥（PSK），可为空，用于增强安全性
	PreSharedKey string `gorm:"type:text" json:"preSharedKey"`

	// AllowedIPs 允许的 IP 地址范围，逗号分隔的 CIDR 列表
	AllowedIPs string `gorm:"type:text" json:"allowedIps"`

	// PersistentKeepaliveInterval 持久保活间隔（秒），0 表示不启用
	PersistentKeepaliveInterval int `gorm:"default:0" json:"persistentKeepaliveInterval"`

	// Enabled Peer 启用状态
	Enabled bool `gorm:"default:true" json:"enabled"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBWireGuardPeer 对应的数据库表名
func (DBWireGuardPeer) TableName() string {
	return "wire_guard_peers"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBWireGuardPeer) ToEntity() *entity.WireGuardPeer {
	return &entity.WireGuardPeer{
		ID:                          db.ID,
		WireGuardTag:                db.WireGuardTag,
		Sort:                        db.Sort,
		Address:                     db.Address,
		Port:                        db.Port,
		PublicKey:                   db.PublicKey,
		PreSharedKey:                db.PreSharedKey,
		AllowedIPs:                  db.AllowedIPs,
		PersistentKeepaliveInterval: db.PersistentKeepaliveInterval,
		Enabled:                     db.Enabled,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBWireGuardPeer) FromEntity(e *entity.WireGuardPeer) {
	db.ID = e.ID
	db.WireGuardTag = e.WireGuardTag
	db.Sort = e.Sort
	db.Address = e.Address
	db.Port = e.Port
	db.PublicKey = e.PublicKey
	db.PreSharedKey = e.PreSharedKey
	db.AllowedIPs = e.AllowedIPs
	db.PersistentKeepaliveInterval = e.PersistentKeepaliveInterval
	db.Enabled = e.Enabled
}

// DBOutbound 出站表数据库模型
// 对应 outbounds 表，用于统一管理系统中的所有 sing-box 出站配置。
// 包括用户手动维护的出站（MANUAL）和从订阅源缓存的出站（SUBSCRIPTION）。
// 通过 Source 字段区分来源，实现手动节点与订阅节点的统一管理。
type DBOutbound struct {
	// ID 数据库自增主键
	ID int64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// Tag 出站唯一标识，建立唯一索引，不能为空
	Tag string `gorm:"uniqueIndex;not null;size:255" json:"tag"`

	// Name 出站展示名称，用于管理界面显示
	Name string `gorm:"size:255" json:"name"`

	// Description 出站描述信息
	Description string `gorm:"type:text" json:"description"`

	// Type 出站类型，对应 sing-box outbound 类型，如 "shadowsocks"、"trojan"、"vmess" 等
	Type string `gorm:"size:50" json:"type"`

	// Enabled 出站启用状态，建立索引以优化查询性能
	Enabled bool `gorm:"default:true;index" json:"enabled"`

	// Sort 出站列表排序值，用于管理界面中的显示顺序
	Sort int `gorm:"default:0" json:"sort"`

	// VisibleDevices 可见设备列表，逗号分隔的设备编码，控制哪些设备能看到这条出站
	VisibleDevices string `gorm:"type:text" json:"visibleDevices"`

	// ConfigJSON 完整的 sing-box outbound JSON 配置文本，不能为空
	ConfigJSON string `gorm:"type:text;not null" json:"configJson"`

	// Source 出站来源标识，取值为 "MANUAL"（手动维护）或 "SUBSCRIPTION"（订阅缓存）
	// 使用 varchar(20) 类型，默认值为 MANUAL，建立索引以优化按来源查询
	Source entity.OutboundSource `gorm:"type:varchar(20);default:MANUAL;index" json:"source"`

	// SubscribeName 所属订阅名称，仅在订阅来源记录中有值，建立索引以优化关联查询
	SubscribeName string `gorm:"index;size:255" json:"subscribeName"`

	// LastFetchTime 订阅来源记录最近一次写入缓存的时间戳
	LastFetchTime *time.Time `json:"lastFetchTime"`

	// CreatedAt 记录创建时间，由 GORM 自动管理
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// UpdatedAt 记录更新时间，由 GORM 自动管理
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定 DBOutbound 对应的数据库表名
func (DBOutbound) TableName() string {
	return "outbounds"
}

// ToEntity 将数据库模型转换为业务实体
func (db *DBOutbound) ToEntity() *entity.Outbound {
	return &entity.Outbound{
		ID:             db.ID,
		Tag:            db.Tag,
		Name:           db.Name,
		Description:    db.Description,
		Type:           db.Type,
		Enabled:        db.Enabled,
		Sort:           db.Sort,
		VisibleDevices: db.VisibleDevices,
		ConfigJSON:     db.ConfigJSON,
		Source:         db.Source,
		SubscribeName:  db.SubscribeName,
		LastFetchTime:  db.LastFetchTime,
		CreatedAt:      db.CreatedAt,
		UpdatedAt:      db.UpdatedAt,
	}
}

// FromEntity 从业务实体创建数据库模型
func (db *DBOutbound) FromEntity(e *entity.Outbound) {
	db.ID = e.ID
	db.Tag = e.Tag
	db.Name = e.Name
	db.Description = e.Description
	db.Type = e.Type
	db.Enabled = e.Enabled
	db.Sort = e.Sort
	db.VisibleDevices = e.VisibleDevices
	db.ConfigJSON = e.ConfigJSON
	db.Source = e.Source
	db.SubscribeName = e.SubscribeName
	db.LastFetchTime = e.LastFetchTime
}
