package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_resolveScoutConfigPath_prefersEnvOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("SCOUT_CONFIG", configPath)

	originalGOOS := currentGOOS
	originalUserConfigDir := userConfigDir
	t.Cleanup(func() {
		currentGOOS = originalGOOS
		userConfigDir = originalUserConfigDir
	})

	currentGOOS = "darwin"
	userConfigDir = func() (string, error) {
		return filepath.Join(t.TempDir(), "Library", "Application Support"), nil
	}

	if got := resolveScoutConfigPath(); got != configPath {
		t.Fatalf("resolveScoutConfigPath() = %q, want %q", got, configPath)
	}
}

func Test_resolveScoutConfigPath_onDarwinUsesDeveloperScopedPath(t *testing.T) {
	t.Setenv("SCOUT_CONFIG", "")

	originalGOOS := currentGOOS
	originalUserConfigDir := userConfigDir
	t.Cleanup(func() {
		currentGOOS = originalGOOS
		userConfigDir = originalUserConfigDir
	})

	configRoot := filepath.Join(t.TempDir(), "Library", "Application Support")
	currentGOOS = "darwin"
	userConfigDir = func() (string, error) {
		return configRoot, nil
	}

	want := filepath.Join(configRoot, "mingda.dev", "scout", "config.yaml")
	if got := resolveScoutConfigPath(); got != want {
		t.Fatalf("resolveScoutConfigPath() = %q, want %q", got, want)
	}
}

func Test_loadScoutConfig_withoutConfigReturnsDefault(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("SCOUT_CONFIG", configPath)

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
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("SCOUT_CONFIG", configPath)

	if err := os.WriteFile(configPath, []byte("dns:\n  - 1.1.1.1\n  - 9.9.9.9\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
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
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("SCOUT_CONFIG", configPath)

	if err := os.WriteFile(configPath, []byte("foo: bar\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := loadScoutConfig()
	if err != nil {
		t.Fatalf("loadScoutConfig() error = %v", err)
	}
	if len(cfg.DNS) != 2 || cfg.DNS[0] != "223.5.5.5" || cfg.DNS[1] != "8.8.8.8" {
		t.Fatalf("cfg.DNS = %#v, want default [223.5.5.5 8.8.8.8]", cfg.DNS)
	}
}
