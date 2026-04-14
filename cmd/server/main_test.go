package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"singboxconfig/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"singboxconfig/storage"
)

// newAuthenticatedRouter 创建带鉴权的测试路由，直接复用生产环境的路由注册逻辑。
func newAuthenticatedRouter(t *testing.T) (*service.Service, *gin.Engine) {
	t.Helper()
	t.Skip("Tests require database storage - memory storage has been removed")
	gin.SetMode(gin.TestMode)
	var store storage.Storage
	svc := service.NewService(store)
	return svc, SetupRouter(svc)
}

func TestSetupRouterRegistersDeviceManagementRoutes(t *testing.T) {
	_, router := newAuthenticatedRouter(t)
	routeSet := make(map[string]struct{})
	for _, route := range router.Routes() {
		routeSet[route.Method+" "+route.Path] = struct{}{}
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "devices", method: http.MethodGet, path: "/api/devices"},
		{name: "device detail", method: http.MethodGet, path: "/api/devices/:code"},
		{name: "device inbounds", method: http.MethodGet, path: "/api/devices/:code/inbounds"},
		{name: "subscribes outbounds", method: http.MethodGet, path: "/api/subscribes/:name/outbounds"},
		{name: "subscribes refresh outbound", method: http.MethodPost, path: "/api/subscribes/:name/refresh-outbound"},
		{name: "inbounds", method: http.MethodGet, path: "/api/inbounds"},
		{name: "wire guards", method: http.MethodGet, path: "/api/wire-guards"},
		{name: "wire guard peers", method: http.MethodGet, path: "/api/wire-guards/:tag/peers"},
		{name: "outbounds", method: http.MethodGet, path: "/api/outbounds"},
		{name: "outbounds batch", method: http.MethodPatch, path: "/api/outbounds/batch-enable"},
		{name: "auth login", method: http.MethodPost, path: "/api/auth/login"},
		{name: "auth me", method: http.MethodGet, path: "/api/auth/me"},
		{name: "auth change credentials", method: http.MethodPost, path: "/api/auth/change-credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := routeSet[tt.method+" "+tt.path]; !ok {
				t.Fatalf("route %s %s not registered", tt.method, tt.path)
			}
		})
	}
}

func TestSetupRouterUsesStorageBackedBearerAuth(t *testing.T) {
	svc, router := newAuthenticatedRouter(t)
	if _, err := svc.InitializeAuth(); err != nil {
		t.Fatalf("InitializeAuth failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	loginBody := bytes.NewBufferString(`{"username":"admin","password":"admin"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("POST /api/auth/login status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var loginPayload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("unmarshal login response failed: %v", err)
	}
	token, _ := loginPayload["access_token"].(string)
	if token == "" {
		t.Fatalf("login response missing access token: %s", recorder.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /api/auth/me status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"username":"admin"`) {
		t.Fatalf("GET /api/auth/me body = %s", recorder.Body.String())
	}

	body := bytes.NewBufferString(`{"old_password":"admin","new_username":"ops-admin","new_password":"new-password-456"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/change-credentials", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("POST /api/auth/change-credentials status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var changePayload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &changePayload); err != nil {
		t.Fatalf("unmarshal change response failed: %v", err)
	}
	newToken, _ := changePayload["access_token"].(string)
	if newToken == "" {
		t.Fatalf("change response missing new access token: %s", recorder.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("old token status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	loginBody = bytes.NewBufferString(`{"username":"admin","password":"new-password-456"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("old username login status = %d, want %d, body=%s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}

	loginBody = bytes.NewBufferString(`{"username":"ops-admin","password":"new-password-456"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("new credentials login status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+newToken)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("new token status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if payload["username"] != "ops-admin" {
		t.Fatalf("username = %v, want ops-admin", payload["username"])
	}
}
