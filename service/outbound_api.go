package service

import (
	"errors"
	"net/http"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type outboundListResponse struct {
	Items []*entity.Outbound `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

type outboundBatchEnableRequest struct {
	IDs     []int64 `json:"ids"`
	Enabled bool    `json:"enabled"`
}

type subscribeCacheConfigRequest struct {
	OutboundCacheDuration    *int `json:"outboundCacheDuration"`
	OutboundCacheDurationAlt *int `json:"outbound_cache_duration"`
}

type subscribeCacheInfo struct {
	LastFetchTime *time.Time `json:"lastFetchTime,omitempty"`
	CacheDuration int        `json:"cacheDuration"`
	IsExpired     bool       `json:"isExpired"`
}

type subscribeOutboundListResponse struct {
	Items              []*entity.Outbound `json:"items"`
	Total              int                `json:"total"`
	Page               int                `json:"page"`
	Limit              int                `json:"limit"`
	SubscribeCacheInfo subscribeCacheInfo `json:"subscribeCacheInfo"`
}

// ListOutbounds 提供统一 Outbound 列表查询，支持来源、订阅、启用状态和关键词筛选。
func (s *Service) ListOutbounds(c *gin.Context) {
	filters, err := parseOutboundFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := s.storage.ListOutbounds(filters...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	search := strings.TrimSpace(strings.ToLower(c.Query("search")))
	filtered := make([]*entity.Outbound, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if search != "" && !outboundMatchesSearch(item, search) {
			continue
		}
		filtered = append(filtered, item)
	}

	sortOutbounds(filtered)
	page, limit := parsePagination(c)
	c.JSON(http.StatusOK, paginateOutbounds(filtered, page, limit))
}

// CreateOutbound 创建新的手工 Outbound。
// 统一接口只允许新增 MANUAL 记录，避免前端误写订阅缓存数据。
func (s *Service) CreateOutbound(c *gin.Context) {
	var outbound entity.Outbound
	if err := c.ShouldBindJSON(&outbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if outbound.Tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag is required"})
		return
	}
	existing, err := s.findOutboundByTag(outbound.Tag)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	}

	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""
	outbound.LastFetchTime = nil
	if err := s.storage.CreateOrUpdateOutbounds([]*entity.Outbound{&outbound}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.findOutboundByTag(outbound.Tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

// UpdateOutbound 编辑手工 Outbound。
// 订阅来源只允许通过刷新流程覆盖，因此这里明确拒绝编辑。
func (s *Service) UpdateOutbound(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var payload entity.Outbound
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.ID != 0 && payload.ID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path id and body id do not match"})
		return
	}

	existing, err := s.storage.GetOutbound(id)
	if err != nil {
		handleStorageNotFound(c, err, "Outbound not found")
		return
	}
	if existing.Source != entity.OutboundSourceManual {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription outbounds are read-only"})
		return
	}
	if payload.Tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag is required"})
		return
	}
	duplicate, err := s.findOutboundByTag(payload.Tag)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if duplicate != nil && duplicate.ID != existing.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag already exists"})
		return
	}

	existing.Tag = payload.Tag
	existing.Name = payload.Name
	existing.Description = payload.Description
	existing.Type = payload.Type
	existing.Enabled = payload.Enabled
	existing.Sort = payload.Sort
	existing.VisibleDevices = payload.VisibleDevices
	existing.ConfigJSON = payload.ConfigJSON
	existing.Source = entity.OutboundSourceManual
	existing.SubscribeName = ""
	existing.LastFetchTime = nil

	if err := s.storage.UpdateOutbound(existing); err != nil {
		handleStorageNotFound(c, err, "Outbound not found")
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteOutbound 删除统一 Outbound 记录。
// 对订阅来源删除属于显式管理动作，但下次刷新时允许被重新写回。
func (s *Service) DeleteOutbound(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := s.storage.DeleteOutbound(id); err != nil {
		handleStorageNotFound(c, err, "Outbound not found")
		return
	}
	c.JSON(http.StatusOK, deleteSuccessResponse)
}

// BatchEnableOutbounds 批量更新 Outbound 启用状态。
func (s *Service) BatchEnableOutbounds(c *gin.Context) {
	var req outboundBatchEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids must not be empty"})
		return
	}

	updated := 0
	for _, id := range req.IDs {
		outbound, err := s.storage.GetOutbound(id)
		if err != nil {
			handleStorageNotFound(c, err, "Outbound not found")
			return
		}
		outbound.Enabled = req.Enabled
		if err := s.storage.UpdateOutbound(outbound); err != nil {
			handleStorageNotFound(c, err, "Outbound not found")
			return
		}
		updated++
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated, "enabled": req.Enabled})
}

// UpdateSubscribeCacheConfig 更新订阅缓存时长配置。
func (s *Service) UpdateSubscribeCacheConfig(c *gin.Context) {
	var req subscribeCacheConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration, ok := req.cacheDuration()
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outboundCacheDuration is required"})
		return
	}
	if duration < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outboundCacheDuration must be >= 0"})
		return
	}

	subscribe, err := s.storage.GetSubscribe(c.Param("name"))
	if err != nil {
		handleStorageNotFound(c, err, "Subscribe not found")
		return
	}
	subscribe.OutboundCacheDuration = duration
	if err := s.storage.UpdateSubscribe(subscribe); err != nil {
		handleStorageNotFound(c, err, "Subscribe not found")
		return
	}
	c.JSON(http.StatusOK, subscribe)
}

// RefreshSubscribeOutbounds 手动刷新单个订阅的 Outbound 缓存，并返回变更统计。
func (s *Service) RefreshSubscribeOutbounds(c *gin.Context) {
	result, err := s.refreshSubscriptionOutboundDetailed(c.Request.Context(), c.Param("name"))
	if err != nil {
		handleStorageNotFound(c, err, "Subscribe not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"added":         result.Added,
		"updated":       result.Updated,
		"deleted":       result.Deleted,
		"lastFetchTime": result.LastFetchTime,
	})
}

// ListSubscribeOutbounds 查询单个订阅下缓存的 Outbound 列表，并附带缓存状态摘要。
func (s *Service) ListSubscribeOutbounds(c *gin.Context) {
	subscribe, err := s.storage.GetSubscribe(c.Param("name"))
	if err != nil {
		handleStorageNotFound(c, err, "Subscribe not found")
		return
	}
	items, err := s.storage.ListOutbounds(
		storage.OutboundFilter{Source: outboundSourcePtr(entity.OutboundSourceSubscription)},
		storage.OutboundFilter{SubscribeName: subscribe.Name},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sortOutbounds(items)
	page, limit := parsePagination(c)

	c.JSON(http.StatusOK, subscribeOutboundListResponse{
		Items: paginateOutbounds(items, page, limit).Items,
		Total: len(items),
		Page:  page,
		Limit: limit,
		SubscribeCacheInfo: subscribeCacheInfo{
			LastFetchTime: subscribe.OutboundLastFetchTime,
			CacheDuration: subscribe.OutboundCacheDuration,
			IsExpired:     needsRefresh(subscribe),
		},
	})
}

func parseOutboundFilters(c *gin.Context) ([]storage.OutboundFilter, error) {
	filters := make([]storage.OutboundFilter, 0, 3)
	if sourceRaw := strings.TrimSpace(c.Query("source")); sourceRaw != "" {
		source := entity.OutboundSource(strings.ToUpper(sourceRaw))
		if source != entity.OutboundSourceManual && source != entity.OutboundSourceSubscription {
			return nil, errors.New("invalid source")
		}
		filters = append(filters, storage.OutboundFilter{Source: &source})
	}
	if enabledRaw := strings.TrimSpace(c.Query("enabled")); enabledRaw != "" {
		enabled, err := strconv.ParseBool(enabledRaw)
		if err != nil {
			return nil, errors.New("invalid enabled")
		}
		filters = append(filters, storage.OutboundFilter{Enabled: &enabled})
	}
	if subscribeName := strings.TrimSpace(c.Query("subscribe_name")); subscribeName != "" {
		filters = append(filters, storage.OutboundFilter{SubscribeName: subscribeName})
	}
	return filters, nil
}

func parsePagination(c *gin.Context) (int, int) {
	page := 1
	limit := 20
	if value := strings.TrimSpace(c.Query("page")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if value := strings.TrimSpace(c.Query("limit")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return page, limit
}

func paginateOutbounds(items []*entity.Outbound, page int, limit int) outboundListResponse {
	total := len(items)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return outboundListResponse{
		Items: items[start:end],
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func outboundMatchesSearch(item *entity.Outbound, search string) bool {
	return strings.Contains(strings.ToLower(item.Tag), search) ||
		strings.Contains(strings.ToLower(item.Name), search)
}

func sortOutbounds(items []*entity.Outbound) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort != items[j].Sort {
			return items[i].Sort < items[j].Sort
		}
		if items[i].Tag != items[j].Tag {
			return items[i].Tag < items[j].Tag
		}
		return items[i].ID < items[j].ID
	})
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outbound id"})
		return 0, false
	}
	return value, true
}

func (s *Service) findOutboundByTag(tag string) (*entity.Outbound, error) {
	items, err := s.storage.ListOutbounds()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item != nil && item.Tag == tag {
			return item, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (r subscribeCacheConfigRequest) cacheDuration() (int, bool) {
	if r.OutboundCacheDuration != nil {
		return *r.OutboundCacheDuration, true
	}
	if r.OutboundCacheDurationAlt != nil {
		return *r.OutboundCacheDurationAlt, true
	}
	return 0, false
}
