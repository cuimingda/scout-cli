package cmd

import "testing"

func Test_loadScoutConfig(t *testing.T) {
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
