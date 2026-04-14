package storage

import (
	"errors"
	"singboxconfig/entity"
	"time"
)

// ErrNotFound 表示记录不存在
var ErrNotFound = errors.New("record not found")

// SubscribeStorage 定义订阅源的持久化能力。
type SubscribeStorage interface {
	CreateSubscribe(subscribe *entity.Subscribe) error
	GetSubscribe(name string) (*entity.Subscribe, error)
	ListSubscribes() ([]*entity.Subscribe, error)
	UpdateSubscribe(subscribe *entity.Subscribe) error
	DeleteSubscribe(name string) error
}

// NodeGroupStorage 定义节点分组的持久化能力。
type NodeGroupStorage interface {
	CreateNodeGroup(group *entity.NodeGroup) error
	GetNodeGroup(tag string) (*entity.NodeGroup, error)
	ListNodeGroups() ([]*entity.NodeGroup, error)
	UpdateNodeGroup(group *entity.NodeGroup) error
	DeleteNodeGroup(tag string) error
}

// RuleSetStorage 定义规则集的持久化能力。
type RuleSetStorage interface {
	CreateRuleSet(ruleSet *entity.RuleSet) error
	GetRuleSet(tag string) (*entity.RuleSet, error)
	ListRuleSets() ([]*entity.RuleSet, error)
	UpdateRuleSet(ruleSet *entity.RuleSet) error
	DeleteRuleSet(tag string) error
}

// GlobalSettingStorage 管理简单 key/value 形式的全局配置。
type GlobalSettingStorage interface {
	SetGlobalSetting(key, value string) error
	GetGlobalSetting(key string) (string, error)
	ListGlobalSettings() (map[string]string, error)
	DeleteGlobalSetting(key string) error
}

// DeviceStorage 管理设备身份信息。
type DeviceStorage interface {
	CreateDevice(device *entity.Device) error
	GetDevice(code string) (*entity.Device, error)
	ListDevices() ([]*entity.Device, error)
	UpdateDevice(device *entity.Device) error
	DeleteDevice(code string) error
}

// InboundStorage 管理可复用的入站模板。
type InboundStorage interface {
	CreateInbound(inbound *entity.Inbound) error
	GetInbound(tag string) (*entity.Inbound, error)
	ListInbounds() ([]*entity.Inbound, error)
	UpdateInbound(inbound *entity.Inbound) error
	DeleteInbound(tag string) error
}

// DeviceInboundStorage 管理设备与入站模板的绑定关系。
type DeviceInboundStorage interface {
	SetDeviceInbounds(deviceCode string, bindings []*entity.DeviceInbound) error
	ListDeviceInbounds(deviceCode string) ([]*entity.DeviceInbound, error)
}

// WireGuardStorage 管理 WireGuard 模板。
type WireGuardStorage interface {
	CreateWireGuard(item *entity.WireGuard) error
	GetWireGuard(tag string) (*entity.WireGuard, error)
	ListWireGuards() ([]*entity.WireGuard, error)
	UpdateWireGuard(item *entity.WireGuard) error
	DeleteWireGuard(tag string) error
}

// WireGuardPeerStorage 管理 WireGuard 模板下的对端列表。
type WireGuardPeerStorage interface {
	CreateWireGuardPeer(item *entity.WireGuardPeer) error
	ListWireGuardPeers(wireGuardTag string) ([]*entity.WireGuardPeer, error)
	UpdateWireGuardPeer(item *entity.WireGuardPeer) error
	DeleteWireGuardPeer(id int64) error
}

// OutboundFilter 描述 Outbound 列表查询条件。
// 指针字段为 nil 时表示不限制对应条件。
type OutboundFilter struct {
	Source        *entity.OutboundSource
	SubscribeName string
	Enabled       *bool
}

// ExtraOutboundStorage 管理订阅节点之外的附加出站。
// 这组接口保留给现有“额外出站”功能使用，只面向 MANUAL 来源数据。
type ExtraOutboundStorage interface {
	CreateExtraOutbound(outbound *entity.Outbound) error
	GetExtraOutbound(tag string) (*entity.Outbound, error)
	ListExtraOutbounds() ([]*entity.Outbound, error)
	UpdateExtraOutbound(outbound *entity.Outbound) error
	DeleteExtraOutbound(tag string) error
}

// OutboundStorage 管理统一的 Outbound 缓存表。
type OutboundStorage interface {
	GetOutbound(id int64) (*entity.Outbound, error)
	ListOutbounds(filters ...OutboundFilter) ([]*entity.Outbound, error)
	GetOutboundsByDevice(deviceCode string) ([]*entity.Outbound, error)
	CreateOrUpdateOutbounds(items []*entity.Outbound) error
	UpdateOutbound(outbound *entity.Outbound) error
	DeleteOutbound(id int64) error
	DeleteOutboundsBySubscribe(subscribeName string) error
	UpdateOutboundCacheTime(subscribeName string, timestamp time.Time) error
}

// Storage 存储接口，支持多种存储后端
type Storage interface {
	SubscribeStorage
	NodeGroupStorage
	RuleSetStorage
	GlobalSettingStorage
	DeviceStorage
	InboundStorage
	DeviceInboundStorage
	WireGuardStorage
	WireGuardPeerStorage
	ExtraOutboundStorage
	OutboundStorage
}
