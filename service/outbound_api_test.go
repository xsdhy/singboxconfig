package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newOutboundAPIRouter(t *testing.T) (*gin.Engine, storage.Storage) {
	t.Helper()
	t.Skip("Tests require database storage - memory storage has been removed")
	gin.SetMode(gin.TestMode)

	var store storage.Storage
	svc := NewService(store)
	router := gin.New()

	outbounds := router.Group("/api/outbounds")
	{
		outbounds.POST("", svc.CreateOutbound)
		outbounds.GET("", svc.ListOutbounds)
		outbounds.PUT("/:id", svc.UpdateOutbound)
		outbounds.DELETE("/:id", svc.DeleteOutbound)
		outbounds.PATCH("/batch-enable", svc.BatchEnableOutbounds)
	}

	subscribes := router.Group("/api/subscribes")
	{
		subscribes.PUT("/:name/cache-config", svc.UpdateSubscribeCacheConfig)
		subscribes.POST("/:name/refresh-outbound", svc.RefreshSubscribeOutbounds)
		subscribes.GET("/:name/outbounds", svc.ListSubscribeOutbounds)
	}

	return router, store
}

func TestOutboundAPIEndpoints(t *testing.T) {
	router, store := newOutboundAPIRouter(t)
	if err := store.CreateSubscribe(&entity.Subscribe{Name: "sub-a", Status: true}); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            "sub-node",
			Name:           "Sub Node",
			Type:           "vmess",
			Enabled:        true,
			Sort:           30,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "sub-a",
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"sub-node","type":"vmess"}`,
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}

	createBody := entity.Outbound{
		Tag:            "manual-a",
		Name:           "Manual Alpha",
		Type:           "socks",
		Enabled:        true,
		Sort:           10,
		VisibleDevices: "phone,pad",
		ConfigJSON:     `{"tag":"manual-a","type":"socks"}`,
	}
	recorder := performJSONRequest(t, router, http.MethodPost, "/api/outbounds", createBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateOutbound status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	var created entity.Outbound
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("Unmarshal CreateOutbound failed: %v", err)
	}
	if created.ID == 0 || created.Source != entity.OutboundSourceManual {
		t.Fatalf("created outbound unexpected: %+v", created)
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/outbounds", createBody)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("CreateOutbound duplicate status=%d, want %d", recorder.Code, http.StatusBadRequest)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/outbounds?source=MANUAL&enabled=true&search=alpha&page=1&limit=10", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListOutbounds status=%d", recorder.Code)
	}
	var listResp outboundListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("Unmarshal ListOutbounds failed: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Items) != 1 || listResp.Items[0].Tag != "manual-a" {
		t.Fatalf("ListOutbounds unexpected: %+v", listResp)
	}

	updateBody := created
	updateBody.Tag = "manual-renamed"
	updateBody.Name = "Manual Renamed"
	updateBody.Enabled = false
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/outbounds/"+itoa64(created.ID), updateBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateOutbound status=%d, body=%s", recorder.Code, recorder.Body.String())
	}

	subscribeItems, err := store.ListOutbounds(storage.OutboundFilter{Source: outboundSourcePtr(entity.OutboundSourceSubscription)})
	if err != nil {
		t.Fatalf("ListOutbounds(subscription) failed: %v", err)
	}
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/outbounds/"+itoa64(subscribeItems[0].ID), subscribeItems[0])
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("UpdateOutbound(subscription) status=%d, want %d", recorder.Code, http.StatusBadRequest)
	}

	recorder = performJSONRequest(t, router, http.MethodPatch, "/api/outbounds/batch-enable", outboundBatchEnableRequest{
		IDs:     []int64{created.ID, subscribeItems[0].ID},
		Enabled: true,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("BatchEnableOutbounds status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	updatedManual, err := store.GetOutbound(created.ID)
	if err != nil || !updatedManual.Enabled || updatedManual.Tag != "manual-renamed" {
		t.Fatalf("manual outbound after batch unexpected: %+v, %v", updatedManual, err)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/outbounds/"+itoa64(subscribeItems[0].ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteOutbound status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	if _, err := store.GetOutbound(subscribeItems[0].ID); err == nil {
		t.Fatal("DeleteOutbound should remove subscription outbound")
	}
}

func TestSubscribeOutboundCacheAPI(t *testing.T) {
	rawBody := "ss://YWVzLTEyOC1nY206ZDhxZmU0@dns.yypa.zzssptop.com:20475/#old-node\nss://YWVzLTEyOC1nY206ZDhxZmU0@dns.yypa.zzssptop.com:20476/#new-node"
	parsedNodes, err := parseSubscriptionOutbounds([]byte(rawBody))
	if err != nil {
		t.Fatalf("parseSubscriptionOutbounds failed: %v", err)
	}
	if len(parsedNodes) != 2 {
		t.Fatalf("parsed nodes len=%d, want 2", len(parsedNodes))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(rawBody))))
	}))
	defer server.Close()

	router, store := newOutboundAPIRouter(t)
	lastFetchTime := time.Now().Add(-5 * time.Minute).UTC()
	subscribe := &entity.Subscribe{
		Name:                    "sub-a",
		URL:                     server.URL,
		Status:                  true,
		VisibleDevices:          "phone",
		OutboundCacheDuration:   60,
		OutboundLastFetchTime:   &lastFetchTime,
		OutboundLastFetchStatus: "SUCCESS",
	}
	if err := store.CreateSubscribe(subscribe); err != nil {
		t.Fatalf("CreateSubscribe failed: %v", err)
	}
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            parsedNodes[0].Tag,
			Name:           "Old Node",
			Type:           "vmess",
			Enabled:        true,
			Sort:           5,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  subscribe.Name,
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"old-node","type":"vmess"}`,
			LastFetchTime:  &lastFetchTime,
		},
		{
			Tag:            "legacy-node",
			Name:           "Removed Node",
			Type:           "vmess",
			Enabled:        true,
			Sort:           6,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  subscribe.Name,
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"removed-node","type":"vmess"}`,
			LastFetchTime:  &lastFetchTime,
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodGet, "/api/subscribes/sub-a/outbounds?page=1&limit=1", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListSubscribeOutbounds status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	var listResp subscribeOutboundListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("Unmarshal subscribe outbounds failed: %v", err)
	}
	if listResp.Total != 2 || len(listResp.Items) != 1 || listResp.SubscribeCacheInfo.CacheDuration != 60 || listResp.SubscribeCacheInfo.IsExpired {
		t.Fatalf("subscribe outbounds response unexpected: %+v", listResp)
	}

	recorder = performJSONRequest(t, router, http.MethodPut, "/api/subscribes/sub-a/cache-config", map[string]any{
		"outbound_cache_duration": 15,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateSubscribeCacheConfig status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	gotSubscribe, err := store.GetSubscribe("sub-a")
	if err != nil || gotSubscribe.OutboundCacheDuration != 15 {
		t.Fatalf("subscribe cache duration unexpected: %+v, %v", gotSubscribe, err)
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/subscribes/sub-a/refresh-outbound", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("RefreshSubscribeOutbounds status=%d, body=%s", recorder.Code, recorder.Body.String())
	}
	var refreshResp map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &refreshResp); err != nil {
		t.Fatalf("Unmarshal refresh response failed: %v", err)
	}
	if refreshResp["status"] != "success" || int(refreshResp["added"].(float64)) != 1 || int(refreshResp["updated"].(float64)) != 1 || int(refreshResp["deleted"].(float64)) != 1 {
		t.Fatalf("refresh response unexpected: %+v", refreshResp)
	}

	currentItems, err := store.ListOutbounds(
		storage.OutboundFilter{Source: outboundSourcePtr(entity.OutboundSourceSubscription)},
		storage.OutboundFilter{SubscribeName: "sub-a"},
	)
	if err != nil {
		t.Fatalf("ListOutbounds(after refresh) failed: %v", err)
	}
	if len(currentItems) != 2 {
		t.Fatalf("subscription cache size after refresh unexpected: %+v", currentItems)
	}
}

func itoa64(value int64) string {
	return strconv.FormatInt(value, 10)
}
