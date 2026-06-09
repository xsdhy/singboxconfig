package service

import (
	"errors"
	"net/http"
	"singboxconfig/convert/shadowrocket"
	"singboxconfig/convert/singbox"
	"singboxconfig/convert/surge"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"sort"

	"github.com/gin-gonic/gin"
)

const dnsConfigSettingKey = "dns_config"

// Generated 按设备输出 sing-box 配置。
// 阶段四开始，这里不再依赖硬编码设备分支，而是优先从存储读取配置，并在空存储时回退到默认数据。
func (s *Service) Generated(c *gin.Context) {
	deviceCode := c.Param("device")
	token := c.Query("token")

	device, err := s.resolveGenerateDevice(deviceCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !device.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Device disabled"})
		return
	}
	if token != device.Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	dnsConfigJSON, err := s.resolveDNSConfigJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	inbounds, err := s.resolveGenerateInbounds(device.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endpoints, err := s.resolveGenerateEndpoints(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	outbounds, err := s.resolveGenerateOutbounds(c.Request.Context(), device.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	singBoxConfig := entity.SingBoxConfig{
		DNS:          singbox.ResolveDNS(dnsConfigJSON),
		Endpoints:    endpoints,
		Route:        singbox.GetRoute(device.Code, ruleSets),
		Experimental: singbox.GetExperimental(device.Code),
		Inbounds:     singbox.GetInbounds(inbounds),
		Outbounds:    outbounds,
	}

	c.JSON(http.StatusOK, singBoxConfig)
}

// SurgeGenerated 按设备输出 Surge 配置文本。
// 设备解析、启用状态、token 鉴权、订阅缓存刷新与 Outbound 可见性过滤均复用 sing-box 生成链路，
// 仅在最后一步改为调用 Surge 渲染器，保证两套输出共享同一份数据治理能力。
func (s *Service) SurgeGenerated(c *gin.Context) {
	deviceCode := c.Param("device")
	token := c.Query("token")

	device, err := s.resolveGenerateDevice(deviceCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !device.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Device disabled"})
		return
	}
	if token != device.Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	dnsConfigJSON, err := s.resolveDNSConfigJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groupRules, err := s.storage.ListNodeGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	outbounds, err := s.resolveGenerateOutbounds(c.Request.Context(), device.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	configText := surge.Render(device.Code, singbox.ResolveDNS(dnsConfigJSON), outbounds, groupRules, ruleSets)
	c.String(http.StatusOK, configText)
}

// ShadowrocketGenerated 按设备输出 Shadowrocket 配置文本。
// 这里与 Surge 输出一样只负责 HTTP 编排：设备解析、启用状态、token 鉴权、DNS、规则集、
// 节点分组和统一 Outbound 均复用现有生成链路，协议能力差异全部留给 Shadowrocket 渲染器处理。
func (s *Service) ShadowrocketGenerated(c *gin.Context) {
	deviceCode := c.Param("device")
	token := c.Query("token")

	device, err := s.resolveGenerateDevice(deviceCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !device.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Device disabled"})
		return
	}
	if token != device.Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	dnsConfigJSON, err := s.resolveDNSConfigJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	groupRules, err := s.storage.ListNodeGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	outbounds, err := s.resolveGenerateOutbounds(c.Request.Context(), device.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	configText := shadowrocket.Render(device.Code, singbox.ResolveDNS(dnsConfigJSON), outbounds, groupRules, ruleSets)
	c.String(http.StatusOK, configText)
}

// resolveGenerateDevice 负责解析生成目标设备。
// 只有在存储完全为空时，才回退到内置默认设备，避免和管理员维护的数据混用。
func (s *Service) resolveGenerateDevice(deviceCode string) (*entity.Device, error) {
	devices, err := s.storage.ListDevices()
	if err != nil {
		return nil, err
	}

	// 仅当存储完全为空时才回退到内置默认设备
	if len(devices) == 0 {
		for _, device := range singbox.GetDefaultDevices() {
			if device != nil && device.Code == deviceCode {
				return device, nil
			}
		}
		return nil, storage.ErrNotFound
	}

	return s.storage.GetDevice(deviceCode)
}

// resolveDNSConfigJSON 读取后台保存的 DNS JSON 配置。
// 未配置时返回空字符串，让下游继续走默认 DNS 逻辑。
func (s *Service) resolveDNSConfigJSON() (string, error) {
	value, err := s.storage.GetGlobalSetting(dnsConfigSettingKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// resolveGenerateInbounds 根据设备绑定关系筛选并排序入站模板。
// 绑定表和入站表都支持在空存储时回退到默认值。
func (s *Service) resolveGenerateInbounds(deviceCode string) ([]*entity.Inbound, error) {
	bindings, err := s.storage.ListDeviceInbounds(deviceCode)
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		bindings = singbox.GetDefaultDeviceInbounds(deviceCode)
	}

	inbounds, err := s.storage.ListInbounds()
	if err != nil {
		return nil, err
	}
	if len(inbounds) == 0 {
		inbounds = singbox.GetDefaultInbounds()
	}

	inboundMap := make(map[string]*entity.Inbound, len(inbounds))
	for _, inbound := range inbounds {
		if inbound == nil || inbound.Tag == "" {
			continue
		}
		inboundMap[inbound.Tag] = inbound
	}

	sortedBindings := append([]*entity.DeviceInbound(nil), bindings...)
	sort.Slice(sortedBindings, func(i, j int) bool {
		if sortedBindings[i].Sort == sortedBindings[j].Sort {
			return sortedBindings[i].InboundTag < sortedBindings[j].InboundTag
		}
		return sortedBindings[i].Sort < sortedBindings[j].Sort
	})

	result := make([]*entity.Inbound, 0, len(sortedBindings))
	for _, binding := range sortedBindings {
		if binding == nil {
			continue
		}
		if inbound, ok := inboundMap[binding.InboundTag]; ok {
			result = append(result, inbound)
		}
	}
	return result, nil
}

// resolveGenerateEndpoints 在设备绑定了 WireGuard 模板时，生成对应 endpoint。
func (s *Service) resolveGenerateEndpoints(device *entity.Device) ([]entity.SingEndpointWireguard, error) {
	if device == nil || device.WireGuardTag == "" {
		return nil, nil
	}

	wg, err := s.storage.GetWireGuard(device.WireGuardTag)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		wg = singbox.GetDefaultWireGuard(device.WireGuardTag)
	}

	peers, err := s.storage.ListWireGuardPeers(device.WireGuardTag)
	if err != nil {
		return nil, err
	}
	if len(peers) == 0 {
		peers = singbox.GetDefaultWireGuardPeers(device.WireGuardTag)
	}

	return singbox.GetWireGuardEndpoints(wg, peers, device), nil
}
