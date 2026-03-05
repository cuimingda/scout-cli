package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func Test_loadScoutConfig(t *testing.T) {
	origConfigHome := xdg.ConfigHome
	defer func() {
		xdg.ConfigHome = origConfigHome
	}()
	xdg.ConfigHome = filepath.Join(os.TempDir(), "scout-cli-test-config-home-default")

	cfg, err := loadScoutConfig()
	if err != nil {
		t.Fatalf("loadScoutConfig() error = %v", err)
	}
	if len(cfg.DNS) != 2 {
		t.Fatalf("len(cfg.DNS) = %d, want 2", len(cfg.DNS))
	}
	if cfg.DNS[0] != "223.5.5.5" || cfg.DNS[1] != "8.8.8.8" {
		t.Fatalf("cfg.DNS = %#v, want [223.5.5.5 8.8.8.8]", cfg.DNS)
	}
}

func Test_loadScoutConfig_withUserConfigOverridesDNS(t *testing.T) {
	origConfigHome := xdg.ConfigHome
	tmpDir := t.TempDir()
	xdg.ConfigHome = tmpDir
	defer func() {
		xdg.ConfigHome = origConfigHome
	}()

	userConfigPath := filepath.Join(xdg.ConfigHome, "scout", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("dns:\n  - 1.1.1.1\n  - 9.9.9.9\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg, err := loadScoutConfig()
	if err != nil {
		t.Fatalf("loadScoutConfig() error = %v", err)
	}
	if len(cfg.DNS) != 2 || cfg.DNS[0] != "1.1.1.1" || cfg.DNS[1] != "9.9.9.9" {
		t.Fatalf("cfg.DNS = %#v, want [1.1.1.1 9.9.9.9]", cfg.DNS)
	}
}

func Test_loadScoutConfig_withUserConfigWithoutDNSKeepsDefault(t *testing.T) {
	origConfigHome := xdg.ConfigHome
	tmpDir := t.TempDir()
	xdg.ConfigHome = tmpDir
	defer func() {
		xdg.ConfigHome = origConfigHome
	}()

	userConfigPath := filepath.Join(xdg.ConfigHome, "scout", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("foo: bar\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg, err := loadScoutConfig()
	if err != nil {
		t.Fatalf("loadScoutConfig() error = %v", err)
	}
	if len(cfg.DNS) != 2 || cfg.DNS[0] != "223.5.5.5" || cfg.DNS[1] != "8.8.8.8" {
		t.Fatalf("cfg.DNS = %#v, want default [223.5.5.5 8.8.8.8]", cfg.DNS)
	}
}
