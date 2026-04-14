package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"singboxconfig/transfer"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// buildConfigTransferData 收集当前存储中的全部可导入导出资源。
// 阶段五之后，设备管理相关实体也需要与原有基础配置一起参与传输。
func (s *Service) buildConfigTransferData() (*transfer.ConfigTransferData, error) {
	subscribes, err := s.storage.ListSubscribes()
	if err != nil {
		return nil, err
	}

	nodeGroups, err := s.storage.ListNodeGroups()
	if err != nil {
		return nil, err
	}

	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		return nil, err
	}

	globalSettings, err := s.storage.ListGlobalSettings()
	if err != nil {
		return nil, err
	}

	devices, err := s.storage.ListDevices()
	if err != nil {
		return nil, err
	}

	inbounds, err := s.storage.ListInbounds()
	if err != nil {
		return nil, err
	}

	wireGuards, err := s.storage.ListWireGuards()
	if err != nil {
		return nil, err
	}

	extraOutbounds, err := s.storage.ListExtraOutbounds()
	if err != nil {
		return nil, err
	}

	data := &transfer.ConfigTransferData{
		Subscribes:     make(map[string]*entity.Subscribe, len(subscribes)),
		NodeGroups:     make(map[string]*entity.NodeGroup, len(nodeGroups)),
		RuleSets:       make(map[string]*entity.RuleSet, len(ruleSets)),
		GlobalSettings: make(map[string]string, len(globalSettings)),
		Devices:        make(map[string]*entity.Device, len(devices)),
		Inbounds:       make(map[string]*entity.Inbound, len(inbounds)),
		DeviceInbounds: make([]*entity.DeviceInbound, 0),
		WireGuards:     make(map[string]*entity.WireGuard, len(wireGuards)),
		WireGuardPeers: make([]*entity.WireGuardPeer, 0),
		ExtraOutbounds: make(map[string]*entity.Outbound, len(extraOutbounds)),
	}

	for _, subscribe := range subscribes {
		if subscribe == nil || subscribe.Name == "" {
			continue
		}
		data.Subscribes[subscribe.Name] = subscribe
	}

	for _, group := range nodeGroups {
		if group == nil || group.Tag == "" {
			continue
		}
		data.NodeGroups[group.Tag] = group
	}

	for _, ruleSet := range ruleSets {
		if ruleSet == nil || ruleSet.Tag == "" {
			continue
		}
		data.RuleSets[ruleSet.Tag] = ruleSet
	}

	for key, value := range globalSettings {
		if isReservedGlobalSettingKey(key) {
			continue
		}
		data.GlobalSettings[key] = value
	}

	for _, device := range devices {
		if device == nil || device.Code == "" {
			continue
		}
		data.Devices[device.Code] = device
		bindings, err := s.storage.ListDeviceInbounds(device.Code)
		if err != nil {
			return nil, err
		}
		data.DeviceInbounds = append(data.DeviceInbounds, bindings...)
	}

	for _, inbound := range inbounds {
		if inbound == nil || inbound.Tag == "" {
			continue
		}
		data.Inbounds[inbound.Tag] = inbound
	}

	for _, wg := range wireGuards {
		if wg == nil || wg.Tag == "" {
			continue
		}
		data.WireGuards[wg.Tag] = wg
		peers, err := s.storage.ListWireGuardPeers(wg.Tag)
		if err != nil {
			return nil, err
		}
		data.WireGuardPeers = append(data.WireGuardPeers, peers...)
	}

	for _, outbound := range extraOutbounds {
		if outbound == nil || outbound.Tag == "" {
			continue
		}
		data.ExtraOutbounds[outbound.Tag] = outbound
	}

	return data, nil
}

// importConfigTransferData 按“尽量导入、避免覆盖”的原则写入配置。
// 对已有主资源采用跳过策略；对设置类资源仍按 key 覆盖；对关联类资源按归属对象整体跳过或整体写入。
func (s *Service) importConfigTransferData(data *transfer.ConfigTransferData) transfer.ConfigImportSummary {
	summary := transfer.ConfigImportSummary{
		Errors: make([]string, 0),
	}

	if data == nil {
		return summary
	}

	for _, subscribe := range data.Subscribes {
		if subscribe == nil || subscribe.Name == "" {
			summary.Subscribes.Failed++
			summary.Errors = append(summary.Errors, "subscribe name is required")
			continue
		}

		existing, err := s.storage.GetSubscribe(subscribe.Name)
		if err == nil && existing != nil {
			summary.Subscribes.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.Subscribes.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check subscribe %q failed: %v", subscribe.Name, err))
			continue
		}

		if err := s.storage.CreateSubscribe(subscribe); err != nil {
			summary.Subscribes.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import subscribe %q failed: %v", subscribe.Name, err))
			continue
		}

		summary.Subscribes.Imported++
	}

	for _, group := range data.NodeGroups {
		if group == nil || group.Tag == "" {
			summary.NodeGroups.Failed++
			summary.Errors = append(summary.Errors, "node group tag is required")
			continue
		}

		existing, err := s.storage.GetNodeGroup(group.Tag)
		if err == nil && existing != nil {
			summary.NodeGroups.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.NodeGroups.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check node group %q failed: %v", group.Tag, err))
			continue
		}

		if err := s.storage.CreateNodeGroup(group); err != nil {
			summary.NodeGroups.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import node group %q failed: %v", group.Tag, err))
			continue
		}

		summary.NodeGroups.Imported++
	}

	for _, ruleSet := range data.RuleSets {
		if ruleSet == nil || ruleSet.Tag == "" {
			summary.RuleSets.Failed++
			summary.Errors = append(summary.Errors, "rule set tag is required")
			continue
		}

		existing, err := s.storage.GetRuleSet(ruleSet.Tag)
		if err == nil && existing != nil {
			summary.RuleSets.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.RuleSets.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check rule set %q failed: %v", ruleSet.Tag, err))
			continue
		}

		if err := s.storage.CreateRuleSet(ruleSet); err != nil {
			summary.RuleSets.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import rule set %q failed: %v", ruleSet.Tag, err))
			continue
		}

		summary.RuleSets.Imported++
	}

	for key, value := range data.GlobalSettings {
		if key == "" {
			summary.GlobalSettings.Failed++
			summary.Errors = append(summary.Errors, "global setting key is required")
			continue
		}
		if isReservedGlobalSettingKey(key) {
			summary.GlobalSettings.Skipped++
			continue
		}

		if err := s.storage.SetGlobalSetting(key, value); err != nil {
			summary.GlobalSettings.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import setting %q failed: %v", key, err))
			continue
		}

		summary.GlobalSettings.Imported++
	}

	for _, device := range data.Devices {
		if device == nil || device.Code == "" {
			summary.Devices.Failed++
			summary.Errors = append(summary.Errors, "device code is required")
			continue
		}

		existing, err := s.storage.GetDevice(device.Code)
		if err == nil && existing != nil {
			summary.Devices.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.Devices.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check device %q failed: %v", device.Code, err))
			continue
		}

		if err := s.storage.CreateDevice(device); err != nil {
			summary.Devices.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import device %q failed: %v", device.Code, err))
			continue
		}

		summary.Devices.Imported++
	}

	for _, inbound := range data.Inbounds {
		if inbound == nil || inbound.Tag == "" {
			summary.Inbounds.Failed++
			summary.Errors = append(summary.Errors, "inbound tag is required")
			continue
		}

		existing, err := s.storage.GetInbound(inbound.Tag)
		if err == nil && existing != nil {
			summary.Inbounds.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.Inbounds.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check inbound %q failed: %v", inbound.Tag, err))
			continue
		}

		if err := s.storage.CreateInbound(inbound); err != nil {
			summary.Inbounds.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import inbound %q failed: %v", inbound.Tag, err))
			continue
		}

		summary.Inbounds.Imported++
	}

	for deviceCode, bindings := range groupDeviceInbounds(data.DeviceInbounds) {
		if deviceCode == "" {
			summary.DeviceInbounds.Failed += len(bindings)
			summary.Errors = append(summary.Errors, "device inbound deviceCode is required")
			continue
		}
		if len(bindings) == 0 {
			continue
		}

		existing, err := s.storage.ListDeviceInbounds(deviceCode)
		if err != nil {
			summary.DeviceInbounds.Failed += len(bindings)
			summary.Errors = append(summary.Errors, fmt.Sprintf("check device inbounds %q failed: %v", deviceCode, err))
			continue
		}
		if len(existing) > 0 {
			summary.DeviceInbounds.Skipped += len(bindings)
			continue
		}

		if err := s.storage.SetDeviceInbounds(deviceCode, bindings); err != nil {
			summary.DeviceInbounds.Failed += len(bindings)
			summary.Errors = append(summary.Errors, fmt.Sprintf("import device inbounds %q failed: %v", deviceCode, err))
			continue
		}

		summary.DeviceInbounds.Imported += len(bindings)
	}

	for _, wg := range data.WireGuards {
		if wg == nil || wg.Tag == "" {
			summary.WireGuards.Failed++
			summary.Errors = append(summary.Errors, "wire guard tag is required")
			continue
		}

		existing, err := s.storage.GetWireGuard(wg.Tag)
		if err == nil && existing != nil {
			summary.WireGuards.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.WireGuards.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check wire guard %q failed: %v", wg.Tag, err))
			continue
		}

		if err := s.storage.CreateWireGuard(wg); err != nil {
			summary.WireGuards.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import wire guard %q failed: %v", wg.Tag, err))
			continue
		}

		summary.WireGuards.Imported++
	}

	for wireGuardTag, peers := range groupWireGuardPeers(data.WireGuardPeers) {
		if wireGuardTag == "" {
			summary.WireGuardPeers.Failed += len(peers)
			summary.Errors = append(summary.Errors, "wire guard peer wireGuardTag is required")
			continue
		}
		if len(peers) == 0 {
			continue
		}

		existing, err := s.storage.ListWireGuardPeers(wireGuardTag)
		if err != nil {
			summary.WireGuardPeers.Failed += len(peers)
			summary.Errors = append(summary.Errors, fmt.Sprintf("check wire guard peers %q failed: %v", wireGuardTag, err))
			continue
		}
		if len(existing) > 0 {
			summary.WireGuardPeers.Skipped += len(peers)
			continue
		}

		groupImported := 0
		groupFailed := false
		for _, peer := range peers {
			if err := s.storage.CreateWireGuardPeer(peer); err != nil {
				groupFailed = true
				summary.WireGuardPeers.Failed += len(peers) - groupImported
				summary.Errors = append(summary.Errors, fmt.Sprintf("import wire guard peer %q failed: %v", wireGuardTag, err))
				break
			}
			groupImported++
		}
		if !groupFailed {
			summary.WireGuardPeers.Imported += groupImported
		}
	}

	for _, outbound := range data.ExtraOutbounds {
		if outbound == nil || outbound.Tag == "" {
			summary.ExtraOutbounds.Failed++
			summary.Errors = append(summary.Errors, "extra outbound tag is required")
			continue
		}

		existing, err := s.storage.GetExtraOutbound(outbound.Tag)
		if err == nil && existing != nil {
			summary.ExtraOutbounds.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			summary.ExtraOutbounds.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("check extra outbound %q failed: %v", outbound.Tag, err))
			continue
		}

		if err := s.storage.CreateExtraOutbound(outbound); err != nil {
			summary.ExtraOutbounds.Failed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("import extra outbound %q failed: %v", outbound.Tag, err))
			continue
		}

		summary.ExtraOutbounds.Imported++
	}

	return summary
}

func (s *Service) ExportConfig(c *gin.Context) {
	data, err := s.buildConfigTransferData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("singboxconfig-export-%s.json", time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
}

func (s *Service) ImportConfig(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing upload file"})
		return
	}

	if ext := strings.ToLower(filepath.Ext(fileHeader.Filename)); ext != "" && ext != ".json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file must be a .json file"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	var data transfer.ConfigTransferData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid config json: %v", err)})
		return
	}

	summary := s.importConfigTransferData(&data)
	c.JSON(http.StatusOK, summary)
}

// groupDeviceInbounds 按设备聚合绑定关系，便于按设备整批导入。
func groupDeviceInbounds(items []*entity.DeviceInbound) map[string][]*entity.DeviceInbound {
	result := make(map[string][]*entity.DeviceInbound)
	for _, item := range items {
		if item == nil || item.DeviceCode == "" || item.InboundTag == "" {
			continue
		}
		result[item.DeviceCode] = append(result[item.DeviceCode], item)
	}
	return result
}

// groupWireGuardPeers 按 WireGuard 模板聚合 peer，便于判断“该模板是否已初始化”。
func groupWireGuardPeers(items []*entity.WireGuardPeer) map[string][]*entity.WireGuardPeer {
	result := make(map[string][]*entity.WireGuardPeer)
	for _, item := range items {
		if item == nil || item.WireGuardTag == "" {
			continue
		}
		result[item.WireGuardTag] = append(result[item.WireGuardTag], item)
	}
	return result
}
