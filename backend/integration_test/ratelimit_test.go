package integration_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stashsphere/backend/cmd"
	"github.com/stashsphere/backend/config"
	testcommon "github.com/stashsphere/backend/test_common"
	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T) config.StashSphereServeConfig {
	imageDir := t.TempDir()
	cacheDir := t.TempDir()

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
			ApiDomain      string   `koanf:"api"`
		}{
			AllowedDomains: []string{"http://localhost"},
			ApiDomain:      "",
		},
		FrontendUrl:  "http://localhost",
		InstanceName: "test",
	}
}

func TestLoginRateLimiting(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, testConfig(t), false, false, "")
	assert.NoError(t, err)

	// Make 6 requests - first 5 should succeed (or fail auth), 6th should be rate limited
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login",
			strings.NewReader(`{"email":"test@example.com","password":"password"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if i < 5 {
			assert.NotEqual(t, http.StatusTooManyRequests, rec.Code,
				"Request %d should not be rate limited", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rec.Code,
				"Request %d should be rate limited", i+1)
		}
	}
}

func TestRegisterRateLimiting(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, testConfig(t), false, false, "")
	assert.NoError(t, err)

	// Make 6 requests - first 5 should succeed (or fail validation), 6th should be rate limited
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/user/register",
			strings.NewReader(`{"email":"invalid","password":"x","name":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if i < 5 {
			assert.NotEqual(t, http.StatusTooManyRequests, rec.Code,
				"Request %d should not be rate limited", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, rec.Code,
				"Request %d should be rate limited", i+1)
		}
	}
}

func TestRateLimitingPerIP(t *testing.T) {
	db, tearDown, err := testcommon.CreateTestSchema()
	assert.NoError(t, err)
	t.Cleanup(tearDown)
	t.Cleanup(func() { db.Close() })

	e, _, err := cmd.SetupWithDB(db, testConfig(t), false, false, "")
	assert.NoError(t, err)

	// Exhaust rate limit for IP 1.1.1.1
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login",
			strings.NewReader(`{"email":"test@example.com","password":"password"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Real-IP", "1.1.1.1")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	// 6th request from same IP should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"email":"test@example.com","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "1.1.1.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code,
		"6th request from same IP should be rate limited")

	// Request from different IP should not be rate limited
	req = httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"email":"test@example.com","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "2.2.2.2")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.NotEqual(t, http.StatusTooManyRequests, rec.Code,
		"Request from different IP should not be rate limited")
}
