package singbox

import (
	"encoding/json"
	"singboxconfig/entity"
)

const (
	// DefaultGenerateToken 开放生成接口的默认令牌
	DefaultGenerateToken = "996007"
)

// GetDefaultDevices 返回空存储下用于兼容旧生成逻辑的内置设备列表。
func GetDefaultDevices() []*entity.Device {
	return []*entity.Device{
		{
			Code:    "default",
			Name:    "default",
			Token:   DefaultGenerateToken,
			Enabled: true,
			Sort:    0,
		},
		{
			Code:                "phone",
			Name:                "phone",
			Token:               DefaultGenerateToken,
			Enabled:             true,
			Sort:                10,
			WireGuardTag:        "",
			WireGuardClientAddr: "",
			WireGuardClientKey:  "",
		},
	}
}

// GetDefaultInbounds 返回空存储下的内置 Inbound 模板。
func GetDefaultInbounds() []*entity.Inbound {
	return []*entity.Inbound{
		{
			Tag:        "tun-default",
			Name:       "TUN Default",
			Type:       "tun",
			Enabled:    true,
			Sort:       0,
			ConfigJSON: mustJSON(entity.SingInbound{Type: "tun", Address: []string{"198.18.0.1/16"}, AutoRoute: true, Stack: "mixed", Sniff: true}),
		},
		{
			Tag:        "http-default",
			Name:       "HTTP Default",
			Type:       "http",
			Enabled:    true,
			Sort:       10,
			ConfigJSON: mustJSON(entity.SingInbound{Type: "http", Tag: "http-in", Listen: "::", ListenPort: 1082}),
		},
		{
			Tag:        "socks-default",
			Name:       "SOCKS Default",
			Type:       "socks",
			Enabled:    true,
			Sort:       20,
			ConfigJSON: mustJSON(entity.SingInbound{Type: "socks", Tag: "socks-in", Listen: "::", ListenPort: 1084}),
		},
		{
			Tag:        "mixed-default",
			Name:       "Mixed Default",
			Type:       "mixed",
			Enabled:    true,
			Sort:       30,
			ConfigJSON: mustJSON(entity.SingInbound{Type: "mixed", Tag: "mixed-in", Listen: "::", ListenPort: 5353}),
		},
	}
}

func GetDefaultDeviceInbounds(deviceCode string) []*entity.DeviceInbound {
	if deviceCode == "phone" {
		return []*entity.DeviceInbound{
			{DeviceCode: deviceCode, InboundTag: "tun-default", Sort: 0},
		}
	}

	return []*entity.DeviceInbound{
		{DeviceCode: deviceCode, InboundTag: "tun-default", Sort: 0},
		{DeviceCode: deviceCode, InboundTag: "http-default", Sort: 10},
		{DeviceCode: deviceCode, InboundTag: "socks-default", Sort: 20},
		{DeviceCode: deviceCode, InboundTag: "mixed-default", Sort: 30},
	}
}

func GetDefaultExtraOutbounds() []*entity.Outbound {
	return []*entity.Outbound{}
}

// GetDefaultWireGuard 根据历史硬编码返回默认模板。
// 只有旧逻辑中实际启用过 WireGuard 的默认设备才会通过该模板继续生成。
func GetDefaultWireGuard(tag string) *entity.WireGuard {
	return nil
}

// GetDefaultWireGuardPeers 返回历史硬编码的默认 peer。
func GetDefaultWireGuardPeers(tag string) []*entity.WireGuardPeer {
	return nil
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}
