package cmd

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

func loadScoutConfig() (scoutConfig, error) {
	cfg := defaultScoutConfig
	configPath := resolveScoutConfigPath()

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return scoutConfig{}, err
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return scoutConfig{}, err
	}

	var userCfg scoutConfig
	if err := v.Unmarshal(&userCfg); err != nil {
		return scoutConfig{}, err
	}
	if len(userCfg.DNS) > 0 {
		cfg.DNS = userCfg.DNS
	}
	return cfg, nil
}

func resolveScoutConfigPath() string {
	if envPath := os.Getenv("SCOUT_CONFIG"); envPath != "" {
		return envPath
	}
	return filepath.Join(xdg.ConfigHome, "scout", "config.yaml")
}
