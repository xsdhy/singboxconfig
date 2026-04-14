package service

import (
	"encoding/json"
	"net/http"
	"singboxconfig/convert/singbox"
	"singboxconfig/entity"
	"singboxconfig/storage"
	"testing"

	"github.com/gin-gonic/gin"
)

func newGeneratedRouter(t *testing.T) (*gin.Engine, storage.Storage) {
	t.Helper()
	t.Skip("Tests require database storage - memory storage has been removed")
	gin.SetMode(gin.TestMode)

	var store storage.Storage
	svc := NewService(store)

	router := gin.New()
	router.GET("/open/generate/:device", svc.Generated)
	return router, store
}

func TestGeneratedFallsBackToLegacyDefaults(t *testing.T) {
	router, _ := newGeneratedRouter(t)

	recorder := performJSONRequest(t, router, http.MethodGet, "/open/generate/phone?token=996007", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Generated(phone) status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var cfg entity.SingBoxConfig
	if err := json.Unmarshal(recorder.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("Unmarshal Generated(phone) failed: %v", err)
	}

	if cfg.Experimental != nil {
		t.Fatalf("phone experimental = %+v, want nil", cfg.Experimental)
	}
	if len(cfg.Inbounds) != 1 || cfg.Inbounds[0].Type != "tun" {
		t.Fatalf("phone inbounds unexpected: %+v", cfg.Inbounds)
	}
	if len(cfg.Endpoints) != 1 || cfg.Endpoints[0].Tag != "wg-ep" || cfg.Endpoints[0].Address[0] != "10.8.0.4/32" {
		t.Fatalf("phone endpoints unexpected: %+v", cfg.Endpoints)
	}
	if cfg.DNS.Final != singbox.GetDefaultDNS().Final {
		t.Fatalf("phone dns final = %s, want %s", cfg.DNS.Final, singbox.GetDefaultDNS().Final)
	}
	hasJGVPN := false
	hasSelfSIP := false
	for _, outbound := range cfg.Outbounds {
		if outbound.Tag == "jg-vpn" {
			hasJGVPN = true
		}
		if outbound.Tag == "selfsip" {
			hasSelfSIP = true
		}
	}
	if !hasJGVPN || !hasSelfSIP {
		t.Fatalf("phone default extra outbounds missing: %+v", cfg.Outbounds)
	}
}

func TestGeneratedUsesStoredDeviceTokenAndFiltersExtraOutbounds(t *testing.T) {
	router, store := newGeneratedRouter(t)

	if err := store.CreateDevice(&entity.Device{Code: "pad", Name: "Pad", Token: "token-pad", Enabled: true}); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	if err := store.CreateInbound(&entity.Inbound{Tag: "tun-custom", Enabled: true, Sort: 0, ConfigJSON: `{"type":"tun","stack":"system"}`}); err != nil {
		t.Fatalf("CreateInbound failed: %v", err)
	}
	if err := store.SetDeviceInbounds("pad", []*entity.DeviceInbound{{DeviceCode: "pad", InboundTag: "tun-custom", Sort: 0}}); err != nil {
		t.Fatalf("SetDeviceInbounds failed: %v", err)
	}
	if err := store.SetGlobalSetting("dns_config", `{"servers":[{"tag":"custom","address":"8.8.8.8"}],"rules":[],"final":"custom"}`); err != nil {
		t.Fatalf("SetGlobalSetting failed: %v", err)
	}
	if err := store.CreateExtraOutbound(&entity.Outbound{
		Tag:            "visible",
		Enabled:        true,
		Sort:           0,
		VisibleDevices: "pad,phone",
		ConfigJSON:     `{"tag":"visible","type":"socks","server":"127.0.0.1","server_port":1080}`,
	}); err != nil {
		t.Fatalf("CreateExtraOutbound(visible) failed: %v", err)
	}
	if err := store.CreateExtraOutbound(&entity.Outbound{
		Tag:            "hidden",
		Enabled:        true,
		Sort:           1,
		VisibleDevices: "tv",
		ConfigJSON:     `{"tag":"hidden","type":"socks","server":"127.0.0.1","server_port":2080}`,
	}); err != nil {
		t.Fatalf("CreateExtraOutbound(hidden) failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodGet, "/open/generate/pad?token=wrong", nil)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("Generated(pad wrong token) status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	recorder = performJSONRequest(t, router, http.MethodGet, "/open/generate/pad?token=token-pad", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Generated(pad) status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var cfg entity.SingBoxConfig
	if err := json.Unmarshal(recorder.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("Unmarshal Generated(pad) failed: %v", err)
	}

	if cfg.DNS.Final != "custom" {
		t.Fatalf("pad dns final = %s, want custom", cfg.DNS.Final)
	}
	if len(cfg.Inbounds) != 1 || cfg.Inbounds[0].Stack != "system" {
		t.Fatalf("pad inbounds unexpected: %+v", cfg.Inbounds)
	}

	hasVisible := false
	hasHidden := false
	for _, outbound := range cfg.Outbounds {
		if outbound.Tag == "visible" {
			hasVisible = true
		}
		if outbound.Tag == "hidden" {
			hasHidden = true
		}
	}
	if !hasVisible || hasHidden {
		t.Fatalf("pad outbounds visibility unexpected: %+v", cfg.Outbounds)
	}
}

func TestGeneratedDoesNotFallbackWhenStorageAlreadyHasDevices(t *testing.T) {
	router, store := newGeneratedRouter(t)

	if err := store.CreateDevice(&entity.Device{Code: "pad", Name: "Pad", Token: "token-pad", Enabled: true}); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	recorder := performJSONRequest(t, router, http.MethodGet, "/open/generate/phone?token=996007", nil)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("Generated(phone after custom devices) status = %d, want %d, body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}
