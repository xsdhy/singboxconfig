package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"singboxconfig/convert/singbox"
	"singboxconfig/entity"
	"singboxconfig/protocol"
	"singboxconfig/storage"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type subscriptionURLScheme string

const (
	// subscriptionURLSchemeSS 表示订阅中的 Shadowsocks 节点链接协议头。
	subscriptionURLSchemeSS subscriptionURLScheme = "ss"
	// subscriptionURLSchemeSSR 表示订阅中的 ShadowsocksR 节点链接协议头。
	subscriptionURLSchemeSSR subscriptionURLScheme = "ssr"
	// subscriptionURLSchemeTrojan 表示订阅中的 Trojan 节点链接协议头。
	subscriptionURLSchemeTrojan subscriptionURLScheme = "trojan"
	// subscriptionURLSchemeVMess 表示订阅中的 VMess 节点链接协议头。
	subscriptionURLSchemeVMess subscriptionURLScheme = "vmess"
	// subscriptionURLSchemeVLESS 表示订阅中的 VLESS 节点链接协议头。
	subscriptionURLSchemeVLESS subscriptionURLScheme = "vless"
)

var subscriptionOutboundConvertMap = map[subscriptionURLScheme]func(string) (*entity.SingBoxOut, error){
	subscriptionURLSchemeSS:     protocol.DecodeSSURLToSingBox,
	subscriptionURLSchemeSSR:    protocol.DecodeSSRURLToSingBox,
	subscriptionURLSchemeTrojan: protocol.DecodeTrojanUrlToSingBox,
	subscriptionURLSchemeVMess:  protocol.DecodeVmessUrlToSingBox,
	subscriptionURLSchemeVLESS:  protocol.DecodeVlessUrlToSingBox,
}

// subscriptionRefreshResult 记录一次手动刷新前后的差异，供 API 直接返回。
type subscriptionRefreshResult struct {
	Added         int       `json:"added"`
	Updated       int       `json:"updated"`
	Deleted       int       `json:"deleted"`
	LastFetchTime time.Time `json:"lastFetchTime"`
}

// needsRefresh 判断订阅 Outbound 缓存是否需要刷新。
// 为空表示从未刷新过；缓存时间小于等于 0 时视为每次生成都尝试刷新。
func needsRefresh(subscribe *entity.Subscribe) bool {
	if subscribe == nil || !subscribe.Status {
		return false
	}
	if subscribe.OutboundLastFetchTime == nil {
		return true
	}
	if subscribe.OutboundCacheDuration <= 0 {
		return true
	}
	expireAt := subscribe.OutboundLastFetchTime.Add(time.Duration(subscribe.OutboundCacheDuration) * time.Minute)
	return time.Now().After(expireAt)
}

// isDeviceVisible 统一处理逗号分隔的设备可见性规则。
func isDeviceVisible(deviceCode, visibleDevices string) bool {
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

// RefreshSubscriptionOutbound 拉取指定订阅并覆盖写入缓存表。
// 失败时仅更新订阅的失败状态，保留旧缓存供生成流程继续使用。
func (s *Service) RefreshSubscriptionOutbound(ctx context.Context, subscribeName string) error {
	_, err := s.refreshSubscriptionOutboundDetailed(ctx, subscribeName)
	return err
}

// refreshSubscriptionOutboundDetailed 在刷新缓存的同时返回新增/更新/删除统计。
func (s *Service) refreshSubscriptionOutboundDetailed(ctx context.Context, subscribeName string) (*subscriptionRefreshResult, error) {
	subscribe, err := s.storage.GetSubscribe(subscribeName)
	if err != nil {
		return nil, err
	}
	if !subscribe.Status {
		return &subscriptionRefreshResult{}, nil
	}
	existingItems, err := s.storage.ListOutbounds(
		storage.OutboundFilter{Source: outboundSourcePtr(entity.OutboundSourceSubscription)},
		storage.OutboundFilter{SubscribeName: subscribe.Name},
	)
	if err != nil {
		return nil, err
	}

	outbounds, fetchTime, err := s.fetchSubscriptionOutbounds(ctx, subscribe)
	if err != nil {
		s.markSubscribeOutboundRefreshFailed(subscribe, err)
		return nil, err
	}
	result := calculateSubscriptionRefreshResult(existingItems, outbounds, fetchTime)

	if err := s.storage.DeleteOutboundsBySubscribe(subscribe.Name); err != nil {
		s.markSubscribeOutboundRefreshFailed(subscribe, err)
		return nil, err
	}
	if err := s.storage.CreateOrUpdateOutbounds(outbounds); err != nil {
		s.markSubscribeOutboundRefreshFailed(subscribe, err)
		return nil, err
	}
	if err := s.storage.UpdateOutboundCacheTime(subscribe.Name, fetchTime); err != nil {
		s.markSubscribeOutboundRefreshFailed(subscribe, err)
		return nil, err
	}
	return result, nil
}

// resolveGenerateOutbounds 在生成配置前处理订阅缓存刷新，并返回最终可用出站列表。
func (s *Service) resolveGenerateOutbounds(ctx context.Context, deviceCode string) ([]entity.SingBoxOut, error) {
	subscribes, err := s.storage.ListSubscribes()
	if err != nil {
		return nil, err
	}

	// 收集禁用或对当前设备不可见的订阅名称，用于后续过滤其缓存 Outbound。
	disabledSubscribes := make(map[string]bool, len(subscribes))
	for _, subscribe := range subscribes {
		if subscribe == nil {
			continue
		}
		if !subscribe.Status || !isDeviceVisible(deviceCode, subscribe.VisibleDevices) {
			disabledSubscribes[subscribe.Name] = true
			continue
		}
		if needsRefresh(subscribe) {
			if err := s.RefreshSubscriptionOutbound(ctx, subscribe.Name); err != nil {
				logrus.WithError(err).WithField("subscribe", subscribe.Name).Warn("resolveGenerateOutbounds: refresh failed, fallback to old cache")
			}
		}
	}

	items, err := s.storage.GetOutboundsByDevice(deviceCode)
	if err != nil {
		return nil, err
	}

	// 过滤掉属于禁用/不可见订阅的缓存 Outbound。
	if len(disabledSubscribes) > 0 {
		filtered := make([]*entity.Outbound, 0, len(items))
		for _, item := range items {
			if item != nil && item.Source == entity.OutboundSourceSubscription && disabledSubscribes[item.SubscribeName] {
				continue
			}
			filtered = append(filtered, item)
		}
		items = filtered
	}

	// 兼容历史默认行为：当存储中完全没有任何 Outbound 时，仍保留默认手工节点兜底。
	if len(items) == 0 {
		items = singbox.GetDefaultExtraOutbounds()
	}

	groupRules, err := s.storage.ListNodeGroups()
	if err != nil {
		return nil, err
	}
	return singbox.GetOutbounds(singbox.GetExtraOutbounds(deviceCode, items), groupRules, deviceCode), nil
}

// resolveManualGenerateOutbounds 返回设备可见且启用的手工 Outbound（不含订阅缓存节点）。
// Surge 输出链路使用：订阅节点改由 Surge 通过 policy-path 自行拉取订阅地址，
// 因此这里不触发订阅缓存刷新，也不把订阅缓存节点展开进配置。
func (s *Service) resolveManualGenerateOutbounds(deviceCode string) ([]entity.SingBoxOut, error) {
	items, err := s.storage.GetOutboundsByDevice(deviceCode)
	if err != nil {
		return nil, err
	}

	manualItems := make([]*entity.Outbound, 0, len(items))
	for _, item := range items {
		if item == nil || item.Source == entity.OutboundSourceSubscription {
			continue
		}
		manualItems = append(manualItems, item)
	}

	// 兼容历史默认行为：当存储中完全没有任何 Outbound 时，仍保留默认手工节点兜底。
	if len(items) == 0 {
		manualItems = singbox.GetDefaultExtraOutbounds()
	}

	return singbox.GetExtraOutbounds(deviceCode, manualItems), nil
}

// resolveVisibleSubscribes 返回启用且对当前设备可见的订阅源，供 Surge 输出为 policy-path 策略组。
func (s *Service) resolveVisibleSubscribes(deviceCode string) ([]*entity.Subscribe, error) {
	subscribes, err := s.storage.ListSubscribes()
	if err != nil {
		return nil, err
	}
	result := make([]*entity.Subscribe, 0, len(subscribes))
	for _, subscribe := range subscribes {
		if subscribe == nil || !subscribe.Status || !isDeviceVisible(deviceCode, subscribe.VisibleDevices) {
			continue
		}
		result = append(result, subscribe)
	}
	return result, nil
}

func (s *Service) fetchSubscriptionOutbounds(ctx context.Context, subscribe *entity.Subscribe) ([]*entity.Outbound, time.Time, error) {
	body, err := s.httpGetBytes(ctx, subscribe.URL, subscribe.UserAgent)
	if err != nil {
		return nil, time.Time{}, err
	}

	nodes, err := parseSubscriptionOutbounds(body)
	if err != nil {
		return nil, time.Time{}, err
	}

	fetchTime := time.Now().UTC()
	result := make([]*entity.Outbound, 0, len(nodes))
	for index, node := range nodes {
		configJSON, err := json.Marshal(node)
		if err != nil {
			logrus.WithError(err).WithField("tag", node.Tag).Warn("fetchSubscriptionOutbounds: marshal outbound failed, skip")
			continue
		}
		result = append(result, &entity.Outbound{
			Tag:            node.Tag,
			Name:           node.Tag,
			Type:           node.Type,
			Enabled:        true,
			Sort:           index,
			VisibleDevices: subscribe.VisibleDevices,
			ConfigJSON:     string(configJSON),
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  subscribe.Name,
			LastFetchTime:  &fetchTime,
		})
	}
	return result, fetchTime, nil
}

func parseSubscriptionOutbounds(httpContent []byte) ([]entity.SingBoxOut, error) {
	body, err := base64.StdEncoding.DecodeString(string(httpContent))
	if err != nil {
		body = httpContent
	}

	lines := strings.Split(string(body), "\n")
	outbounds := make([]entity.SingBoxOut, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 跳过状态信息行（如 STATUS=...）
		if strings.HasPrefix(line, "STATUS=") || strings.HasPrefix(line, "status=") {
			continue
		}
		// 必须包含 :// 才可能是节点链接
		if !strings.Contains(line, "://") {
			continue
		}
		protocolName := subscriptionURLScheme(strings.SplitN(line, "://", 2)[0])
		decodeFunc, ok := subscriptionOutboundConvertMap[protocolName]
		if !ok {
			continue
		}
		node, err := decodeFunc(line)
		if err != nil {
			logrus.WithError(err).WithField("line", line).Warn("parseSubscriptionOutbounds: decode failed, skip")
			continue
		}
		outbounds = append(outbounds, *node)
	}
	return outbounds, nil
}

func (s *Service) httpGetBytes(ctx context.Context, url string, userAgent string) ([]byte, error) {
	// 创建独立的超时 context，避免受上层 HTTP 请求取消影响
	fetchCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(fetchCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("httpGetBytes: failed to create request: %w", err)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpGetBytes: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("httpGetBytes: unexpected HTTP status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpGetBytes: failed to read response body: %w", err)
	}
	fmt.Printf("httpGetBytes: fetched %d bytes from %s\n", len(body), url)
	return body, nil
}

func (s *Service) markSubscribeOutboundRefreshFailed(subscribe *entity.Subscribe, refreshErr error) {
	if subscribe == nil {
		return
	}
	subscribe.OutboundLastFetchStatus = "FAILED"
	subscribe.OutboundLastFetchError = refreshErr.Error()
	if subscribe.OutboundLastFetchTime == nil {
		now := time.Now().UTC()
		subscribe.OutboundLastFetchTime = &now
	}
	if err := s.storage.UpdateSubscribe(subscribe); err != nil {
		logrus.WithError(err).WithField("subscribe", subscribe.Name).Warn("markSubscribeOutboundRefreshFailed: update subscribe failed")
	}
}

func calculateSubscriptionRefreshResult(previous []*entity.Outbound, current []*entity.Outbound, fetchTime time.Time) *subscriptionRefreshResult {
	previousByTag := make(map[string]*entity.Outbound, len(previous))
	for _, item := range previous {
		if item == nil || item.Tag == "" {
			continue
		}
		previousByTag[item.Tag] = item
	}

	currentByTag := make(map[string]*entity.Outbound, len(current))
	result := &subscriptionRefreshResult{LastFetchTime: fetchTime}
	for _, item := range current {
		if item == nil || item.Tag == "" {
			continue
		}
		currentByTag[item.Tag] = item
		if oldItem, ok := previousByTag[item.Tag]; ok {
			if !subscriptionOutboundEqual(oldItem, item) {
				result.Updated++
			}
			continue
		}
		result.Added++
	}
	for _, item := range previous {
		if item == nil || item.Tag == "" {
			continue
		}
		if _, ok := currentByTag[item.Tag]; !ok {
			result.Deleted++
		}
	}
	return result
}

// subscriptionOutboundEqual 只比较会影响缓存展示与配置生成的关键字段。
func subscriptionOutboundEqual(left *entity.Outbound, right *entity.Outbound) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Tag == right.Tag &&
		left.Name == right.Name &&
		left.Description == right.Description &&
		left.Type == right.Type &&
		left.Enabled == right.Enabled &&
		left.Sort == right.Sort &&
		left.VisibleDevices == right.VisibleDevices &&
		left.ConfigJSON == right.ConfigJSON &&
		left.Source == right.Source &&
		left.SubscribeName == right.SubscribeName
}

func outboundSourcePtr(source entity.OutboundSource) *entity.OutboundSource {
	return &source
}
