package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"singboxconfig/storage"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestInitializeAuthCreatesCredentials(t *testing.T) {
	svc := newTestService(t)

	result, err := svc.InitializeAuth()
	if err != nil {
		t.Fatalf("InitializeAuth failed: %v", err)
	}
	if !result.Initialized {
		t.Fatalf("InitializeAuth should initialize on empty storage")
	}
	if result.Username != "admin" {
		t.Fatalf("username = %q, want %q", result.Username, "admin")
	}
	if result.Password != "admin" {
		t.Fatalf("password = %q, want %q", result.Password, "admin")
	}

	hash, err := svc.storage.GetGlobalSetting(authPasswordHashSettingKey)
	if err != nil {
		t.Fatalf("GetGlobalSetting(password hash) failed: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin")); err != nil {
		t.Fatalf("stored password hash does not match original password: %v", err)
	}

	tokenSecret, err := svc.storage.GetGlobalSetting(authTokenSecretSettingKey)
	if err != nil {
		t.Fatalf("GetGlobalSetting(token secret) failed: %v", err)
	}
	if strings.TrimSpace(tokenSecret) == "" {
		t.Fatalf("token secret should be generated")
	}
}

func TestInitializeAuthRejectsCorruptPartialConfig(t *testing.T) {
	svc := newTestService(t)
	if err := svc.storage.SetGlobalSetting(authUsernameSettingKey, "admin"); err != nil {
		t.Fatalf("SetGlobalSetting failed: %v", err)
	}

	_, err := svc.InitializeAuth()
	if !errors.Is(err, ErrAuthConfigCorrupt) {
		t.Fatalf("InitializeAuth error = %v, want %v", err, ErrAuthConfigCorrupt)
	}
}

func TestResetPasswordInitializesMissingAuth(t *testing.T) {
	svc := newTestService(t)

	result, err := svc.ResetPassword("", "strong-password-123")
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}
	if !result.Initialized {
		t.Fatalf("ResetPassword should initialize missing auth config")
	}

	username, err := svc.storage.GetGlobalSetting(authUsernameSettingKey)
	if err != nil {
		t.Fatalf("GetGlobalSetting(username) failed: %v", err)
	}
	if username != "admin" {
		t.Fatalf("username = %q, want %q", username, "admin")
	}
}

func TestReservedSettingsAreBlockedAndFiltered(t *testing.T) {
	svc := newTestService(t)

	createRecorder := httptest.NewRecorder()
	createCtx, _ := gin.CreateTestContext(createRecorder)
	createCtx.Request = httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(`{"key":"auth.username","value":"admin"}`))
	createCtx.Request.Header.Set("Content-Type", "application/json")
	svc.CreateSetting(createCtx)
	if createRecorder.Code != http.StatusForbidden {
		t.Fatalf("CreateSetting status = %d, want %d", createRecorder.Code, http.StatusForbidden)
	}

	if err := svc.storage.SetGlobalSetting(authUsernameSettingKey, "admin"); err != nil {
		t.Fatalf("SetGlobalSetting(auth.username) failed: %v", err)
	}
	if err := svc.storage.SetGlobalSetting("theme", "dark"); err != nil {
		t.Fatalf("SetGlobalSetting(theme) failed: %v", err)
	}

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	svc.ListSettings(listCtx)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("ListSettings status = %d, want %d", listRecorder.Code, http.StatusOK)
	}

	var payload map[string]string
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal list response failed: %v", err)
	}
	if _, exists := payload[authUsernameSettingKey]; exists {
		t.Fatalf("reserved auth setting should not be listed: %+v", payload)
	}
	if payload["theme"] != "dark" {
		t.Fatalf("theme = %q, want %q", payload["theme"], "dark")
	}
}

func TestGetAuthMeReturnsMetadata(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.InitializeAuth(); err != nil {
		t.Fatalf("InitializeAuth failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)

	svc.GetAuthMe(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GetAuthMe status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"username":"admin"`) {
		t.Fatalf("GetAuthMe body = %s", recorder.Body.String())
	}
}

func TestLoginWithPasswordIssuesBearerToken(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.InitializeAuth(); err != nil {
		t.Fatalf("InitializeAuth failed: %v", err)
	}

	result, err := svc.LoginWithPassword("admin", "admin")
	if err != nil {
		t.Fatalf("LoginWithPassword failed: %v", err)
	}
	if result.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if result.Username != "admin" {
		t.Fatalf("username = %q, want admin", result.Username)
	}
}

func TestChangeCredentialsRotatesSessionAndUpdatesUsername(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.InitializeAuth(); err != nil {
		t.Fatalf("InitializeAuth failed: %v", err)
	}

	firstLogin, err := svc.LoginWithPassword("admin", "admin")
	if err != nil {
		t.Fatalf("LoginWithPassword failed: %v", err)
	}

	updated, err := svc.ChangeCredentials("admin", "ops-admin", "new-password-456")
	if err != nil {
		t.Fatalf("ChangeCredentials failed: %v", err)
	}
	if updated.Username != "ops-admin" {
		t.Fatalf("updated username = %q, want ops-admin", updated.Username)
	}
	if updated.Token == "" {
		t.Fatalf("expected replacement token")
	}

	config, err := svc.getAuthConfig(true)
	if err != nil {
		t.Fatalf("getAuthConfig failed: %v", err)
	}
	if _, err := validateAuthToken(firstLogin.Token, config, time.Now().UTC()); !errors.Is(err, ErrInvalidAuthToken) {
		t.Fatalf("validateAuthToken(old token) error = %v, want %v", err, ErrInvalidAuthToken)
	}
	if _, err := validateAuthToken(updated.Token, config, time.Now().UTC()); err != nil {
		t.Fatalf("validateAuthToken(new token) failed: %v", err)
	}

	if _, err := svc.LoginWithPassword("admin", "new-password-456"); !errors.Is(err, ErrInvalidAuthToken) {
		t.Fatalf("old username should not log in, got %v", err)
	}
	if _, err := svc.LoginWithPassword("ops-admin", "new-password-456"); err != nil {
		t.Fatalf("new credentials login failed: %v", err)
	}
}

func TestReadAuthSettingDistinguishesNotFound(t *testing.T) {
	svc := newTestService(t)

	value, exists, err := svc.readAuthSetting(authUsernameSettingKey)
	if err != nil {
		t.Fatalf("readAuthSetting failed: %v", err)
	}
	if exists || value != "" {
		t.Fatalf("readAuthSetting = (%q, %v), want empty missing value", value, exists)
	}

	if err := svc.storage.SetGlobalSetting(authUsernameSettingKey, "admin"); err != nil {
		t.Fatalf("SetGlobalSetting failed: %v", err)
	}

	value, exists, err = svc.readAuthSetting(authUsernameSettingKey)
	if err != nil {
		t.Fatalf("readAuthSetting after set failed: %v", err)
	}
	if !exists || value != "admin" {
		t.Fatalf("readAuthSetting = (%q, %v), want (%q, true)", value, exists, "admin")
	}

	_, err = svc.storage.GetGlobalSetting("missing-key")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("GetGlobalSetting should return ErrNotFound, got %v", err)
	}
}
