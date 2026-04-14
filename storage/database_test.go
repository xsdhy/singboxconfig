package storage

import (
	"errors"
	"singboxconfig/entity"
	"testing"
	"time"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *DatabaseStorage {
	// DatabaseStorage tests require a real database connection
	t.Skip("DatabaseStorage tests require a real database connection (PostgreSQL or MySQL)")
	return nil
}

func TestDatabaseStorage_Subscribe(t *testing.T) {
	storage := setupTestDB(t)

	subscribe := &entity.Subscribe{
		Name:                    "test-provider",
		URL:                     "https://example.com/sub",
		UserAgent:               "test-agent",
		Status:                  true,
		VisibleDevices:          "phone,pad",
		OutboundLastFetchTime:   ptrTime(time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)),
		OutboundCacheDuration:   60,
		OutboundLastFetchStatus: "SUCCESS",
		OutboundLastFetchError:  "",
	}

	// Test Create
	err := storage.CreateSubscribe(subscribe)
	if err != nil {
		t.Errorf("CreateSubscribe failed: %v", err)
	}

	// Test Get
	got, err := storage.GetSubscribe(subscribe.Name)
	if err != nil {
		t.Errorf("GetSubscribe failed: %v", err)
	}
	if got.Name != subscribe.Name || got.URL != subscribe.URL {
		t.Errorf("GetSubscribe = %v, want %v", got, subscribe)
	}
	if got.VisibleDevices != subscribe.VisibleDevices || got.OutboundCacheDuration != subscribe.OutboundCacheDuration || got.OutboundLastFetchStatus != subscribe.OutboundLastFetchStatus {
		t.Fatalf("GetSubscribe new fields mismatch: got=%+v want=%+v", got, subscribe)
	}

	// Test List
	list, err := storage.ListSubscribes()
	if err != nil {
		t.Errorf("ListSubscribes failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListSubscribes count = %d, want 1", len(list))
	}

	// Test Update
	subscribe.URL = "https://example.com/new-sub"
	err = storage.UpdateSubscribe(subscribe)
	if err != nil {
		t.Errorf("UpdateSubscribe failed: %v", err)
	}

	got, err = storage.GetSubscribe(subscribe.Name)
	if err != nil {
		t.Errorf("GetSubscribe after update failed: %v", err)
	}
	if got.URL != subscribe.URL {
		t.Errorf("GetSubscribe after update URL = %s, want %s", got.URL, subscribe.URL)
	}

	// Test Delete
	err = storage.DeleteSubscribe(subscribe.Name)
	if err != nil {
		t.Errorf("DeleteSubscribe failed: %v", err)
	}

	got, err = storage.GetSubscribe(subscribe.Name)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetSubscribe after delete should return ErrNotFound, got: %v", err)
	}
	if got != nil {
		t.Errorf("GetSubscribe after delete = %v, want nil", got)
	}
}

func TestDatabaseStorage_NodeGroup(t *testing.T) {
	storage := setupTestDB(t)

	group := &entity.NodeGroup{
		Name:      "Test Group",
		Tag:       "test-group",
		GroupType: "selector",
		TestURL:   "https://www.google.com",
		Include:   "HK,TW",
		Exclude:   "expired",
	}

	// Test Create
	err := storage.CreateNodeGroup(group)
	if err != nil {
		t.Errorf("CreateNodeGroup failed: %v", err)
	}

	// Test Get
	got, err := storage.GetNodeGroup(group.Tag)
	if err != nil {
		t.Errorf("GetNodeGroup failed: %v", err)
	}
	if got.Tag != group.Tag || got.Name != group.Name {
		t.Errorf("GetNodeGroup = %v, want %v", got, group)
	}

	// Test List
	list, err := storage.ListNodeGroups()
	if err != nil {
		t.Errorf("ListNodeGroups failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListNodeGroups count = %d, want 1", len(list))
	}

	// Test Update
	group.Name = "Updated Group"
	err = storage.UpdateNodeGroup(group)
	if err != nil {
		t.Errorf("UpdateNodeGroup failed: %v", err)
	}

	got, err = storage.GetNodeGroup(group.Tag)
	if err != nil {
		t.Errorf("GetNodeGroup after update failed: %v", err)
	}
	if got.Name != group.Name {
		t.Errorf("GetNodeGroup after update Name = %s, want %s", got.Name, group.Name)
	}

	// Test Delete
	err = storage.DeleteNodeGroup(group.Tag)
	if err != nil {
		t.Errorf("DeleteNodeGroup failed: %v", err)
	}

	got, err = storage.GetNodeGroup(group.Tag)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetNodeGroup after delete should return ErrNotFound, got: %v", err)
	}
	if got != nil {
		t.Errorf("GetNodeGroup after delete = %v, want nil", got)
	}
}

func TestDatabaseStorage_RuleSet(t *testing.T) {
	storage := setupTestDB(t)

	ruleSet := &entity.RuleSet{
		Name:           "Test RuleSet",
		Tag:            "test-ruleset",
		RuleSetType:    "remote",
		Format:         "binary",
		URL:            "https://example.com/rules.srs",
		Outbound:       "proxy",
		DownloadDetour: "direct",
		AbleDevices:    "ios,android",
		Sort:           1,
	}

	// Test Create
	err := storage.CreateRuleSet(ruleSet)
	if err != nil {
		t.Errorf("CreateRuleSet failed: %v", err)
	}

	// Test Get
	got, err := storage.GetRuleSet(ruleSet.Tag)
	if err != nil {
		t.Errorf("GetRuleSet failed: %v", err)
	}
	if got.Tag != ruleSet.Tag || got.Name != ruleSet.Name {
		t.Errorf("GetRuleSet = %v, want %v", got, ruleSet)
	}

	// Test List
	list, err := storage.ListRuleSets()
	if err != nil {
		t.Errorf("ListRuleSets failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListRuleSets count = %d, want 1", len(list))
	}

	// Test Update
	ruleSet.Name = "Updated RuleSet"
	err = storage.UpdateRuleSet(ruleSet)
	if err != nil {
		t.Errorf("UpdateRuleSet failed: %v", err)
	}

	got, err = storage.GetRuleSet(ruleSet.Tag)
	if err != nil {
		t.Errorf("GetRuleSet after update failed: %v", err)
	}
	if got.Name != ruleSet.Name {
		t.Errorf("GetRuleSet after update Name = %s, want %s", got.Name, ruleSet.Name)
	}

	// Test Delete
	err = storage.DeleteRuleSet(ruleSet.Tag)
	if err != nil {
		t.Errorf("DeleteRuleSet failed: %v", err)
	}

	got, err = storage.GetRuleSet(ruleSet.Tag)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetRuleSet after delete should return ErrNotFound, got: %v", err)
	}
	if got != nil {
		t.Errorf("GetRuleSet after delete = %v, want nil", got)
	}
}

func TestDatabaseStorage_GlobalSetting(t *testing.T) {
	storage := setupTestDB(t)

	key := "test-key"
	value := "test-value"

	// Test Set (Create)
	err := storage.SetGlobalSetting(key, value)
	if err != nil {
		t.Errorf("SetGlobalSetting failed: %v", err)
	}

	// Test Get
	got, err := storage.GetGlobalSetting(key)
	if err != nil {
		t.Errorf("GetGlobalSetting failed: %v", err)
	}
	if got != value {
		t.Errorf("GetGlobalSetting = %s, want %s", got, value)
	}

	// Test List
	list, err := storage.ListGlobalSettings()
	if err != nil {
		t.Errorf("ListGlobalSettings failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListGlobalSettings count = %d, want 1", len(list))
	}
	if list[key] != value {
		t.Errorf("ListGlobalSettings[%s] = %s, want %s", key, list[key], value)
	}

	// Test Set (Update)
	newValue := "new-value"
	err = storage.SetGlobalSetting(key, newValue)
	if err != nil {
		t.Errorf("SetGlobalSetting (update) failed: %v", err)
	}

	got, err = storage.GetGlobalSetting(key)
	if err != nil {
		t.Errorf("GetGlobalSetting after update failed: %v", err)
	}
	if got != newValue {
		t.Errorf("GetGlobalSetting after update = %s, want %s", got, newValue)
	}

	// Test Delete
	err = storage.DeleteGlobalSetting(key)
	if err != nil {
		t.Errorf("DeleteGlobalSetting failed: %v", err)
	}

	got, err = storage.GetGlobalSetting(key)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetGlobalSetting after delete should return ErrNotFound, got: %v", err)
	}
	if got != "" {
		t.Errorf("GetGlobalSetting after delete = %s, want empty string", got)
	}
}

func TestDatabaseStorage_DeviceManagementEntities(t *testing.T) {
	storage := setupTestDB(t)

	device := &entity.Device{
		Code:                "phone",
		Name:                "Phone",
		Token:               "token",
		Enabled:             true,
		WireGuardTag:        "wg-main",
		WireGuardClientAddr: "10.8.0.4",
		WireGuardClientKey:  "private-key",
	}
	if err := storage.CreateDevice(device); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	gotDevice, err := storage.GetDevice(device.Code)
	if err != nil || gotDevice.Token != "token" {
		t.Fatalf("GetDevice failed: device=%+v err=%v", gotDevice, err)
	}

	inbound := &entity.Inbound{Tag: "tun", Type: "tun", Enabled: true, ConfigJSON: `{"type":"tun"}`}
	if err := storage.CreateInbound(inbound); err != nil {
		t.Fatalf("CreateInbound failed: %v", err)
	}
	if err := storage.SetDeviceInbounds(device.Code, []*entity.DeviceInbound{
		{DeviceCode: device.Code, InboundTag: inbound.Tag, Sort: 1},
	}); err != nil {
		t.Fatalf("SetDeviceInbounds failed: %v", err)
	}
	bindings, err := storage.ListDeviceInbounds(device.Code)
	if err != nil || len(bindings) != 1 || bindings[0].InboundTag != inbound.Tag {
		t.Fatalf("ListDeviceInbounds failed: bindings=%+v err=%v", bindings, err)
	}

	wg := &entity.WireGuard{Tag: "wg-main", EndpointTag: "wg-ep", Enabled: true, MTU: 1408}
	if err := storage.CreateWireGuard(wg); err != nil {
		t.Fatalf("CreateWireGuard failed: %v", err)
	}
	peer := &entity.WireGuardPeer{
		WireGuardTag:                wg.Tag,
		Address:                     "1.1.1.1",
		Port:                        51820,
		PublicKey:                   "pub",
		AllowedIPs:                  "0.0.0.0/0",
		PersistentKeepaliveInterval: 25,
		Enabled:                     true,
	}
	if err := storage.CreateWireGuardPeer(peer); err != nil {
		t.Fatalf("CreateWireGuardPeer failed: %v", err)
	}
	if peer.ID == 0 {
		t.Fatal("CreateWireGuardPeer did not assign ID")
	}
	peers, err := storage.ListWireGuardPeers(wg.Tag)
	if err != nil || len(peers) != 1 || peers[0].ID != peer.ID {
		t.Fatalf("ListWireGuardPeers failed: peers=%+v err=%v", peers, err)
	}
	peer.Port = 51821
	if err := storage.UpdateWireGuardPeer(peer); err != nil {
		t.Fatalf("UpdateWireGuardPeer failed: %v", err)
	}

	fetchTime := time.Date(2026, 4, 12, 11, 0, 0, 0, time.UTC)
	outbound := &entity.Outbound{
		Tag:           "selfsip",
		Type:          "vmess",
		Enabled:       true,
		ConfigJSON:    `{"tag":"selfsip"}`,
		Source:        entity.OutboundSourceManual,
		LastFetchTime: &fetchTime,
	}
	if err := storage.CreateExtraOutbound(outbound); err != nil {
		t.Fatalf("CreateExtraOutbound failed: %v", err)
	}
	gotOutbound, err := storage.GetExtraOutbound(outbound.Tag)
	if err != nil || gotOutbound.Type != "vmess" {
		t.Fatalf("GetExtraOutbound failed: outbound=%+v err=%v", gotOutbound, err)
	}
	if gotOutbound.Source != entity.OutboundSourceManual || gotOutbound.SubscribeName != "" || gotOutbound.LastFetchTime == nil || !gotOutbound.LastFetchTime.Equal(fetchTime) {
		t.Fatalf("GetExtraOutbound metadata mismatch: %+v", gotOutbound)
	}

	if err := storage.DeleteWireGuardPeer(peer.ID); err != nil {
		t.Fatalf("DeleteWireGuardPeer failed: %v", err)
	}
	if err := storage.DeleteWireGuard(wg.Tag); err != nil {
		t.Fatalf("DeleteWireGuard failed: %v", err)
	}
	gotDevice, err = storage.GetDevice(device.Code)
	if err != nil {
		t.Fatalf("GetDevice after DeleteWireGuard failed: %v", err)
	}
	if gotDevice.WireGuardTag != "" || gotDevice.WireGuardClientAddr != "" || gotDevice.WireGuardClientKey != "" {
		t.Fatalf("DeleteWireGuard should clear device wireguard fields: %+v", gotDevice)
	}
	if err := storage.DeleteInbound(inbound.Tag); err != nil {
		t.Fatalf("DeleteInbound failed: %v", err)
	}
	bindings, err = storage.ListDeviceInbounds(device.Code)
	if err != nil || len(bindings) != 0 {
		t.Fatalf("DeleteInbound should clear bindings: bindings=%+v err=%v", bindings, err)
	}
	if err := storage.DeleteExtraOutbound(outbound.Tag); err != nil {
		t.Fatalf("DeleteExtraOutbound failed: %v", err)
	}
	if err := storage.DeleteDevice(device.Code); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}
	if _, err := storage.GetDevice(device.Code); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetDevice after delete should return ErrNotFound, got %v", err)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func TestDatabaseStorage_OutboundRepository(t *testing.T) {
	storage := setupTestDB(t)
	subscribe := &entity.Subscribe{
		Name:                  "sub-a",
		URL:                   "https://example.com/sub-a",
		Status:                true,
		OutboundCacheDuration: 30,
	}
	if err := storage.CreateSubscribe(subscribe); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}

	now := time.Date(2026, 4, 13, 8, 0, 0, 0, time.UTC)
	items := []*entity.Outbound{
		{
			Tag:            "manual-visible",
			Type:           "socks",
			Enabled:        true,
			Source:         entity.OutboundSourceManual,
			VisibleDevices: "phone,pad",
			ConfigJSON:     `{"tag":"manual-visible","type":"socks"}`,
		},
		{
			Tag:            "sub-visible",
			Type:           "vmess",
			Enabled:        true,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "sub-a",
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"sub-visible","type":"vmess"}`,
			LastFetchTime:  ptrTime(now),
		},
		{
			Tag:            "sub-hidden",
			Type:           "vmess",
			Enabled:        false,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "sub-a",
			VisibleDevices: "tv",
			ConfigJSON:     `{"tag":"sub-hidden","type":"vmess"}`,
		},
	}
	if err := storage.CreateOrUpdateOutbounds(items); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}
	allItems, err := storage.ListOutbounds()
	if err != nil {
		t.Fatalf("ListOutbounds(all) failed: %v", err)
	}
	var manualItem *entity.Outbound
	for _, item := range allItems {
		if item != nil && item.Tag == "manual-visible" {
			manualItem = item
			break
		}
	}
	if manualItem == nil || manualItem.ID == 0 {
		t.Fatalf("manual-visible outbound missing id: %+v", allItems)
	}
	gotByID, err := storage.GetOutbound(manualItem.ID)
	if err != nil || gotByID.Tag != "manual-visible" {
		t.Fatalf("GetOutbound(%d) = %+v, %v", manualItem.ID, gotByID, err)
	}
	gotByID.Enabled = false
	gotByID.Tag = "manual-visible-updated"
	if err := storage.UpdateOutbound(gotByID); err != nil {
		t.Fatalf("UpdateOutbound failed: %v", err)
	}
	updated, err := storage.GetOutbound(manualItem.ID)
	if err != nil || updated.Tag != "manual-visible-updated" || updated.Enabled {
		t.Fatalf("updated outbound unexpected: %+v, %v", updated, err)
	}
	if err := storage.DeleteOutbound(manualItem.ID); err != nil {
		t.Fatalf("DeleteOutbound failed: %v", err)
	}
	if _, err := storage.GetOutbound(manualItem.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetOutbound after delete err=%v, want ErrNotFound", err)
	}
	updated.Enabled = true
	if err := storage.CreateOrUpdateOutbounds([]*entity.Outbound{updated}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds(recreate) failed: %v", err)
	}

	source := entity.OutboundSourceSubscription
	enabled := true
	filtered, err := storage.ListOutbounds(
		OutboundFilter{Source: &source},
		OutboundFilter{SubscribeName: "sub-a"},
		OutboundFilter{Enabled: &enabled},
	)
	if err != nil {
		t.Fatalf("ListOutbounds failed: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Tag != "sub-visible" {
		t.Fatalf("ListOutbounds filtered result unexpected: %+v", filtered)
	}

	visible, err := storage.GetOutboundsByDevice("phone")
	if err != nil {
		t.Fatalf("GetOutboundsByDevice failed: %v", err)
	}
	if len(visible) != 2 {
		t.Fatalf("GetOutboundsByDevice(phone) len=%d, want 2, items=%+v", len(visible), visible)
	}

	if err := storage.DeleteOutboundsBySubscribe("sub-a"); err != nil {
		t.Fatalf("DeleteOutboundsBySubscribe failed: %v", err)
	}
	remaining, err := storage.ListOutbounds()
	if err != nil {
		t.Fatalf("ListOutbounds after delete failed: %v", err)
	}
	if len(remaining) != 1 || remaining[0].Tag != "manual-visible-updated" {
		t.Fatalf("remaining outbounds unexpected: %+v", remaining)
	}

	if err := storage.UpdateOutboundCacheTime("sub-a", now); err != nil {
		t.Fatalf("UpdateOutboundCacheTime failed: %v", err)
	}
	gotSubscribe, err := storage.GetSubscribe("sub-a")
	if err != nil {
		t.Fatalf("GetSubscribe failed: %v", err)
	}
	if gotSubscribe.OutboundLastFetchTime == nil || !gotSubscribe.OutboundLastFetchTime.Equal(now) {
		t.Fatalf("OutboundLastFetchTime mismatch: %+v", gotSubscribe.OutboundLastFetchTime)
	}
	if gotSubscribe.OutboundLastFetchStatus != "SUCCESS" || gotSubscribe.OutboundLastFetchError != "" {
		t.Fatalf("Outbound cache status mismatch: %+v", gotSubscribe)
	}
}
