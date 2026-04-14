package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"singboxconfig/storage"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	authUsernameSettingKey          = "auth.username"
	authPasswordHashSettingKey      = "auth.password_hash"
	authInitializedAtSettingKey     = "auth.initialized_at"
	authPasswordChangedAtSettingKey = "auth.password_changed_at"
	authTokenSecretSettingKey       = "auth.token_secret"
	authSessionVersionSettingKey    = "auth.session_version"
	defaultAuthUsername             = "admin"
	minAdminPasswordLength          = 8
	authCacheTTL                    = 5 * time.Second
	authTokenTTL                    = 24 * time.Hour
	authPasswordCharset             = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%^&*()-_=+"
)

var (
	ErrAuthNotInitialized = errors.New("auth not initialized")
	ErrAuthConfigCorrupt  = errors.New("auth config is corrupt")
	ErrInvalidAuthToken   = errors.New("invalid auth token")
)

type AuthConfig struct {
	Username          string
	PasswordHash      string
	InitializedAt     time.Time
	PasswordChangedAt *time.Time
	TokenSecret       string
	SessionVersion    string
}

type AuthInitResult struct {
	Initialized       bool
	Username          string
	Password          string
	GeneratedPassword bool
	InitializedAt     time.Time
}

type AuthResetResult struct {
	Username          string
	Initialized       bool
	InitializedAt     time.Time
	PasswordChangedAt time.Time
}

type AuthLoginResult struct {
	Username          string
	Token             string
	ExpiresAt         time.Time
	IssuedAt          time.Time
	InitializedAt     time.Time
	PasswordChangedAt *time.Time
}

type authTokenClaims struct {
	Subject        string `json:"sub"`
	ExpiresAtUnix  int64  `json:"exp"`
	IssuedAtUnix   int64  `json:"iat"`
	SessionVersion string `json:"sv"`
	Type           string `json:"typ"`
}

type authCache struct {
	mu        sync.RWMutex
	ttl       time.Duration
	config    *AuthConfig
	expiresAt time.Time
}

func newAuthCache(ttl time.Duration) *authCache {
	return &authCache{ttl: ttl}
}

func (c *authCache) get(now time.Time) (*AuthConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.config == nil || now.After(c.expiresAt) {
		return nil, false
	}
	return cloneAuthConfig(c.config), true
}

func (c *authCache) set(config *AuthConfig, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cloneAuthConfig(config)
	c.expiresAt = now.Add(c.ttl)
}

func (c *authCache) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = nil
	c.expiresAt = time.Time{}
}

func cloneAuthConfig(config *AuthConfig) *AuthConfig {
	if config == nil {
		return nil
	}
	cloned := *config
	if config.PasswordChangedAt != nil {
		changedAt := *config.PasswordChangedAt
		cloned.PasswordChangedAt = &changedAt
	}
	return &cloned
}

func isReservedGlobalSettingKey(key string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(key)), "auth.")
}

func normalizeAuthUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return defaultAuthUsername
	}
	return username
}

func validateAdminPassword(password string) error {
	if len(password) < minAdminPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minAdminPasswordLength)
	}
	return nil
}

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid password length: %d", length)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	chars := []byte(authPasswordCharset)
	for i := range bytes {
		bytes[i] = chars[int(bytes[i])%len(chars)]
	}
	return string(bytes), nil
}

func generateRandomURLSafeToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid token length: %d", length)
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseRFC3339Setting(key, value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid %s: %v", ErrAuthConfigCorrupt, key, err)
	}
	return ts.UTC(), nil
}

func (s *Service) invalidateAuthCache() {
	if s.auth != nil {
		s.auth.invalidate()
	}
}

func (s *Service) getAuthConfig(forceRefresh bool) (*AuthConfig, error) {
	now := time.Now().UTC()
	if !forceRefresh && s.auth != nil {
		if cached, ok := s.auth.get(now); ok {
			return cached, nil
		}
	}

	username, usernameExists, err := s.readAuthSetting(authUsernameSettingKey)
	if err != nil {
		return nil, err
	}
	passwordHash, passwordExists, err := s.readAuthSetting(authPasswordHashSettingKey)
	if err != nil {
		return nil, err
	}

	switch {
	case !usernameExists && !passwordExists:
		return nil, ErrAuthNotInitialized
	case usernameExists != passwordExists:
		return nil, fmt.Errorf("%w: username/password hash mismatch", ErrAuthConfigCorrupt)
	}

	config := &AuthConfig{
		Username:     username,
		PasswordHash: passwordHash,
	}

	initializedAtValue, initializedAtExists, err := s.readAuthSetting(authInitializedAtSettingKey)
	if err != nil {
		return nil, err
	}
	if initializedAtExists {
		config.InitializedAt, err = parseRFC3339Setting(authInitializedAtSettingKey, initializedAtValue)
		if err != nil {
			return nil, err
		}
	}

	passwordChangedAtValue, passwordChangedAtExists, err := s.readAuthSetting(authPasswordChangedAtSettingKey)
	if err != nil {
		return nil, err
	}
	if passwordChangedAtExists {
		parsed, err := parseRFC3339Setting(authPasswordChangedAtSettingKey, passwordChangedAtValue)
		if err != nil {
			return nil, err
		}
		if !parsed.IsZero() {
			config.PasswordChangedAt = &parsed
		}
	}

	tokenSecret, tokenSecretExists, err := s.readAuthSetting(authTokenSecretSettingKey)
	if err != nil {
		return nil, err
	}
	sessionVersion, sessionVersionExists, err := s.readAuthSetting(authSessionVersionSettingKey)
	if err != nil {
		return nil, err
	}

	needsPersist := false
	if !tokenSecretExists || strings.TrimSpace(tokenSecret) == "" {
		tokenSecret, err = generateRandomURLSafeToken(32)
		if err != nil {
			return nil, err
		}
		needsPersist = true
	}
	if !sessionVersionExists || strings.TrimSpace(sessionVersion) == "" {
		sessionVersion, err = generateRandomURLSafeToken(18)
		if err != nil {
			return nil, err
		}
		needsPersist = true
	}

	config.TokenSecret = tokenSecret
	config.SessionVersion = sessionVersion

	if needsPersist {
		if err := s.saveAuthConfig(config); err != nil {
			return nil, err
		}
	}

	if s.auth != nil {
		s.auth.set(config, now)
	}
	return config, nil
}

func (s *Service) readAuthSetting(key string) (string, bool, error) {
	value, err := s.storage.GetGlobalSetting(key)
	if err == nil {
		return value, true, nil
	}
	if errors.Is(err, storage.ErrNotFound) {
		return "", false, nil
	}
	return "", false, err
}

func (s *Service) saveAuthConfig(config *AuthConfig) error {
	if strings.TrimSpace(config.Username) == "" {
		return fmt.Errorf("%w: missing username", ErrAuthConfigCorrupt)
	}
	if strings.TrimSpace(config.PasswordHash) == "" {
		return fmt.Errorf("%w: missing password hash", ErrAuthConfigCorrupt)
	}
	if strings.TrimSpace(config.TokenSecret) == "" {
		return fmt.Errorf("%w: missing token secret", ErrAuthConfigCorrupt)
	}
	if strings.TrimSpace(config.SessionVersion) == "" {
		return fmt.Errorf("%w: missing session version", ErrAuthConfigCorrupt)
	}

	if err := s.storage.SetGlobalSetting(authUsernameSettingKey, config.Username); err != nil {
		return err
	}
	if err := s.storage.SetGlobalSetting(authPasswordHashSettingKey, config.PasswordHash); err != nil {
		return err
	}
	if err := s.storage.SetGlobalSetting(authTokenSecretSettingKey, config.TokenSecret); err != nil {
		return err
	}
	if err := s.storage.SetGlobalSetting(authSessionVersionSettingKey, config.SessionVersion); err != nil {
		return err
	}

	if !config.InitializedAt.IsZero() {
		if err := s.storage.SetGlobalSetting(authInitializedAtSettingKey, formatRFC3339(config.InitializedAt)); err != nil {
			return err
		}
	}
	if config.PasswordChangedAt != nil {
		if err := s.storage.SetGlobalSetting(authPasswordChangedAtSettingKey, formatRFC3339(*config.PasswordChangedAt)); err != nil {
			return err
		}
	} else {
		if err := s.storage.DeleteGlobalSetting(authPasswordChangedAtSettingKey); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
	}

	s.invalidateAuthCache()
	return nil
}

func (s *Service) issueAuthToken(config *AuthConfig, now time.Time) (string, time.Time, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := now.Add(authTokenTTL).UTC()
	claimsJSON, err := json.Marshal(authTokenClaims{
		Subject:        config.Username,
		ExpiresAtUnix:  expiresAt.Unix(),
		IssuedAtUnix:   now.Unix(),
		SessionVersion: config.SessionVersion,
		Type:           "access",
	})
	if err != nil {
		return "", time.Time{}, err
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsPart := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerPart + "." + claimsPart

	mac := hmac.New(sha256.New, []byte(config.TokenSecret))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", time.Time{}, err
	}
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + signaturePart, expiresAt, nil
}

func parseBearerToken(headerValue string) (string, bool) {
	if strings.TrimSpace(headerValue) == "" {
		return "", false
	}
	parts := strings.SplitN(strings.TrimSpace(headerValue), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

func validateAuthToken(token string, config *AuthConfig, now time.Time) (*authTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidAuthToken
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(config.TokenSecret))
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return nil, err
	}
	expectedSignature := mac.Sum(nil)
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidAuthToken
	}
	if subtle.ConstantTimeCompare(actualSignature, expectedSignature) != 1 {
		return nil, ErrInvalidAuthToken
	}

	claimsPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidAuthToken
	}

	var claims authTokenClaims
	if err := json.Unmarshal(claimsPayload, &claims); err != nil {
		return nil, ErrInvalidAuthToken
	}
	if claims.Type != "access" {
		return nil, ErrInvalidAuthToken
	}
	if claims.ExpiresAtUnix <= now.Unix() {
		return nil, ErrInvalidAuthToken
	}
	if subtle.ConstantTimeCompare([]byte(claims.Subject), []byte(config.Username)) != 1 {
		return nil, ErrInvalidAuthToken
	}
	if subtle.ConstantTimeCompare([]byte(claims.SessionVersion), []byte(config.SessionVersion)) != 1 {
		return nil, ErrInvalidAuthToken
	}
	return &claims, nil
}

func (s *Service) InitializeAuth() (*AuthInitResult, error) {
	config, err := s.getAuthConfig(true)
	if err == nil {
		return &AuthInitResult{
			Initialized:   false,
			Username:      config.Username,
			InitializedAt: config.InitializedAt,
		}, nil
	}
	if !errors.Is(err, ErrAuthNotInitialized) {
		return nil, err
	}

	// 首次启动，使用默认的 admin/admin
	username := "admin"
	password := "admin"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	initializedAt := time.Now().UTC()
	sessionVersion, err := generateRandomURLSafeToken(18)
	if err != nil {
		return nil, err
	}
	tokenSecret, err := generateRandomURLSafeToken(32)
	if err != nil {
		return nil, err
	}

	if err := s.saveAuthConfig(&AuthConfig{
		Username:       username,
		PasswordHash:   string(passwordHash),
		InitializedAt:  initializedAt,
		TokenSecret:    tokenSecret,
		SessionVersion: sessionVersion,
	}); err != nil {
		return nil, err
	}

	return &AuthInitResult{
		Initialized:       true,
		Username:          username,
		Password:          password,
		GeneratedPassword: false,
		InitializedAt:     initializedAt,
	}, nil
}

func (s *Service) ResetPassword(_, newPassword string) (*AuthResetResult, error) {
	if err := validateAdminPassword(newPassword); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	config, err := s.getAuthConfig(true)
	initialized := false

	switch {
	case err == nil:
	case errors.Is(err, ErrAuthNotInitialized):
		initialized = true
		sessionVersion, versionErr := generateRandomURLSafeToken(18)
		if versionErr != nil {
			return nil, versionErr
		}
		tokenSecret, secretErr := generateRandomURLSafeToken(32)
		if secretErr != nil {
			return nil, secretErr
		}
		config = &AuthConfig{
			Username:       "admin",
			InitializedAt:  now,
			TokenSecret:    tokenSecret,
			SessionVersion: sessionVersion,
		}
	default:
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	newSessionVersion, err := generateRandomURLSafeToken(18)
	if err != nil {
		return nil, err
	}

	config.PasswordHash = string(passwordHash)
	config.SessionVersion = newSessionVersion
	config.PasswordChangedAt = &now
	if config.InitializedAt.IsZero() {
		config.InitializedAt = now
	}

	if err := s.saveAuthConfig(config); err != nil {
		return nil, err
	}

	return &AuthResetResult{
		Username:          config.Username,
		Initialized:       initialized,
		InitializedAt:     config.InitializedAt,
		PasswordChangedAt: now,
	}, nil
}

func (s *Service) LoginWithPassword(username, password string) (*AuthLoginResult, error) {
	config, err := s.getAuthConfig(false)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(username), []byte(config.Username)) != 1 {
		return nil, ErrInvalidAuthToken
	}
	if err := bcrypt.CompareHashAndPassword([]byte(config.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidAuthToken
	}

	now := time.Now().UTC()
	token, expiresAt, err := s.issueAuthToken(config, now)
	if err != nil {
		return nil, err
	}

	return &AuthLoginResult{
		Username:          config.Username,
		Token:             token,
		ExpiresAt:         expiresAt,
		IssuedAt:          now,
		InitializedAt:     config.InitializedAt,
		PasswordChangedAt: config.PasswordChangedAt,
	}, nil
}

func (s *Service) ChangeCredentials(oldPassword, newUsername, newPassword string) (*AuthLoginResult, error) {
	config, err := s.getAuthConfig(true)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(config.PasswordHash), []byte(oldPassword)); err != nil {
		return nil, fmt.Errorf("old password is incorrect")
	}

	updated := cloneAuthConfig(config)
	nextUsername := strings.TrimSpace(newUsername)
	usernameChanged := nextUsername != "" && nextUsername != config.Username
	passwordChanged := strings.TrimSpace(newPassword) != ""

	if !usernameChanged && !passwordChanged {
		return nil, fmt.Errorf("no credential changes provided")
	}

	if usernameChanged {
		updated.Username = nextUsername
	}
	if passwordChanged {
		if err := validateAdminPassword(newPassword); err != nil {
			return nil, err
		}
		if oldPassword == newPassword {
			return nil, fmt.Errorf("new password must be different")
		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updated.PasswordHash = string(passwordHash)
		now := time.Now().UTC()
		updated.PasswordChangedAt = &now
	}

	updated.SessionVersion, err = generateRandomURLSafeToken(18)
	if err != nil {
		return nil, err
	}

	if err := s.saveAuthConfig(updated); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	token, expiresAt, err := s.issueAuthToken(updated, now)
	if err != nil {
		return nil, err
	}

	return &AuthLoginResult{
		Username:          updated.Username,
		Token:             token,
		ExpiresAt:         expiresAt,
		IssuedAt:          now,
		InitializedAt:     updated.InitializedAt,
		PasswordChangedAt: updated.PasswordChangedAt,
	}, nil
}

func (s *Service) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := parseBearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortUnauthorized(c)
			return
		}

		config, err := s.getAuthConfig(false)
		if err != nil {
			log.Printf("auth config load failed: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load auth config"})
			return
		}

		claims, err := validateAuthToken(token, config, time.Now().UTC())
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(gin.AuthUserKey, claims.Subject)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", `Bearer realm="admin"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}

func (s *Service) Login(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.LoginWithPassword(strings.TrimSpace(request.Username), request.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidAuthToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "username or password is incorrect"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"access_token": result.Token,
		"token_type":   "Bearer",
		"expires_at":   formatRFC3339(result.ExpiresAt),
		"username":     result.Username,
	}
	if result.PasswordChangedAt != nil {
		response["password_changed_at"] = formatRFC3339(*result.PasswordChangedAt)
	}

	c.JSON(http.StatusOK, response)
}

func (s *Service) GetAuthMe(c *gin.Context) {
	config, err := s.getAuthConfig(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"username":       config.Username,
		"initialized_at": formatRFC3339(config.InitializedAt),
	}
	if config.PasswordChangedAt != nil {
		response["password_changed_at"] = formatRFC3339(*config.PasswordChangedAt)
	}

	c.JSON(http.StatusOK, response)
}

func (s *Service) ChangeCredentialsHandler(c *gin.Context) {
	var request struct {
		OldPassword string `json:"old_password"`
		NewUsername string `json:"new_username"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(request.OldPassword) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is required"})
		return
	}

	result, err := s.ChangeCredentials(request.OldPassword, request.NewUsername, request.NewPassword)
	if err != nil {
		switch err.Error() {
		case "old password is incorrect", "no credential changes provided", "new password must be different":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "password must be at least") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"message":      "credentials updated successfully",
		"username":     result.Username,
		"access_token": result.Token,
		"token_type":   "Bearer",
		"expires_at":   formatRFC3339(result.ExpiresAt),
	}
	if result.PasswordChangedAt != nil {
		response["password_changed_at"] = formatRFC3339(*result.PasswordChangedAt)
	}

	c.JSON(http.StatusOK, response)
}
