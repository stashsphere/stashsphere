package testcommon

import (
	"testing"
	"time"

	"github.com/stashsphere/backend/config"
)

func BaseTestConfig(t *testing.T) config.StashSphereServeConfig {
	t.Helper()
	imageDir := t.TempDir()
	cacheDir := t.TempDir()
	tmpDir := t.TempDir()

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
		},
		FrontendUrl:  "http://localhost",
		InstanceName: "test",
		TmpPath:      tmpDir,
		Export: struct {
			StorePath         string        "koanf:\"storePath\""
			RetentionDuration time.Duration "koanf:\"retentionDuration\""
		}{
			StorePath:         tmpDir,
			RetentionDuration: time.Hour,
		},
		Import: struct {
			MaxUploadMB int64 "koanf:\"maxUploadMb\""
		}{
			MaxUploadMB: 1024,
		},
	}
}
