package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newDeviceManagementRouter 创建只包含本次新增 API 的测试路由，便于聚焦验证阶段二行为。
func newDeviceManagementRouter(t *testing.T) (*gin.Engine, storage.Storage) {
	t.Helper()
	t.Skip("Tests require database storage - memory storage has been removed")
	gin.SetMode(gin.TestMode)

	var store storage.Storage
	svc := NewService(store)

	router := gin.New()

	devices := router.Group("/api/devices")
	{
		devices.POST("", svc.CreateDevice)
		devices.GET("", svc.ListDevices)
		devices.GET("/:code", svc.GetDevice)
		devices.PUT("/:code", svc.UpdateDevice)
		devices.DELETE("/:code", svc.DeleteDevice)
		devices.PUT("/:code/inbounds", svc.SetDeviceInbounds)
		devices.GET("/:code/inbounds", svc.ListDeviceInbounds)
	}

	inbounds := router.Group("/api/inbounds")
	{
		inbounds.POST("", svc.CreateInbound)
		inbounds.GET("", svc.ListInbounds)
		inbounds.GET("/:tag", svc.GetInbound)
		inbounds.PUT("/:tag", svc.UpdateInbound)
		inbounds.DELETE("/:tag", svc.DeleteInbound)
	}

	wireGuards := router.Group("/api/wire-guards")
	{
		wireGuards.POST("", svc.CreateWireGuard)
		wireGuards.GET("", svc.ListWireGuards)
		wireGuards.GET("/:tag", svc.GetWireGuard)
		wireGuards.PUT("/:tag", svc.UpdateWireGuard)
		wireGuards.DELETE("/:tag", svc.DeleteWireGuard)
		wireGuards.POST("/:tag/peers", svc.CreateWireGuardPeer)
		wireGuards.GET("/:tag/peers", svc.ListWireGuardPeers)
		wireGuards.PUT("/:tag/peers/:id", svc.UpdateWireGuardPeer)
		wireGuards.DELETE("/:tag/peers/:id", svc.DeleteWireGuardPeer)
	}

	extraOutbounds := router.Group("/api/extra-outbounds")
	{
		extraOutbounds.POST("", svc.CreateExtraOutbound)
		extraOutbounds.GET("", svc.ListExtraOutbounds)
		extraOutbounds.GET("/:tag", svc.GetExtraOutbound)
		extraOutbounds.PUT("/:tag", svc.UpdateExtraOutbound)
		extraOutbounds.DELETE("/:tag", svc.DeleteExtraOutbound)
	}

	return router, store
}

// performJSONRequest 统一构造 JSON 请求，减少测试样板代码。
func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal body failed: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestDeviceCRUDAPI(t *testing.T) {
	router, _ := newDeviceManagementRouter(t)

	createBody := entity.Device{
		Code:                "phone",
		Name:                "Phone",
		Description:         "移动端设备",
		Token:               "token-phone",
		Enabled:             true,
		Sort:                20,
		WireGuardTag:        "wg-b",
		WireGuardClientAddr: "10.0.0.2",
		WireGuardClientKey:  "private-key",
	}
	recorder := performJSONRequest(t, router, http.MethodPost, "/api/devices", createBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateDevice status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/devices", entity.Device{Code: "pad"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateDevice(second) status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/devices", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListDevices status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var devices []entity.Device
	if err := json.Unmarshal(recorder.Body.Bytes(), &devices); err != nil {
		t.Fatalf("Unmarshal ListDevices failed: %v", err)
	}
	if len(devices) != 2 || devices[0].Code != "pad" || devices[1].Code != "phone" {
		t.Fatalf("unexpected devices order: %+v", devices)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/devices/phone", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GetDevice status = %d, want %d", recorder.Code, http.StatusOK)
	}

	updateBody := createBody
	updateBody.Name = "Phone Updated"
	updateBody.Sort = 1
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/devices/phone", updateBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateDevice status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPut, "/api/devices/phone", entity.Device{Code: "other"})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("UpdateDevice mismatch status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/devices/phone", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteDevice status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/devices/phone", nil)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GetDevice after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestInboundCRUDAPI(t *testing.T) {
	router, _ := newDeviceManagementRouter(t)

	createBody := entity.Inbound{
		Tag:         "mixed-in",
		Name:        "混合入口",
		Description: "默认入口",
		Type:        "mixed",
		Enabled:     true,
		Sort:        10,
		ConfigJSON:  `{"type":"mixed","listen":"::"}`,
	}
	recorder := performJSONRequest(t, router, http.MethodPost, "/api/inbounds", createBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateInbound status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/inbounds", entity.Inbound{Tag: "tun-in", Sort: 1})
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateInbound(second) status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/inbounds", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListInbounds status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var inbounds []entity.Inbound
	if err := json.Unmarshal(recorder.Body.Bytes(), &inbounds); err != nil {
		t.Fatalf("Unmarshal ListInbounds failed: %v", err)
	}
	if len(inbounds) != 2 || inbounds[0].Tag != "tun-in" || inbounds[1].Tag != "mixed-in" {
		t.Fatalf("unexpected inbounds order: %+v", inbounds)
	}

	updateBody := createBody
	updateBody.Name = "混合入口-已更新"
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/inbounds/mixed-in", updateBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateInbound status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/inbounds/mixed-in", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteInbound status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/inbounds/mixed-in", nil)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GetInbound after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeviceInboundAPI(t *testing.T) {
	router, store := newDeviceManagementRouter(t)

	if err := store.CreateDevice(&entity.Device{Code: "tv", Name: "TV"}); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	if err := store.CreateInbound(&entity.Inbound{Tag: "tun", Sort: 20}); err != nil {
		t.Fatalf("CreateInbound(tun) failed: %v", err)
	}
	if err := store.CreateInbound(&entity.Inbound{Tag: "mixed", Sort: 10}); err != nil {
		t.Fatalf("CreateInbound(mixed) failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodPut, "/api/devices/tv/inbounds", []entity.DeviceInbound{
		{InboundTag: "tun", Sort: 20},
		{DeviceCode: "tv", InboundTag: "mixed", Sort: 10},
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("SetDeviceInbounds status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/devices/tv/inbounds", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListDeviceInbounds status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var bindings []entity.DeviceInbound
	if err := json.Unmarshal(recorder.Body.Bytes(), &bindings); err != nil {
		t.Fatalf("Unmarshal ListDeviceInbounds failed: %v", err)
	}
	if len(bindings) != 2 || bindings[0].InboundTag != "mixed" || bindings[1].InboundTag != "tun" {
		t.Fatalf("unexpected bindings order: %+v", bindings)
	}
	if bindings[0].DeviceCode != "tv" || bindings[1].DeviceCode != "tv" {
		t.Fatalf("device code not injected correctly: %+v", bindings)
	}

	recorder = performJSONRequest(t, router, http.MethodPut, "/api/devices/tv/inbounds", []entity.DeviceInbound{
		{DeviceCode: "phone", InboundTag: "mixed", Sort: 1},
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("SetDeviceInbounds mismatch status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	recorder = performJSONRequest(t, router, http.MethodPut, "/api/devices/missing/inbounds", []entity.DeviceInbound{})
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("SetDeviceInbounds missing device status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestWireGuardAndPeerAPI(t *testing.T) {
	router, store := newDeviceManagementRouter(t)

	createWG := entity.WireGuard{
		Tag:         "wg-main",
		Name:        "主 WireGuard",
		Description: "主隧道",
		Enabled:     true,
		Sort:        20,
		EndpointTag: "wireguard-out",
		MTU:         1420,
	}
	recorder := performJSONRequest(t, router, http.MethodPost, "/api/wire-guards", createWG)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateWireGuard status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/wire-guards", entity.WireGuard{Tag: "wg-backup", Sort: 1})
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateWireGuard(second) status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/wire-guards", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListWireGuards status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var wireGuards []entity.WireGuard
	if err := json.Unmarshal(recorder.Body.Bytes(), &wireGuards); err != nil {
		t.Fatalf("Unmarshal ListWireGuards failed: %v", err)
	}
	if len(wireGuards) != 2 || wireGuards[0].Tag != "wg-backup" || wireGuards[1].Tag != "wg-main" {
		t.Fatalf("unexpected wire guards order: %+v", wireGuards)
	}

	createPeer := entity.WireGuardPeer{
		Sort:                        10,
		Address:                     "1.2.3.4",
		Port:                        51820,
		PublicKey:                   "public-key",
		PreSharedKey:                "pre-shared-key",
		AllowedIPs:                  "0.0.0.0/0",
		PersistentKeepaliveInterval: 25,
		Enabled:                     true,
	}
	recorder = performJSONRequest(t, router, http.MethodPost, "/api/wire-guards/wg-main/peers", createPeer)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateWireGuardPeer status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var createdPeer entity.WireGuardPeer
	if err := json.Unmarshal(recorder.Body.Bytes(), &createdPeer); err != nil {
		t.Fatalf("Unmarshal CreateWireGuardPeer failed: %v", err)
	}
	if createdPeer.ID == 0 || createdPeer.WireGuardTag != "wg-main" {
		t.Fatalf("unexpected created peer: %+v", createdPeer)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/wire-guards/wg-main/peers", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListWireGuardPeers status = %d, want %d", recorder.Code, http.StatusOK)
	}

	createdPeer.Address = "5.6.7.8"
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/wire-guards/wg-main/peers/"+strconv.FormatInt(createdPeer.ID, 10), createdPeer)
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateWireGuardPeer status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/wire-guards/wg-main/peers/"+strconv.FormatInt(createdPeer.ID, 10), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteWireGuardPeer status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/wire-guards/wg-main", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteWireGuard status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if err := store.CreateDevice(&entity.Device{
		Code:                "phone",
		Name:                "Phone",
		Token:               "token",
		Enabled:             true,
		WireGuardTag:        "wg-backup",
		WireGuardClientAddr: "10.0.0.2",
		WireGuardClientKey:  "private-key",
	}); err != nil {
		t.Fatalf("CreateDevice for delete wireguard verification failed: %v", err)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/wire-guards/wg-backup", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteWireGuard(second) status = %d, want %d", recorder.Code, http.StatusOK)
	}

	device, err := store.GetDevice("phone")
	if err != nil {
		t.Fatalf("GetDevice after DeleteWireGuard failed: %v", err)
	}
	if device.WireGuardTag != "" || device.WireGuardClientAddr != "" || device.WireGuardClientKey != "" {
		t.Fatalf("DeleteWireGuard should clear device binding fields: %+v", device)
	}
}

func TestExtraOutboundCRUDAPI(t *testing.T) {
	router, _ := newDeviceManagementRouter(t)

	createBody := entity.Outbound{
		Tag:            "selfsip",
		Name:           "Self SIP",
		Description:    "额外出站",
		Type:           "vmess",
		Enabled:        true,
		Sort:           20,
		VisibleDevices: "phone,tv",
		ConfigJSON:     `{"tag":"selfsip","type":"vmess"}`,
	}
	recorder := performJSONRequest(t, router, http.MethodPost, "/api/extra-outbounds", createBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateExtraOutbound status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/extra-outbounds", entity.Outbound{Tag: "backup", Sort: 1})
	if recorder.Code != http.StatusOK {
		t.Fatalf("CreateExtraOutbound(second) status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/extra-outbounds", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListExtraOutbounds status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var outbounds []entity.Outbound
	if err := json.Unmarshal(recorder.Body.Bytes(), &outbounds); err != nil {
		t.Fatalf("Unmarshal ListExtraOutbounds failed: %v", err)
	}
	if len(outbounds) != 2 || outbounds[0].Tag != "backup" || outbounds[1].Tag != "selfsip" {
		t.Fatalf("unexpected outbounds order: %+v", outbounds)
	}

	updateBody := createBody
	updateBody.Name = "Self SIP Updated"
	recorder = performJSONRequest(t, router, http.MethodPut, "/api/extra-outbounds/selfsip", updateBody)
	if recorder.Code != http.StatusOK {
		t.Fatalf("UpdateExtraOutbound status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/extra-outbounds/selfsip", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("DeleteExtraOutbound status = %d, want %d", recorder.Code, http.StatusOK)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/api/extra-outbounds/selfsip", nil)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GetExtraOutbound after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestExtraOutboundAPIOnlyHandlesManualRecords(t *testing.T) {
	router, store := newDeviceManagementRouter(t)
	if err := store.CreateOrUpdateOutbounds([]*entity.Outbound{
		{
			Tag:            "sub-node",
			Type:           "vmess",
			Enabled:        true,
			Source:         entity.OutboundSourceSubscription,
			SubscribeName:  "provider-a",
			VisibleDevices: "phone",
			ConfigJSON:     `{"tag":"sub-node","type":"vmess"}`,
		},
	}); err != nil {
		t.Fatalf("CreateOrUpdateOutbounds failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodGet, "/api/extra-outbounds", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListExtraOutbounds status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var outbounds []entity.Outbound
	if err := json.Unmarshal(recorder.Body.Bytes(), &outbounds); err != nil {
		t.Fatalf("Unmarshal ListExtraOutbounds failed: %v", err)
	}
	if len(outbounds) != 0 {
		t.Fatalf("subscription outbounds should be hidden from manual api: %+v", outbounds)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/extra-outbounds/sub-node", nil)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("DeleteExtraOutbound(subscription) status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeviceManagementAPIValidation(t *testing.T) {
	router, store := newDeviceManagementRouter(t)

	if err := store.CreateDevice(&entity.Device{Code: "phone"}); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	if err := store.CreateInbound(&entity.Inbound{Tag: "mixed"}); err != nil {
		t.Fatalf("CreateInbound failed: %v", err)
	}
	if err := store.CreateWireGuard(&entity.WireGuard{Tag: "wg-main"}); err != nil {
		t.Fatalf("CreateWireGuard failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodPut, "/api/devices/phone/inbounds", []entity.DeviceInbound{
		{InboundTag: "missing"},
	})
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "Inbound not found") {
		t.Fatalf("SetDeviceInbounds validation failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/wire-guards/wg-main/peers", entity.WireGuardPeer{
		WireGuardTag: "other",
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("CreateWireGuardPeer mismatch status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	recorder = performJSONRequest(t, router, http.MethodDelete, "/api/wire-guards/wg-main/peers/not-number", nil)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("DeleteWireGuardPeer invalid id status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
