package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"singboxconfig/entity"
	"strconv"
	"strings"
	"time"
)

// SupabaseStorage 通过 Supabase PostgREST HTTP API 实现 Storage 接口
type SupabaseStorage struct {
	baseURL    string // e.g. "https://xxxx.supabase.co/rest/v1"
	apiKey     string // service_role key
	httpClient *http.Client
}

// NewSupabaseStorage 创建 SupabaseStorage 实例
func NewSupabaseStorage(supabaseURL, apiKey string) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:    supabaseURL + "/rest/v1",
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Supabase 数据模型（snake_case JSON tags 匹配 PostgreSQL 列名）---

type supabaseSubscribe struct {
	Name                    string     `json:"name"`
	URL                     string     `json:"url"`
	UserAgent               string     `json:"user_agent"`
	Status                  bool       `json:"status"`
	VisibleDevices          string     `json:"visible_devices"`
	OutboundLastFetchTime   *time.Time `json:"outbound_last_fetch_time"`
	OutboundCacheDuration   int        `json:"outbound_cache_duration"`
	OutboundLastFetchStatus string     `json:"outbound_last_fetch_status"`
	OutboundLastFetchError  string     `json:"outbound_last_fetch_error"`
}

func (s *supabaseSubscribe) toEntity() *entity.Subscribe {
	return &entity.Subscribe{
		Name:                    s.Name,
		URL:                     s.URL,
		UserAgent:               s.UserAgent,
		Status:                  s.Status,
		VisibleDevices:          s.VisibleDevices,
		OutboundLastFetchTime:   s.OutboundLastFetchTime,
		OutboundCacheDuration:   s.OutboundCacheDuration,
		OutboundLastFetchStatus: s.OutboundLastFetchStatus,
		OutboundLastFetchError:  s.OutboundLastFetchError,
	}
}

func supabaseSubscribeFromEntity(e *entity.Subscribe) *supabaseSubscribe {
	return &supabaseSubscribe{
		Name:                    e.Name,
		URL:                     e.URL,
		UserAgent:               e.UserAgent,
		Status:                  e.Status,
		VisibleDevices:          e.VisibleDevices,
		OutboundLastFetchTime:   e.OutboundLastFetchTime,
		OutboundCacheDuration:   e.OutboundCacheDuration,
		OutboundLastFetchStatus: e.OutboundLastFetchStatus,
		OutboundLastFetchError:  e.OutboundLastFetchError,
	}
}

type supabaseNodeGroup struct {
	Tag                 string `json:"tag"`
	Name                string `json:"name"`
	GroupType           string `json:"group_type"`
	TestURL             string `json:"test_url"`
	Include             string `json:"include"`
	Exclude             string `json:"exclude"`
	DeviceTypeOverrides string `json:"device_type_overrides"`
}

func (s *supabaseNodeGroup) toEntity() *entity.NodeGroup {
	return &entity.NodeGroup{
		Tag:                 s.Tag,
		Name:                s.Name,
		GroupType:           s.GroupType,
		TestURL:             s.TestURL,
		Include:             s.Include,
		Exclude:             s.Exclude,
		DeviceTypeOverrides: s.DeviceTypeOverrides,
	}
}

func supabaseNodeGroupFromEntity(e *entity.NodeGroup) *supabaseNodeGroup {
	return &supabaseNodeGroup{
		Tag:                 e.Tag,
		Name:                e.Name,
		GroupType:           e.GroupType,
		TestURL:             e.TestURL,
		Include:             e.Include,
		Exclude:             e.Exclude,
		DeviceTypeOverrides: e.DeviceTypeOverrides,
	}
}

type supabaseRuleSet struct {
	Tag            string `json:"tag"`
	Name           string `json:"name"`
	RuleSetType    string `json:"rule_set_type"`
	Format         string `json:"format"`
	Content        string `json:"content"`
	URL            string `json:"url"`
	Outbound       string `json:"outbound"`
	DownloadDetour string `json:"download_detour"`
	AbleDevices    string `json:"able_devices"`
	Sort           int    `json:"sort"`
}

func (s *supabaseRuleSet) toEntity() *entity.RuleSet {
	return &entity.RuleSet{
		Tag:            s.Tag,
		Name:           s.Name,
		RuleSetType:    s.RuleSetType,
		Format:         s.Format,
		Content:        s.Content,
		URL:            s.URL,
		Outbound:       s.Outbound,
		DownloadDetour: s.DownloadDetour,
		AbleDevices:    s.AbleDevices,
		Sort:           s.Sort,
	}
}

func supabaseRuleSetFromEntity(e *entity.RuleSet) *supabaseRuleSet {
	return &supabaseRuleSet{
		Tag:            e.Tag,
		Name:           e.Name,
		RuleSetType:    e.RuleSetType,
		Format:         e.Format,
		Content:        e.Content,
		URL:            e.URL,
		Outbound:       e.Outbound,
		DownloadDetour: e.DownloadDetour,
		AbleDevices:    e.AbleDevices,
		Sort:           e.Sort,
	}
}

type supabaseGlobalSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type supabaseDevice struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Token               string `json:"token"`
	Enabled             bool   `json:"enabled"`
	Sort                int    `json:"sort"`
	WireGuardTag        string `json:"wire_guard_tag"`
	WireGuardClientAddr string `json:"wire_guard_client_addr"`
	WireGuardClientKey  string `json:"wire_guard_client_key"`
}

func (s *supabaseDevice) toEntity() *entity.Device {
	return &entity.Device{
		Code:                s.Code,
		Name:                s.Name,
		Description:         s.Description,
		Token:               s.Token,
		Enabled:             s.Enabled,
		Sort:                s.Sort,
		WireGuardTag:        s.WireGuardTag,
		WireGuardClientAddr: s.WireGuardClientAddr,
		WireGuardClientKey:  s.WireGuardClientKey,
	}
}

func supabaseDeviceFromEntity(e *entity.Device) *supabaseDevice {
	return &supabaseDevice{
		Code:                e.Code,
		Name:                e.Name,
		Description:         e.Description,
		Token:               e.Token,
		Enabled:             e.Enabled,
		Sort:                e.Sort,
		WireGuardTag:        e.WireGuardTag,
		WireGuardClientAddr: e.WireGuardClientAddr,
		WireGuardClientKey:  e.WireGuardClientKey,
	}
}

type supabaseInbound struct {
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
	ConfigJSON  string `json:"config_json"`
}

func (s *supabaseInbound) toEntity() *entity.Inbound {
	return &entity.Inbound{
		Tag:         s.Tag,
		Name:        s.Name,
		Description: s.Description,
		Type:        s.Type,
		Enabled:     s.Enabled,
		Sort:        s.Sort,
		ConfigJSON:  s.ConfigJSON,
	}
}

func supabaseInboundFromEntity(e *entity.Inbound) *supabaseInbound {
	return &supabaseInbound{
		Tag:         e.Tag,
		Name:        e.Name,
		Description: e.Description,
		Type:        e.Type,
		Enabled:     e.Enabled,
		Sort:        e.Sort,
		ConfigJSON:  e.ConfigJSON,
	}
}

type supabaseDeviceInbound struct {
	DeviceCode string `json:"device_code"`
	InboundTag string `json:"inbound_tag"`
	Sort       int    `json:"sort"`
}

func (s *supabaseDeviceInbound) toEntity() *entity.DeviceInbound {
	return &entity.DeviceInbound{
		DeviceCode: s.DeviceCode,
		InboundTag: s.InboundTag,
		Sort:       s.Sort,
	}
}

func supabaseDeviceInboundFromEntity(e *entity.DeviceInbound) *supabaseDeviceInbound {
	return &supabaseDeviceInbound{
		DeviceCode: e.DeviceCode,
		InboundTag: e.InboundTag,
		Sort:       e.Sort,
	}
}

type supabaseWireGuard struct {
	Tag         string `json:"tag"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Sort        int    `json:"sort"`
	EndpointTag string `json:"endpoint_tag"`
	MTU         int    `json:"mtu"`
}

func (s *supabaseWireGuard) toEntity() *entity.WireGuard {
	return &entity.WireGuard{
		Tag:         s.Tag,
		Name:        s.Name,
		Description: s.Description,
		Enabled:     s.Enabled,
		Sort:        s.Sort,
		EndpointTag: s.EndpointTag,
		MTU:         s.MTU,
	}
}

func supabaseWireGuardFromEntity(e *entity.WireGuard) *supabaseWireGuard {
	return &supabaseWireGuard{
		Tag:         e.Tag,
		Name:        e.Name,
		Description: e.Description,
		Enabled:     e.Enabled,
		Sort:        e.Sort,
		EndpointTag: e.EndpointTag,
		MTU:         e.MTU,
	}
}

type supabaseWireGuardPeer struct {
	ID                          int64  `json:"id,omitempty"`
	WireGuardTag                string `json:"wire_guard_tag"`
	Sort                        int    `json:"sort"`
	Address                     string `json:"address"`
	Port                        int    `json:"port"`
	PublicKey                   string `json:"public_key"`
	PreSharedKey                string `json:"pre_shared_key"`
	AllowedIPs                  string `json:"allowed_ips"`
	PersistentKeepaliveInterval int    `json:"persistent_keepalive_interval"`
	Enabled                     bool   `json:"enabled"`
}

func (s *supabaseWireGuardPeer) toEntity() *entity.WireGuardPeer {
	return &entity.WireGuardPeer{
		ID:                          s.ID,
		WireGuardTag:                s.WireGuardTag,
		Sort:                        s.Sort,
		Address:                     s.Address,
		Port:                        s.Port,
		PublicKey:                   s.PublicKey,
		PreSharedKey:                s.PreSharedKey,
		AllowedIPs:                  s.AllowedIPs,
		PersistentKeepaliveInterval: s.PersistentKeepaliveInterval,
		Enabled:                     s.Enabled,
	}
}

func supabaseWireGuardPeerFromEntity(e *entity.WireGuardPeer) *supabaseWireGuardPeer {
	return &supabaseWireGuardPeer{
		ID:                          e.ID,
		WireGuardTag:                e.WireGuardTag,
		Sort:                        e.Sort,
		Address:                     e.Address,
		Port:                        e.Port,
		PublicKey:                   e.PublicKey,
		PreSharedKey:                e.PreSharedKey,
		AllowedIPs:                  e.AllowedIPs,
		PersistentKeepaliveInterval: e.PersistentKeepaliveInterval,
		Enabled:                     e.Enabled,
	}
}

type supabaseExtraOutbound struct {
	// ID 仅在已存在记录或按主键更新时回传给 Supabase。
	// 新建记录若传入 id=0，PostgREST 会把它当成显式主键写入，导致主键冲突，
	// 因此这里必须使用 omitempty，让数据库自己分配 identity 值。
	ID             int64                 `json:"id,omitempty"`
	Tag            string                `json:"tag"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Type           string                `json:"type"`
	Enabled        bool                  `json:"enabled"`
	Sort           int                   `json:"sort"`
	VisibleDevices string                `json:"visible_devices"`
	ConfigJSON     string                `json:"config_json"`
	Source         entity.OutboundSource `json:"source"`
	SubscribeName  string                `json:"subscribe_name"`
	LastFetchTime  *time.Time            `json:"last_fetch_time"`
}

func (s *supabaseExtraOutbound) toEntity() *entity.Outbound {
	return &entity.Outbound{
		ID:             s.ID,
		Tag:            s.Tag,
		Name:           s.Name,
		Description:    s.Description,
		Type:           s.Type,
		Enabled:        s.Enabled,
		Sort:           s.Sort,
		VisibleDevices: s.VisibleDevices,
		ConfigJSON:     s.ConfigJSON,
		Source:         s.Source,
		SubscribeName:  s.SubscribeName,
		LastFetchTime:  s.LastFetchTime,
	}
}

func supabaseExtraOutboundFromEntity(e *entity.Outbound) *supabaseExtraOutbound {
	return &supabaseExtraOutbound{
		ID:             e.ID,
		Tag:            e.Tag,
		Name:           e.Name,
		Description:    e.Description,
		Type:           e.Type,
		Enabled:        e.Enabled,
		Sort:           e.Sort,
		VisibleDevices: e.VisibleDevices,
		ConfigJSON:     e.ConfigJSON,
		Source:         e.Source,
		SubscribeName:  e.SubscribeName,
		LastFetchTime:  e.LastFetchTime,
	}
}

// --- HTTP 请求封装 ---

// request 发送 HTTP 请求到 Supabase PostgREST
func (s *SupabaseStorage) request(method, path string, body interface{}, headers map[string]string) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	// 公共 Header
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 额外 Header
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// --- Subscribe CRUD ---

func (s *SupabaseStorage) CreateSubscribe(subscribe *entity.Subscribe) error {
	sb := supabaseSubscribeFromEntity(subscribe)
	_, statusCode, err := s.request("POST", "/subscribes", sb, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("subscribe %q already exists", subscribe.Name)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create subscribe failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetSubscribe(name string) (*entity.Subscribe, error) {
	path := "/subscribes?name=eq." + url.QueryEscape(name) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get subscribe failed, status=%d body=%s", statusCode, string(body))
	}
	var sb supabaseSubscribe
	if err := json.Unmarshal(body, &sb); err != nil {
		return nil, fmt.Errorf("unmarshal subscribe: %w", err)
	}
	return sb.toEntity(), nil
}

func (s *SupabaseStorage) ListSubscribes() ([]*entity.Subscribe, error) {
	body, statusCode, err := s.request("GET", "/subscribes?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list subscribes failed, status=%d body=%s", statusCode, string(body))
	}
	var sbs []supabaseSubscribe
	if err := json.Unmarshal(body, &sbs); err != nil {
		return nil, fmt.Errorf("unmarshal subscribes: %w", err)
	}
	result := make([]*entity.Subscribe, len(sbs))
	for i := range sbs {
		result[i] = sbs[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateSubscribe(subscribe *entity.Subscribe) error {
	sb := supabaseSubscribeFromEntity(subscribe)
	path := "/subscribes?name=eq." + url.QueryEscape(subscribe.Name)
	body, statusCode, err := s.request("PATCH", path, sb, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update subscribe failed, status=%d body=%s", statusCode, string(body))
	}
	// 检查是否有行被更新（返回空数组表示未找到）
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteSubscribe(name string) error {
	path := "/subscribes?name=eq." + url.QueryEscape(name)
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete subscribe failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- NodeGroup CRUD ---

func (s *SupabaseStorage) CreateNodeGroup(group *entity.NodeGroup) error {
	sg := supabaseNodeGroupFromEntity(group)
	_, statusCode, err := s.request("POST", "/node_groups", sg, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("node group %q already exists", group.Tag)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create node group failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetNodeGroup(tag string) (*entity.NodeGroup, error) {
	path := "/node_groups?tag=eq." + url.QueryEscape(tag) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get node group failed, status=%d body=%s", statusCode, string(body))
	}
	var sg supabaseNodeGroup
	if err := json.Unmarshal(body, &sg); err != nil {
		return nil, fmt.Errorf("unmarshal node group: %w", err)
	}
	return sg.toEntity(), nil
}

func (s *SupabaseStorage) ListNodeGroups() ([]*entity.NodeGroup, error) {
	body, statusCode, err := s.request("GET", "/node_groups?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list node groups failed, status=%d body=%s", statusCode, string(body))
	}
	var sgs []supabaseNodeGroup
	if err := json.Unmarshal(body, &sgs); err != nil {
		return nil, fmt.Errorf("unmarshal node groups: %w", err)
	}
	result := make([]*entity.NodeGroup, len(sgs))
	for i := range sgs {
		result[i] = sgs[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateNodeGroup(group *entity.NodeGroup) error {
	sg := supabaseNodeGroupFromEntity(group)
	path := "/node_groups?tag=eq." + url.QueryEscape(group.Tag)
	body, statusCode, err := s.request("PATCH", path, sg, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update node group failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteNodeGroup(tag string) error {
	path := "/node_groups?tag=eq." + url.QueryEscape(tag)
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete node group failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- RuleSet CRUD ---

func (s *SupabaseStorage) CreateRuleSet(ruleSet *entity.RuleSet) error {
	sr := supabaseRuleSetFromEntity(ruleSet)
	_, statusCode, err := s.request("POST", "/rule_sets", sr, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("rule set %q already exists", ruleSet.Tag)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create rule set failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetRuleSet(tag string) (*entity.RuleSet, error) {
	path := "/rule_sets?tag=eq." + url.QueryEscape(tag) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get rule set failed, status=%d body=%s", statusCode, string(body))
	}
	var sr supabaseRuleSet
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("unmarshal rule set: %w", err)
	}
	return sr.toEntity(), nil
}

func (s *SupabaseStorage) ListRuleSets() ([]*entity.RuleSet, error) {
	body, statusCode, err := s.request("GET", "/rule_sets?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list rule sets failed, status=%d body=%s", statusCode, string(body))
	}
	var srs []supabaseRuleSet
	if err := json.Unmarshal(body, &srs); err != nil {
		return nil, fmt.Errorf("unmarshal rule sets: %w", err)
	}
	result := make([]*entity.RuleSet, len(srs))
	for i := range srs {
		result[i] = srs[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateRuleSet(ruleSet *entity.RuleSet) error {
	sr := supabaseRuleSetFromEntity(ruleSet)
	path := "/rule_sets?tag=eq." + url.QueryEscape(ruleSet.Tag)
	body, statusCode, err := s.request("PATCH", path, sr, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update rule set failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteRuleSet(tag string) error {
	path := "/rule_sets?tag=eq." + url.QueryEscape(tag)
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete rule set failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- GlobalSetting CRUD ---

func (s *SupabaseStorage) SetGlobalSetting(key, value string) error {
	sg := supabaseGlobalSetting{Key: key, Value: value}
	_, statusCode, err := s.request("POST", "/global_settings", sg, map[string]string{
		"Prefer": "resolution=merge-duplicates,return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: set global setting failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetGlobalSetting(key string) (string, error) {
	path := "/global_settings?key=eq." + url.QueryEscape(key) + "&select=value"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return "", err
	}
	if statusCode == http.StatusNotAcceptable {
		return "", ErrNotFound
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("supabase: get global setting failed, status=%d body=%s", statusCode, string(body))
	}
	var sg supabaseGlobalSetting
	if err := json.Unmarshal(body, &sg); err != nil {
		return "", fmt.Errorf("unmarshal global setting: %w", err)
	}
	return sg.Value, nil
}

func (s *SupabaseStorage) ListGlobalSettings() (map[string]string, error) {
	body, statusCode, err := s.request("GET", "/global_settings?select=key,value", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list global settings failed, status=%d body=%s", statusCode, string(body))
	}
	var sgs []supabaseGlobalSetting
	if err := json.Unmarshal(body, &sgs); err != nil {
		return nil, fmt.Errorf("unmarshal global settings: %w", err)
	}
	result := make(map[string]string, len(sgs))
	for _, sg := range sgs {
		result[sg.Key] = sg.Value
	}
	return result, nil
}

func (s *SupabaseStorage) DeleteGlobalSetting(key string) error {
	path := "/global_settings?key=eq." + url.QueryEscape(key)
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete global setting failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- Device CRUD ---

func (s *SupabaseStorage) CreateDevice(device *entity.Device) error {
	sd := supabaseDeviceFromEntity(device)
	_, statusCode, err := s.request("POST", "/devices", sd, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("device %q already exists", device.Code)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create device failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetDevice(code string) (*entity.Device, error) {
	path := "/devices?code=eq." + url.QueryEscape(code) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get device failed, status=%d body=%s", statusCode, string(body))
	}
	var sd supabaseDevice
	if err := json.Unmarshal(body, &sd); err != nil {
		return nil, fmt.Errorf("unmarshal device: %w", err)
	}
	return sd.toEntity(), nil
}

func (s *SupabaseStorage) ListDevices() ([]*entity.Device, error) {
	body, statusCode, err := s.request("GET", "/devices?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list devices failed, status=%d body=%s", statusCode, string(body))
	}
	var sds []supabaseDevice
	if err := json.Unmarshal(body, &sds); err != nil {
		return nil, fmt.Errorf("unmarshal devices: %w", err)
	}
	result := make([]*entity.Device, len(sds))
	for i := range sds {
		result[i] = sds[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateDevice(device *entity.Device) error {
	sd := supabaseDeviceFromEntity(device)
	path := "/devices?code=eq." + url.QueryEscape(device.Code)
	body, statusCode, err := s.request("PATCH", path, sd, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update device failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteDevice(code string) error {
	// 先删绑定关系，兼容未配置外键级联的 Supabase 库。
	bindingPath := "/device_inbounds?device_code=eq." + url.QueryEscape(code)
	body, statusCode, err := s.request("DELETE", bindingPath, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: clear device inbounds failed, status=%d body=%s", statusCode, string(body))
	}

	path := "/devices?code=eq." + url.QueryEscape(code)
	body, statusCode, err = s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete device failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- Inbound CRUD ---

func (s *SupabaseStorage) CreateInbound(inbound *entity.Inbound) error {
	si := supabaseInboundFromEntity(inbound)
	_, statusCode, err := s.request("POST", "/inbounds", si, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("inbound %q already exists", inbound.Tag)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create inbound failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetInbound(tag string) (*entity.Inbound, error) {
	path := "/inbounds?tag=eq." + url.QueryEscape(tag) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get inbound failed, status=%d body=%s", statusCode, string(body))
	}
	var si supabaseInbound
	if err := json.Unmarshal(body, &si); err != nil {
		return nil, fmt.Errorf("unmarshal inbound: %w", err)
	}
	return si.toEntity(), nil
}

func (s *SupabaseStorage) ListInbounds() ([]*entity.Inbound, error) {
	body, statusCode, err := s.request("GET", "/inbounds?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list inbounds failed, status=%d body=%s", statusCode, string(body))
	}
	var sis []supabaseInbound
	if err := json.Unmarshal(body, &sis); err != nil {
		return nil, fmt.Errorf("unmarshal inbounds: %w", err)
	}
	result := make([]*entity.Inbound, len(sis))
	for i := range sis {
		result[i] = sis[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateInbound(inbound *entity.Inbound) error {
	si := supabaseInboundFromEntity(inbound)
	path := "/inbounds?tag=eq." + url.QueryEscape(inbound.Tag)
	body, statusCode, err := s.request("PATCH", path, si, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update inbound failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteInbound(tag string) error {
	// 删除 Inbound 前先清理设备绑定，避免因为引用关系导致删除失败。
	bindingPath := "/device_inbounds?inbound_tag=eq." + url.QueryEscape(tag)
	body, statusCode, err := s.request("DELETE", bindingPath, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: clear inbound bindings failed, status=%d body=%s", statusCode, string(body))
	}

	path := "/inbounds?tag=eq." + url.QueryEscape(tag)
	body, statusCode, err = s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete inbound failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- DeviceInbound CRUD ---

func (s *SupabaseStorage) SetDeviceInbounds(deviceCode string, bindings []*entity.DeviceInbound) error {
	path := "/device_inbounds?device_code=eq." + url.QueryEscape(deviceCode)
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: clear device inbounds failed, status=%d body=%s", statusCode, string(body))
	}

	records := make([]*supabaseDeviceInbound, 0, len(bindings))
	for _, binding := range bindings {
		if binding == nil {
			continue
		}
		records = append(records, supabaseDeviceInboundFromEntity(binding))
	}
	if len(records) == 0 {
		return nil
	}

	body, statusCode, err = s.request("POST", "/device_inbounds", records, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create device inbounds failed, status=%d body=%s", statusCode, string(body))
	}
	return nil
}

func (s *SupabaseStorage) ListDeviceInbounds(deviceCode string) ([]*entity.DeviceInbound, error) {
	path := "/device_inbounds?device_code=eq." + url.QueryEscape(deviceCode) + "&select=*&order=sort.asc"
	body, statusCode, err := s.request("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list device inbounds failed, status=%d body=%s", statusCode, string(body))
	}
	var sdis []supabaseDeviceInbound
	if err := json.Unmarshal(body, &sdis); err != nil {
		return nil, fmt.Errorf("unmarshal device inbounds: %w", err)
	}
	result := make([]*entity.DeviceInbound, len(sdis))
	for i := range sdis {
		result[i] = sdis[i].toEntity()
	}
	return result, nil
}

// --- WireGuard CRUD ---

func (s *SupabaseStorage) CreateWireGuard(item *entity.WireGuard) error {
	sw := supabaseWireGuardFromEntity(item)
	_, statusCode, err := s.request("POST", "/wire_guards", sw, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("wire guard %q already exists", item.Tag)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create wire guard failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetWireGuard(tag string) (*entity.WireGuard, error) {
	path := "/wire_guards?tag=eq." + url.QueryEscape(tag) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get wire guard failed, status=%d body=%s", statusCode, string(body))
	}
	var sw supabaseWireGuard
	if err := json.Unmarshal(body, &sw); err != nil {
		return nil, fmt.Errorf("unmarshal wire guard: %w", err)
	}
	return sw.toEntity(), nil
}

func (s *SupabaseStorage) ListWireGuards() ([]*entity.WireGuard, error) {
	body, statusCode, err := s.request("GET", "/wire_guards?select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list wire guards failed, status=%d body=%s", statusCode, string(body))
	}
	var sws []supabaseWireGuard
	if err := json.Unmarshal(body, &sws); err != nil {
		return nil, fmt.Errorf("unmarshal wire guards: %w", err)
	}
	result := make([]*entity.WireGuard, len(sws))
	for i := range sws {
		result[i] = sws[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateWireGuard(item *entity.WireGuard) error {
	sw := supabaseWireGuardFromEntity(item)
	path := "/wire_guards?tag=eq." + url.QueryEscape(item.Tag)
	body, statusCode, err := s.request("PATCH", path, sw, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update wire guard failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteWireGuard(tag string) error {
	// 先解除设备绑定并删除 peer，再删模板本身，避免残留悬挂引用。
	devicePath := "/devices?wire_guard_tag=eq." + url.QueryEscape(tag)
	deviceBody := map[string]string{
		"wire_guard_tag":         "",
		"wire_guard_client_addr": "",
		"wire_guard_client_key":  "",
	}
	body, statusCode, err := s.request("PATCH", devicePath, deviceBody, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: clear wire guard devices failed, status=%d body=%s", statusCode, string(body))
	}

	peerPath := "/wire_guard_peers?wire_guard_tag=eq." + url.QueryEscape(tag)
	body, statusCode, err = s.request("DELETE", peerPath, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: clear wire guard peers failed, status=%d body=%s", statusCode, string(body))
	}

	path := "/wire_guards?tag=eq." + url.QueryEscape(tag)
	body, statusCode, err = s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete wire guard failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- WireGuardPeer CRUD ---

func (s *SupabaseStorage) CreateWireGuardPeer(item *entity.WireGuardPeer) error {
	sp := supabaseWireGuardPeerFromEntity(item)
	body, statusCode, err := s.request("POST", "/wire_guard_peers", sp, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create wire guard peer failed, status=%d body=%s", statusCode, string(body))
	}
	var peers []supabaseWireGuardPeer
	if err := json.Unmarshal(body, &peers); err == nil && len(peers) > 0 {
		item.ID = peers[0].ID
	}
	return nil
}

func (s *SupabaseStorage) ListWireGuardPeers(wireGuardTag string) ([]*entity.WireGuardPeer, error) {
	path := "/wire_guard_peers?wire_guard_tag=eq." + url.QueryEscape(wireGuardTag) + "&select=*&order=sort.asc,id.asc"
	body, statusCode, err := s.request("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list wire guard peers failed, status=%d body=%s", statusCode, string(body))
	}
	var sps []supabaseWireGuardPeer
	if err := json.Unmarshal(body, &sps); err != nil {
		return nil, fmt.Errorf("unmarshal wire guard peers: %w", err)
	}
	result := make([]*entity.WireGuardPeer, len(sps))
	for i := range sps {
		result[i] = sps[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateWireGuardPeer(item *entity.WireGuardPeer) error {
	sp := supabaseWireGuardPeerFromEntity(item)
	path := "/wire_guard_peers?id=eq." + url.QueryEscape(fmt.Sprintf("%d", item.ID))
	body, statusCode, err := s.request("PATCH", path, sp, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update wire guard peer failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteWireGuardPeer(id int64) error {
	path := "/wire_guard_peers?id=eq." + url.QueryEscape(fmt.Sprintf("%d", id))
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete wire guard peer failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// --- ExtraOutbound CRUD ---

func (s *SupabaseStorage) CreateExtraOutbound(outbound *entity.Outbound) error {
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""
	so := supabaseExtraOutboundFromEntity(outbound)
	_, statusCode, err := s.request("POST", "/outbounds", so, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode == http.StatusConflict {
		return fmt.Errorf("extra outbound %q already exists", outbound.Tag)
	}
	if statusCode != http.StatusCreated {
		return fmt.Errorf("supabase: create extra outbound failed, status=%d", statusCode)
	}
	return nil
}

func (s *SupabaseStorage) GetExtraOutbound(tag string) (*entity.Outbound, error) {
	path := "/outbounds?tag=eq." + url.QueryEscape(tag) + "&source=eq." + url.QueryEscape(string(entity.OutboundSourceManual)) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	})
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotAcceptable {
		return nil, ErrNotFound
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get extra outbound failed, status=%d body=%s", statusCode, string(body))
	}
	var so supabaseExtraOutbound
	if err := json.Unmarshal(body, &so); err != nil {
		return nil, fmt.Errorf("unmarshal extra outbound: %w", err)
	}
	return so.toEntity(), nil
}

func (s *SupabaseStorage) ListExtraOutbounds() ([]*entity.Outbound, error) {
	body, statusCode, err := s.request("GET", "/outbounds?source=eq."+url.QueryEscape(string(entity.OutboundSourceManual))+"&select=*", nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list extra outbounds failed, status=%d body=%s", statusCode, string(body))
	}
	var sos []supabaseExtraOutbound
	if err := json.Unmarshal(body, &sos); err != nil {
		return nil, fmt.Errorf("unmarshal extra outbounds: %w", err)
	}
	result := make([]*entity.Outbound, len(sos))
	for i := range sos {
		result[i] = sos[i].toEntity()
	}
	return result, nil
}

func (s *SupabaseStorage) UpdateExtraOutbound(outbound *entity.Outbound) error {
	outbound.Source = entity.OutboundSourceManual
	outbound.SubscribeName = ""
	so := supabaseExtraOutboundFromEntity(outbound)
	path := "/outbounds?tag=eq." + url.QueryEscape(outbound.Tag) + "&source=eq." + url.QueryEscape(string(entity.OutboundSourceManual))
	body, statusCode, err := s.request("PATCH", path, so, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update extra outbound failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func (s *SupabaseStorage) DeleteExtraOutbound(tag string) error {
	path := "/outbounds?tag=eq." + url.QueryEscape(tag) + "&source=eq." + url.QueryEscape(string(entity.OutboundSourceManual))
	body, statusCode, err := s.request("DELETE", path, nil, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: delete extra outbound failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// ListOutbounds 按条件查询统一 Outbound 表。
func (s *SupabaseStorage) GetOutbound(id int64) (*entity.Outbound, error) {
	path := "/outbounds?id=eq." + url.QueryEscape(strconv.FormatInt(id, 10)) + "&select=*"
	body, statusCode, err := s.request("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: get outbound failed, status=%d body=%s", statusCode, string(body))
	}

	var rows []supabaseExtraOutbound
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("unmarshal outbound: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	return rows[0].toEntity(), nil
}

// ListOutbounds 按条件查询统一 Outbound 表。
func (s *SupabaseStorage) ListOutbounds(filters ...OutboundFilter) ([]*entity.Outbound, error) {
	params := url.Values{}
	params.Set("select", "*")
	for _, filter := range filters {
		if filter.Source != nil {
			params.Set("source", "eq."+string(*filter.Source))
		}
		if filter.SubscribeName != "" {
			params.Set("subscribe_name", "eq."+filter.SubscribeName)
		}
		if filter.Enabled != nil {
			if *filter.Enabled {
				params.Set("enabled", "eq.true")
			} else {
				params.Set("enabled", "eq.false")
			}
		}
	}

	body, statusCode, err := s.request("GET", "/outbounds?"+params.Encode(), nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase: list outbounds failed, status=%d body=%s", statusCode, string(body))
	}

	var rows []supabaseExtraOutbound
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("unmarshal outbounds: %w", err)
	}
	result := make([]*entity.Outbound, len(rows))
	for i := range rows {
		result[i] = rows[i].toEntity()
	}
	return result, nil
}

// UpdateOutbound 按主键更新统一 Outbound 记录。
func (s *SupabaseStorage) UpdateOutbound(outbound *entity.Outbound) error {
	if outbound == nil || outbound.ID == 0 {
		return ErrNotFound
	}
	path := "/outbounds?id=eq." + url.QueryEscape(strconv.FormatInt(outbound.ID, 10))
	body, statusCode, err := s.request("PATCH", path, supabaseExtraOutboundFromEntity(outbound), map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update outbound failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

// DeleteOutbound 按主键删除统一 Outbound 记录。
func (s *SupabaseStorage) DeleteOutbound(id int64) error {
	path := "/outbounds?id=eq." + url.QueryEscape(strconv.FormatInt(id, 10))
	body, statusCode, err := s.request("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("supabase: delete outbound failed, status=%d body=%s", statusCode, string(body))
	}
	return nil
}

// GetOutboundsByDevice 返回某个设备可见且启用的 Outbound。
func (s *SupabaseStorage) GetOutboundsByDevice(deviceCode string) ([]*entity.Outbound, error) {
	enabled := true
	outbounds, err := s.ListOutbounds(OutboundFilter{Enabled: &enabled})
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Outbound, 0, len(outbounds))
	for _, outbound := range outbounds {
		if outbound == nil || !supabaseOutboundVisible(outbound.VisibleDevices, deviceCode) {
			continue
		}
		result = append(result, outbound)
	}
	return result, nil
}

// CreateOrUpdateOutbounds 批量写入 Outbound。
func (s *SupabaseStorage) CreateOrUpdateOutbounds(items []*entity.Outbound) error {
	rows := make([]supabaseExtraOutbound, 0, len(items))
	for _, outbound := range items {
		if outbound == nil || outbound.Tag == "" {
			continue
		}
		rows = append(rows, *supabaseExtraOutboundFromEntity(outbound))
	}
	if len(rows) == 0 {
		return nil
	}

	body, statusCode, err := s.request("POST", "/outbounds?on_conflict=tag", rows, map[string]string{
		"Prefer": "resolution=merge-duplicates,return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return fmt.Errorf("supabase: create or update outbounds failed, status=%d body=%s", statusCode, string(body))
	}
	return nil
}

// DeleteOutboundsBySubscribe 删除指定订阅源下的缓存 Outbound。
func (s *SupabaseStorage) DeleteOutboundsBySubscribe(subscribeName string) error {
	path := "/outbounds?source=eq." + url.QueryEscape(string(entity.OutboundSourceSubscription)) + "&subscribe_name=eq." + url.QueryEscape(subscribeName)
	_, statusCode, err := s.request("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return fmt.Errorf("supabase: delete subscribe outbounds failed, status=%d", statusCode)
	}
	return nil
}

// UpdateOutboundCacheTime 更新订阅的缓存时间和状态。
func (s *SupabaseStorage) UpdateOutboundCacheTime(subscribeName string, timestamp time.Time) error {
	payload := map[string]any{
		"outbound_last_fetch_time":   timestamp,
		"outbound_last_fetch_status": "SUCCESS",
		"outbound_last_fetch_error":  "",
	}
	path := "/subscribes?name=eq." + url.QueryEscape(subscribeName)
	body, statusCode, err := s.request("PATCH", path, payload, map[string]string{
		"Prefer": "return=representation",
	})
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("supabase: update outbound cache time failed, status=%d body=%s", statusCode, string(body))
	}
	if string(body) == "[]" {
		return ErrNotFound
	}
	return nil
}

func supabaseOutboundVisible(visibleDevices, deviceCode string) bool {
	if strings.TrimSpace(visibleDevices) == "" {
		return true
	}
	for _, item := range strings.Split(visibleDevices, ",") {
		if strings.TrimSpace(item) == deviceCode {
			return true
		}
	}
	return false
}
