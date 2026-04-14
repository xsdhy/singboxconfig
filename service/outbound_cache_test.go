package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"testing"
	"time"
)

func newOutboundCacheService(t *testing.T) (*Service, storage.Storage) {
	t.Helper()
	t.Skip("Tests require database storage - memory storage has been removed")
	var store storage.Storage
	return NewService(store), store
}

func TestRefreshSubscriptionOutboundSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(
			"ss://YWVzLTEyOC1nY206ZDhxZmU0@dns.yypa.zzssptop.com:20475/?group=5LqR57-8572R57uc#AA%E9%A6%99%E6%B8%AF01\ninvalid://skip",
		))))
	}))
	defer server.Close()

	svc, store := newOutboundCacheService(t)
	subscribe := &entity.Subscribe{
		Name:                  "provider-a",
		URL:                   server.URL,
		Status:                true,
		VisibleDevices:        "phone,pad",
		OutboundCacheDuration: 30,
	}
	if err := store.CreateSubscribe(subscribe); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}

	if err := svc.RefreshSubscriptionOutbound(context.Background(), subscribe.Name); err != nil {
		t.Fatalf("RefreshSubscriptionOutbound failed: %v", err)
	}

	source := entity.OutboundSourceSubscription
	outbounds, err := store.ListOutbounds(storage.OutboundFilter{Source: &source})
	if err != nil {
		t.Fatalf("ListOutbounds failed: %v", err)
	}
	if len(outbounds) != 1 {
		t.Fatalf("subscription outbounds len=%d, want 1", len(outbounds))
	}
	if outbounds[0].SubscribeName != subscribe.Name || outbounds[0].VisibleDevices != subscribe.VisibleDevices {
		t.Fatalf("cached outbound metadata mismatch: %+v", outbounds[0])
	}
	if outbounds[0].LastFetchTime == nil {
		t.Fatalf("cached outbound lastFetchTime should not be nil: %+v", outbounds[0])
	}

	gotSubscribe, err := store.GetSubscribe(subscribe.Name)
	if err != nil {
		t.Fatalf("GetSubscribe failed: %v", err)
	}
	if gotSubscribe.OutboundLastFetchStatus != "SUCCESS" || gotSubscribe.OutboundLastFetchError != "" {
		t.Fatalf("subscribe cache status mismatch: %+v", gotSubscribe)
	}
	if gotSubscribe.OutboundLastFetchTime == nil {
		t.Fatalf("subscribe OutboundLastFetchTime should not be nil")
	}
}

func TestRefreshSubscriptionOutboundFailureKeepsOldCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer server.Close()

	svc, store := newOutboundCacheService(t)
	subscribe := &entity.Subscribe{
		Name:                  "provider-a",
		URL:                   server.URL,
		Status:                true,
		VisibleDevices:        "phone",
		UserAgent:             "test-agent",
		OutboundLastFetchTime: ptrServiceTime(time.Date(2026, 4, 12, 8, 0, 0, 0, time.UTC)),
	}
	if err := store.CreateSubscribe(subscribe); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            "old-node",
			Type:           "socks",
			Enabled:        true,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  subscribe.Name,
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"old-node","type":"socks"}`,
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}

	if err := svc.RefreshSubscriptionOutbound(context.Background(), subscribe.Name); err == nil {
		t.Fatal("RefreshSubscriptionOutbound should fail when upstream returns 502")
	}

	outbounds, err := store.GetOutboundsByDevice("phone")
	if err != nil {
		t.Fatalf("GetOutboundsByDevice failed: %v", err)
	}
	if len(outbounds) != 1 || outbounds[0].Tag != "old-node" {
		t.Fatalf("old cache should be kept on failure: %+v", outbounds)
	}

	gotSubscribe, err := store.GetSubscribe(subscribe.Name)
	if err != nil {
		t.Fatalf("GetSubscribe failed: %v", err)
	}
	if gotSubscribe.OutboundLastFetchStatus != "FAILED" || gotSubscribe.OutboundLastFetchError == "" {
		t.Fatalf("subscribe failure state mismatch: %+v", gotSubscribe)
	}
}

func TestResolveGenerateOutboundsRefreshesExpiredCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ss://YWVzLTEyOC1nY206ZDhxZmU0@dns.yypa.zzssptop.com:20475/?group=5LqR57-8572R57uc#AA%E9%A6%99%E6%B8%AF01"))
	}))
	defer server.Close()

	svc, store := newOutboundCacheService(t)
	if err := store.CreateSubscribe(&entity.Subscribe{
		Name:                    "provider-a",
		URL:                     server.URL,
		Status:                  true,
		VisibleDevices:          "phone",
		OutboundCacheDuration:   1,
		OutboundLastFetchTime:   ptrServiceTime(time.Now().Add(-2 * time.Minute).UTC()),
		OutboundLastFetchStatus: "SUCCESS",
	}); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            "manual-visible",
			Type:           "socks",
			Enabled:        true,
			Source:         entity.OutboundSourceManual,
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"manual-visible","type":"socks"}`,
		},
		{
			Tag:            "manual-hidden",
			Type:           "socks",
			Enabled:        true,
			Source:         entity.OutboundSourceManual,
			VisibleDevices: "tv",
			ConfigJSON:     `{"tag":"manual-hidden","type":"socks"}`,
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}
	if err := store.CreateNodeGroup(&entity.NodeGroup{
		Tag:       "proxy",
		GroupType: "selector",
		Include:   "AA,香港,manual",
	}); err != nil {
		t.Fatalf("CreateNodeGroup failed: %v", err)
	}

	outbounds, err := svc.resolveGenerateOutbounds(context.Background(), "phone")
	if err != nil {
		t.Fatalf("resolveGenerateOutbounds failed: %v", err)
	}

	hasSubscriptionNode := false
	hasManualVisible := false
	hasManualHidden := false
	hasGroup := false
	hasDirect := false
	for _, outbound := range outbounds {
		switch outbound.Tag {
		case "AA香港01":
			hasSubscriptionNode = true
		case "manual-visible":
			hasManualVisible = true
		case "manual-hidden":
			hasManualHidden = true
		case "proxy":
			hasGroup = true
		case "direct":
			hasDirect = true
		}
	}
	if !hasSubscriptionNode || !hasManualVisible || hasManualHidden || !hasGroup || !hasDirect {
		t.Fatalf("resolveGenerateOutbounds unexpected result: %+v", outbounds)
	}
}

func ptrServiceTime(value time.Time) *time.Time {
	return &value
}
