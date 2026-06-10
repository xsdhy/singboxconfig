package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// Service 服务层
type Service struct {
	storage storage.Storage
	auth    *authCache
}

// NewService 创建服务实例
func NewService(store storage.Storage) *Service {
	return &Service{
		storage: store,
		auth:    newAuthCache(authCacheTTL),
	}
}

// Subscribe 相关方法
func (s *Service) CreateSubscribe(c *gin.Context) {
	var subscribe entity.Subscribe
	if err := c.ShouldBindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.storage.CreateSubscribe(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

func (s *Service) UpdateSubscribe(c *gin.Context) {
	var subscribe entity.Subscribe
	if err := c.ShouldBindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.UpdateSubscribe(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

func (s *Service) DeleteSubscribe(c *gin.Context) {
	name := c.Param("name")
	if err := s.storage.DeleteSubscribe(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (s *Service) ListSubscribes(c *gin.Context) {
	subscribes, err := s.storage.ListSubscribes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscribes)
}

// NodeGroup 相关方法
func (s *Service) CreateNodeGroup(c *gin.Context) {
	var group entity.NodeGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查名称是否存在，不允许重复
	if _, err := s.storage.GetNodeGroup(group.Tag); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (s *Service) UpdateNodeGroup(c *gin.Context) {
	var group entity.NodeGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.UpdateNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (s *Service) DeleteNodeGroup(c *gin.Context) {
	if err := s.storage.DeleteNodeGroup(c.Param("tag")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (s *Service) ListNodeGroups(c *gin.Context) {
	groups, err := s.storage.ListNodeGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// RuleSet 相关方法
func (s *Service) CreateRuleSet(c *gin.Context) {
	var ruleSet entity.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查名称是否存在，不允许重复
	if _, err := s.storage.GetRuleSet(ruleSet.Tag); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	} else if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

func (s *Service) UpdateRuleSet(c *gin.Context) {
	var ruleSet entity.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 去掉content中的\n 和空格
	ruleSet.Content = strings.ReplaceAll(ruleSet.Content, "\n", "")
	ruleSet.Content = strings.ReplaceAll(ruleSet.Content, " ", "")

	if err := s.storage.UpdateRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

func (s *Service) DeleteRuleSet(c *gin.Context) {
	if err := s.storage.DeleteRuleSet(c.Param("tag")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (s *Service) GetRuleSetByTag(c *gin.Context) {
	tag := c.Param("tag")
	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, ruleSet := range ruleSets {
		if ruleSet.Tag == tag {
			var res json.RawMessage
			err := json.Unmarshal([]byte(ruleSet.Content), &res)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Rule set not found"})
}

func (s *Service) ListRuleSets(c *gin.Context) {
	ruleSets, err := s.storage.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 按 sort 字段排序
	sort.Slice(ruleSets, func(i, j int) bool {
		return ruleSets[i].Sort < ruleSets[j].Sort
	})

	c.JSON(http.StatusOK, ruleSets)
}

// GlobalSetting 相关方法
func (s *Service) CreateSetting(c *gin.Context) {
	var setting struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&setting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isReservedGlobalSettingKey(setting.Key) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reserved setting key"})
		return
	}

	// 系统 Host 等特殊 key 需要在落库前做规范化与合法性校验，非法值直接拒绝。
	normalizedValue, err := normalizeGlobalSettingValue(setting.Key, setting.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting.Value = normalizedValue

	if err := s.storage.SetGlobalSetting(setting.Key, setting.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (s *Service) UpdateSetting(c *gin.Context) {
	var setting struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&setting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := c.Param("key")
	if key == "" {
		key = setting.Key
	}
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}
	if isReservedGlobalSettingKey(key) || (setting.Key != "" && isReservedGlobalSettingKey(setting.Key)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reserved setting key"})
		return
	}
	setting.Key = key

	// 系统 Host 等特殊 key 需要在落库前做规范化与合法性校验，非法值直接拒绝。
	normalizedValue, err := normalizeGlobalSettingValue(setting.Key, setting.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting.Value = normalizedValue

	if err := s.storage.SetGlobalSetting(setting.Key, setting.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (s *Service) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if isReservedGlobalSettingKey(key) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reserved setting key"})
		return
	}
	if err := s.storage.DeleteGlobalSetting(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (s *Service) GetSetting(c *gin.Context) {
	key := c.Param("key")
	if isReservedGlobalSettingKey(key) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reserved setting key"})
		return
	}
	value, err := s.storage.GetGlobalSetting(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

func (s *Service) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	if isReservedGlobalSettingKey(key) {
		c.JSON(http.StatusForbidden, gin.H{"error": "reserved setting key"})
		return
	}
	value, err := s.storage.GetGlobalSetting(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

func (s *Service) ListSettings(c *gin.Context) {
	settings, err := s.storage.ListGlobalSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for key := range settings {
		if isReservedGlobalSettingKey(key) {
			delete(settings, key)
		}
	}

	c.JSON(http.StatusOK, settings)
}
