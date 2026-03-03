package integration_test

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stashsphere/backend/cmd"
	"github.com/stashsphere/backend/config"
	"github.com/stashsphere/backend/crypto"
	"github.com/stashsphere/backend/factories"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/services"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

func oidcBaseConfig(t *testing.T) config.StashSphereServeConfig {
	imageDir := t.TempDir()
	cacheDir := t.TempDir()

	keyStr, err := crypto.GenerateEd25519StringKey()
	assert.NoError(t, err)

	return config.StashSphereServeConfig{
		ListenAddress: ":8081",
		Image: struct {
			Path      string `koanf:"path"`
			CachePath string `koanf:"cachePath"`
		}{
			Path:      imageDir,
			CachePath: cacheDir,
		},
		Domains: struct {
			AllowedDomains []string `koanf:"allowed"`
			CookieDomain   string   `koanf:"cookieDomain"`
			ApiDomain      string   `koanf:"api"`
		}{
			AllowedDomains: []string{"http://localhost"},
			CookieDomain:   "",
			ApiDomain:      "",
		},
		Auth: struct {
			PrivateKey           string            `koanf:"privateKey"`
			DisableSecureCookies bool              `koanf:"disableSecureCookies"`
			OIDC                 config.OIDCConfig `koanf:"oidc"`
		}{
			PrivateKey:           keyStr,
			DisableSecureCookies: true,
		},
		FrontendUrl:  "http://localhost",
		InstanceName: "test",
		BaseURL:      "http://localhost:8081",
	}
}

func oidcTestConfig(t *testing.T) config.StashSphereServeConfig {
	cfg := oidcBaseConfig(t)
	cfg.Auth.OIDC = config.OIDCConfig{
		Enabled: true,
		Providers: []config.OIDCProviderConfig{
			{
				Name:         "test",
				DisplayName:  "Test Provider",
				IssuerURL:    "https://issuer.example.com",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
		},
	}
	return cfg
}

func oidcServiceFromConfig(t *testing.T, db *sql.DB, cfg config.StashSphereServeConfig) *services.OIDCService {
	t.Helper()
	privateKey, err := crypto.LoadEd22519PrivateKeyFromString(cfg.Auth.PrivateKey)
	assert.NoError(t, err)
	return services.NewOIDCService(db, cfg.Auth.OIDC, cfg.BaseURL, privateKey, privateKey.Public().(ed25519.PublicKey))
}

func TestInfoReturnsOIDCProviders(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &body)
	assert.NoError(t, err)

	providers, ok := body["oidcProviders"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, providers, 1)

	p := providers[0].(map[string]interface{})
	assert.Equal(t, "test", p["name"])
	assert.Equal(t, "Test Provider", p["displayName"])
}

func TestInfoOIDCDisabledReturnsEmptyArray(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	cfg := oidcBaseConfig(t)
	cfg.Auth.OIDC = config.OIDCConfig{Enabled: false}
	e, _, err := cmd.SetupWithDB(db, cfg, false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &body)
	assert.NoError(t, err)

	providers, ok := body["oidcProviders"].([]interface{})
	assert.True(t, ok)
	assert.Empty(t, providers)
}

func TestAuthorizeUnknownProvider(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/nonexistent/authorize", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCallbackMissingStateCookie(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/test/callback?state=abc&code=xyz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCallbackStateMismatch(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/test/callback?state=bbb&code=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oidc-state", Value: "aaa"})
	req.AddCookie(&http.Cookie{Name: "oidc-nonce", Value: "nonce"})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCallbackProviderErrorRedirect(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet,
		"/api/auth/oidc/test/callback?state="+state+"&error=access_denied&error_description=user+denied", nil)
	req.AddCookie(&http.Cookie{Name: "oidc-state", Value: state})
	req.AddCookie(&http.Cookie{Name: "oidc-nonce", Value: "nonce"})
	req.AddCookie(&http.Cookie{Name: "oidc-redirect", Value: "http://localhost/auth/callback"})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "http://localhost/auth/callback")
	assert.Contains(t, location, "error=access_denied")
	assert.Contains(t, location, "error_description=user+denied")
}

func TestLinkMissingParams(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/test/link",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLinkInvalidChallengeToken(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/test/link",
		strings.NewReader(`{"password":"somepass","challengeToken":"garbage-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLinkWrongPassword(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	cfg := oidcTestConfig(t)
	e, _, err := cmd.SetupWithDB(db, cfg, false, false, "")
	assert.NoError(t, err)

	ctx := context.Background()
	userService := services.NewUserService(db, false, "", 60, nil)
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	_, err = userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)

	oidcService := oidcServiceFromConfig(t, db, cfg)
	result, err := oidcService.FindOrLinkUser(ctx, "test", &services.OIDCUserInfo{
		Subject: "sub-wrong-pw",
		Email:   testUserParams.Email,
		Name:    "Test",
	})
	assert.NoError(t, err)
	assert.Equal(t, "link_required", result.Action)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/test/link",
		strings.NewReader(`{"password":"wrong-password","challengeToken":"`+result.ChallengeToken+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLinkSuccess(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	cfg := oidcTestConfig(t)
	e, _, err := cmd.SetupWithDB(db, cfg, false, false, "")
	assert.NoError(t, err)

	ctx := context.Background()
	userService := services.NewUserService(db, false, "", 60, nil)
	testUserParams := factories.UserFactory.MustCreate().(*services.CreateUserParams)
	_, err = userService.CreateUser(ctx, *testUserParams)
	assert.NoError(t, err)

	oidcService := oidcServiceFromConfig(t, db, cfg)
	result, err := oidcService.FindOrLinkUser(ctx, "test", &services.OIDCUserInfo{
		Subject: "sub-link-success",
		Email:   testUserParams.Email,
		Name:    "Link Test",
	})
	assert.NoError(t, err)
	assert.Equal(t, "link_required", result.Action)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/oidc/test/link",
		strings.NewReader(`{"password":"`+testUserParams.Password+`","challengeToken":"`+result.ChallengeToken+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify auth cookies are set
	cookies := rec.Result().Cookies()
	cookieNames := make(map[string]bool)
	for _, c := range cookies {
		cookieNames[c.Name] = true
	}
	assert.True(t, cookieNames["stashsphere-access"], "expected stashsphere-access cookie")
	assert.True(t, cookieNames["stashsphere-info"], "expected stashsphere-info cookie")

	// Verify external_auth row exists
	count, err := models.ExternalAuths(
		models.ExternalAuthWhere.Provider.EQ("test"),
		models.ExternalAuthWhere.Subject.EQ("sub-link-success"),
	).Count(ctx, db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestCallbackRequiresRedirectCookie(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, oidcTestConfig(t), false, false, "")
	assert.NoError(t, err)

	state := "valid-state"
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/test/callback?state="+state+"&code=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oidc-state", Value: state})
	req.AddCookie(&http.Cookie{Name: "oidc-nonce", Value: "nonce"})
	// Missing oidc-redirect cookie
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
