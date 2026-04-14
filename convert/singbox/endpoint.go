package singbox

import (
	"singboxconfig/entity"
	"sort"
	"strings"
)

// GetWireGuardEndpoints 按模板、Peer 明细和设备侧参数拼出最终的 WireGuard endpoint。
// 这是阶段四里从“按设备硬编码”切到“模板 + 设备参数”的核心逻辑。
func GetWireGuardEndpoints(wg *entity.WireGuard, peers []*entity.WireGuardPeer, device *entity.Device) []entity.SingEndpointWireguard {
	if wg == nil || device == nil {
		return nil
	}
	if device.WireGuardTag == "" || wg.Tag != device.WireGuardTag || !wg.Enabled {
		return nil
	}

	clientAddr := normalizeWireGuardClientAddr(device.WireGuardClientAddr)
	if clientAddr == "" || strings.TrimSpace(device.WireGuardClientKey) == "" {
		return nil
	}

	sortedPeers := append([]*entity.WireGuardPeer(nil), peers...)
	sort.Slice(sortedPeers, func(i, j int) bool {
		if sortedPeers[i].Sort == sortedPeers[j].Sort {
			return sortedPeers[i].ID < sortedPeers[j].ID
		}
		return sortedPeers[i].Sort < sortedPeers[j].Sort
	})

	endpointPeers := make([]entity.SingEndpointWireguardPeer, 0, len(sortedPeers))
	for _, peer := range sortedPeers {
		if peer == nil || !peer.Enabled {
			continue
		}
		endpointPeers = append(endpointPeers, entity.SingEndpointWireguardPeer{
			Address:                     peer.Address,
			Port:                        peer.Port,
			PublicKey:                   peer.PublicKey,
			PreSharedKey:                peer.PreSharedKey,
			AllowedIps:                  splitAllowedIPs(peer.AllowedIPs),
			PersistentKeepaliveInterval: peer.PersistentKeepaliveInterval,
		})
	}

	return []entity.SingEndpointWireguard{
		{
			Type:       "wireguard",
			Tag:        wg.EndpointTag,
			MTU:        wg.MTU,
			Address:    []string{clientAddr},
			PrivateKey: device.WireGuardClientKey,
			Peers:      endpointPeers,
		},
	}
}

func normalizeWireGuardClientAddr(addr string) string {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "/") {
		return trimmed
	}
	return trimmed + "/32"
}

func splitAllowedIPs(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
