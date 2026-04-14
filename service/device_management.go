package service

import (
	"errors"
	"net/http"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

// deleteSuccessResponse 统一删除接口的返回结构，避免重复构造字面量。
var deleteSuccessResponse = gin.H{"message": "Deleted successfully"}

// CreateDevice 创建设备。
func (s *Service) CreateDevice(c *gin.Context) {
	var device entity.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.storage.GetDevice(device.Code); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateDevice(&device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

// GetDevice 获取单个设备详情。
func (s *Service) GetDevice(c *gin.Context) {
	device, err := s.storage.GetDevice(c.Param("code"))
	if err != nil {
		handleStorageNotFound(c, err, "Device not found")
		return
	}
	c.JSON(http.StatusOK, device)
}

// ListDevices 列出所有设备，并按 sort、code 稳定排序，方便前端直接展示。
func (s *Service) ListDevices(c *gin.Context) {
	devices, err := s.storage.ListDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Sort == devices[j].Sort {
			return devices[i].Code < devices[j].Code
		}
		return devices[i].Sort < devices[j].Sort
	})

	c.JSON(http.StatusOK, devices)
}

// UpdateDevice 更新设备。
func (s *Service) UpdateDevice(c *gin.Context) {
	var device entity.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if pathCode := c.Param("code"); pathCode != "" && pathCode != device.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path code and body code do not match"})
		return
	}

	if err := s.storage.UpdateDevice(&device); err != nil {
		handleStorageNotFound(c, err, "Device not found")
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice 删除设备。
func (s *Service) DeleteDevice(c *gin.Context) {
	if err := s.storage.DeleteDevice(c.Param("code")); err != nil {
		handleStorageNotFound(c, err, "Device not found")
		return
	}

	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// CreateInbound 创建 inbound 模板。
func (s *Service) CreateInbound(c *gin.Context) {
	var inbound entity.Inbound
	if err := c.ShouldBindJSON(&inbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.storage.GetInbound(inbound.Tag); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateInbound(&inbound); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inbound)
}

// GetInbound 获取单个 inbound 模板。
func (s *Service) GetInbound(c *gin.Context) {
	inbound, err := s.storage.GetInbound(c.Param("tag"))
	if err != nil {
		handleStorageNotFound(c, err, "Inbound not found")
		return
	}
	c.JSON(http.StatusOK, inbound)
}

// ListInbounds 列出 inbound 模板。
func (s *Service) ListInbounds(c *gin.Context) {
	inbounds, err := s.storage.ListInbounds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(inbounds, func(i, j int) bool {
		if inbounds[i].Sort == inbounds[j].Sort {
			return inbounds[i].Tag < inbounds[j].Tag
		}
		return inbounds[i].Sort < inbounds[j].Sort
	})

	c.JSON(http.StatusOK, inbounds)
}

// UpdateInbound 更新 inbound 模板。
func (s *Service) UpdateInbound(c *gin.Context) {
	var inbound entity.Inbound
	if err := c.ShouldBindJSON(&inbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if pathTag := c.Param("tag"); pathTag != "" && pathTag != inbound.Tag {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path tag and body tag do not match"})
		return
	}

	if err := s.storage.UpdateInbound(&inbound); err != nil {
		handleStorageNotFound(c, err, "Inbound not found")
		return
	}

	c.JSON(http.StatusOK, inbound)
}

// DeleteInbound 删除 inbound 模板。
func (s *Service) DeleteInbound(c *gin.Context) {
	if err := s.storage.DeleteInbound(c.Param("tag")); err != nil {
		handleStorageNotFound(c, err, "Inbound not found")
		return
	}

	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// SetDeviceInbounds 全量更新设备绑定的 inbound 列表。
func (s *Service) SetDeviceInbounds(c *gin.Context) {
	deviceCode := c.Param("code")
	if _, err := s.storage.GetDevice(deviceCode); err != nil {
		handleStorageNotFound(c, err, "Device not found")
		return
	}

	var bindings []*entity.DeviceInbound
	if err := c.ShouldBindJSON(&bindings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for index, binding := range bindings {
		if binding == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bindings must not contain null item"})
			return
		}
		if binding.DeviceCode != "" && binding.DeviceCode != deviceCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Binding deviceCode does not match path code"})
			return
		}
		binding.DeviceCode = deviceCode

		if _, err := s.storage.GetInbound(binding.InboundTag); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Inbound not found for binding at index " + strconv.Itoa(index),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := s.storage.SetDeviceInbounds(deviceCode, bindings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bindings)
}

// ListDeviceInbounds 获取设备已绑定的 inbound 列表。
func (s *Service) ListDeviceInbounds(c *gin.Context) {
	deviceCode := c.Param("code")
	if _, err := s.storage.GetDevice(deviceCode); err != nil {
		handleStorageNotFound(c, err, "Device not found")
		return
	}

	bindings, err := s.storage.ListDeviceInbounds(deviceCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Sort == bindings[j].Sort {
			return bindings[i].InboundTag < bindings[j].InboundTag
		}
		return bindings[i].Sort < bindings[j].Sort
	})

	c.JSON(http.StatusOK, bindings)
}

// CreateWireGuard 创建 WireGuard 配置。
func (s *Service) CreateWireGuard(c *gin.Context) {
	var item entity.WireGuard
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := s.storage.GetWireGuard(item.Tag); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateWireGuard(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetWireGuard 获取单个 WireGuard 配置。
func (s *Service) GetWireGuard(c *gin.Context) {
	item, err := s.storage.GetWireGuard(c.Param("tag"))
	if err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}
	c.JSON(http.StatusOK, item)
}

// ListWireGuards 列出 WireGuard 配置。
func (s *Service) ListWireGuards(c *gin.Context) {
	items, err := s.storage.ListWireGuards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].Tag < items[j].Tag
		}
		return items[i].Sort < items[j].Sort
	})

	c.JSON(http.StatusOK, items)
}

// UpdateWireGuard 更新 WireGuard 配置。
func (s *Service) UpdateWireGuard(c *gin.Context) {
	var item entity.WireGuard
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if pathTag := c.Param("tag"); pathTag != "" && pathTag != item.Tag {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path tag and body tag do not match"})
		return
	}

	if err := s.storage.UpdateWireGuard(&item); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteWireGuard 删除 WireGuard 配置。
func (s *Service) DeleteWireGuard(c *gin.Context) {
	if err := s.storage.DeleteWireGuard(c.Param("tag")); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// CreateWireGuardPeer 创建指定 WireGuard 下的 peer。
func (s *Service) CreateWireGuardPeer(c *gin.Context) {
	wireGuardTag := c.Param("tag")
	if _, err := s.storage.GetWireGuard(wireGuardTag); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	var item entity.WireGuardPeer
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if item.WireGuardTag != "" && item.WireGuardTag != wireGuardTag {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path tag and body wireGuardTag do not match"})
		return
	}
	item.WireGuardTag = wireGuardTag

	if err := s.storage.CreateWireGuardPeer(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ListWireGuardPeers 列出指定 WireGuard 下的 peer。
func (s *Service) ListWireGuardPeers(c *gin.Context) {
	wireGuardTag := c.Param("tag")
	if _, err := s.storage.GetWireGuard(wireGuardTag); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	items, err := s.storage.ListWireGuardPeers(wireGuardTag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			return items[i].ID < items[j].ID
		}
		return items[i].Sort < items[j].Sort
	})

	c.JSON(http.StatusOK, items)
}

// UpdateWireGuardPeer 更新 peer。
func (s *Service) UpdateWireGuardPeer(c *gin.Context) {
	wireGuardTag := c.Param("tag")
	if _, err := s.storage.GetWireGuard(wireGuardTag); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	var item entity.WireGuardPeer
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if item.WireGuardTag != "" && item.WireGuardTag != wireGuardTag {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path tag and body wireGuardTag do not match"})
		return
	}
	item.WireGuardTag = wireGuardTag

	if pathID := c.Param("id"); pathID != "" {
		id, err := strconv.ParseInt(pathID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer id"})
			return
		}
		if item.ID != 0 && item.ID != id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path id and body id do not match"})
			return
		}
		item.ID = id
	}

	if err := s.storage.UpdateWireGuardPeer(&item); err != nil {
		handleStorageNotFound(c, err, "WireGuard peer not found")
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteWireGuardPeer 删除 peer。
func (s *Service) DeleteWireGuardPeer(c *gin.Context) {
	if _, err := s.storage.GetWireGuard(c.Param("tag")); err != nil {
		handleStorageNotFound(c, err, "WireGuard not found")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid peer id"})
		return
	}

	if err := s.storage.DeleteWireGuardPeer(id); err != nil {
		handleStorageNotFound(c, err, "WireGuard peer not found")
		return
	}

	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// CreateExtraOutbound 创建额外出站配置。
func (s *Service) CreateExtraOutbound(c *gin.Context) {
	var outbound entity.Outbound
	if err := c.ShouldBindJSON(&outbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 现有“额外出站”接口只允许管理手工维护的 MANUAL 记录。
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""

	if _, err := s.storage.GetExtraOutbound(outbound.Tag); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateExtraOutbound(&outbound); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, outbound)
}

// GetExtraOutbound 获取单个额外出站配置。
func (s *Service) GetExtraOutbound(c *gin.Context) {
	outbound, err := s.storage.GetExtraOutbound(c.Param("tag"))
	if err != nil {
		handleStorageNotFound(c, err, "Extra outbound not found")
		return
	}
	c.JSON(http.StatusOK, outbound)
}

// ListExtraOutbounds 列出额外出站配置。
func (s *Service) ListExtraOutbounds(c *gin.Context) {
	outbounds, err := s.storage.ListExtraOutbounds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(outbounds, func(i, j int) bool {
		if outbounds[i].Sort == outbounds[j].Sort {
			return outbounds[i].Tag < outbounds[j].Tag
		}
		return outbounds[i].Sort < outbounds[j].Sort
	})

	c.JSON(http.StatusOK, outbounds)
}

// UpdateExtraOutbound 更新额外出站配置。
func (s *Service) UpdateExtraOutbound(c *gin.Context) {
	var outbound entity.Outbound
	if err := c.ShouldBindJSON(&outbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if pathTag := c.Param("tag"); pathTag != "" && pathTag != outbound.Tag {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path tag and body tag do not match"})
		return
	}
	existing, err := s.storage.GetExtraOutbound(outbound.Tag)
	if err != nil {
		handleStorageNotFound(c, err, "Extra outbound not found")
		return
	}
	if existing.Source != entity.OutboundSourceManual {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription outbounds are read-only"})
		return
	}
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""

	if err := s.storage.UpdateExtraOutbound(&outbound); err != nil {
		handleStorageNotFound(c, err, "Extra outbound not found")
		return
	}

	c.JSON(http.StatusOK, outbound)
}

// DeleteExtraOutbound 删除额外出站配置。
func (s *Service) DeleteExtraOutbound(c *gin.Context) {
	outbound, err := s.storage.GetExtraOutbound(c.Param("tag"))
	if err != nil {
		handleStorageNotFound(c, err, "Extra outbound not found")
		return
	}
	if outbound.Source != entity.OutboundSourceManual {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription outbounds are read-only"})
		return
	}
	if err := s.storage.DeleteExtraOutbound(c.Param("tag")); err != nil {
		handleStorageNotFound(c, err, "Extra outbound not found")
		return
	}

	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// handleStorageNotFound 将存储层的未找到错误统一映射为 404，其他错误保持 500。
func handleStorageNotFound(c *gin.Context, err error, notFoundMessage string) {
	if errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMessage})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
