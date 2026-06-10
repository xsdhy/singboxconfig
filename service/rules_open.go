package service

import (
	"errors"
	"net/http"
	"singboxconfig/convert/ruleset"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetRulesBySoftware 处理“规则集级” open 接口：GET /open/rules/:tag/:software/:device?token=...
//
// 它针对单个规则集、按指定软件格式输出该规则集的规则内容（而非整份配置），
// 供对应客户端作为远程规则集直接加载。鉴权与可见性完全复用整份配置接口的语义：
//   - 设备不存在 → 404；设备禁用 → 403；token 不匹配 → 401；
//   - 规则集 tag 未命中 → 404；规则集对当前设备不可见（AbleDevices）→ 404（避免泄露存在性）；
//   - 非法 software → 400；
//   - remote 规则集 → 400（remote 已有原始 URL，本接口不做代理转换，避免外部请求与 SSRF 风险）；
//   - local / inline 内容非法 → 500。
//
// 真正的解析与渲染交给 convert/ruleset 纯函数完成，本函数只负责 HTTP 编排，因此不写 HTTP 层单测。
func (s *Service) GetRulesBySoftware(c *gin.Context) {
	tag := c.Param("tag")
	softwareParam := c.Param("software")
	deviceCode := c.Param("device")
	token := c.Query("token")

	// 1. 校验目标软件，未知软件返回 400。
	software, ok := entity.ParseSoftware(softwareParam)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported software"})
		return
	}

	// 2. 复用整份配置接口的设备解析与鉴权。
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

	// 3. 查找规则集，未命中或对当前设备不可见都按 404 处理，避免泄露规则集存在性。
	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ruleSet := findVisibleRuleSet(ruleSets, tag, device.Code)
	if ruleSet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule set not found"})
		return
	}

	// 4. remote 规则集已有原始 URL，本接口不做代理转换。
	if entity.RuleSetType(ruleSet.RuleSetType) == entity.RuleSetTypeRemote {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Remote ruleset has its own url and is not proxied by this endpoint"})
		return
	}

	// 5. 按软件渲染该规则集的规则内容。
	switch software {
	case entity.SoftwareSingbox:
		normalized, err := ruleset.NormalizeSingbox(ruleSet.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid ruleset content"})
			return
		}
		c.Data(http.StatusOK, "application/json", normalized)
	case entity.SoftwareSurge, entity.SoftwareShadowrocket:
		lines, warnings, err := ruleset.RenderLines(ruleSet.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid ruleset content"})
			return
		}
		for _, warning := range warnings {
			logrus.Warnf("GetRulesBySoftware: tag=%s %s", ruleSet.Tag, warning)
		}
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(strings.Join(lines, "\n")))
	}
}

// findVisibleRuleSet 在规则集列表中按 tag 查找，并校验对当前设备可见。
// AbleDevices 为空表示全部可见；非空时沿用 strings.Contains(ableDevices, deviceCode) 判定。
// 命中但不可见时返回 nil，由调用方按 404 处理，避免泄露规则集存在性。
func findVisibleRuleSet(ruleSets []*entity.RuleSet, tag string, deviceCode string) *entity.RuleSet {
	for _, ruleSet := range ruleSets {
		if ruleSet == nil || ruleSet.Tag != tag {
			continue
		}
		if ruleSet.AbleDevices != "" && !strings.Contains(ruleSet.AbleDevices, deviceCode) {
			return nil
		}
		return ruleSet
	}
	return nil
}
