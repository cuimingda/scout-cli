package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type scoutConfig struct {
	DNS []string `mapstructure:"dns"`
}

func loadScoutConfig() (scoutConfig, error) {
	v := viper.New()
	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath("../config")
	v.AddConfigPath(".")
	v.SetDefault("dns", []string{})

	if err := v.ReadInConfig(); err != nil {
		return scoutConfig{}, fmt.Errorf("failed to load config: %w", err)
	}

	var cfg scoutConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return scoutConfig{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	userCfgPath := filepath.Join(xdg.ConfigHome, "scout", "config.yaml")
	if _, err := os.Stat(userCfgPath); err == nil {
		userV := viper.New()
		userV.SetConfigFile(userCfgPath)
		userV.SetConfigType("yaml")
		if err := userV.ReadInConfig(); err != nil {
			return scoutConfig{}, fmt.Errorf("failed to load user config: %w", err)
		}

		userCfg := scoutConfig{}
		if err := userV.Unmarshal(&userCfg); err != nil {
			return scoutConfig{}, fmt.Errorf("failed to unmarshal user config: %w", err)
		}
		if userV.IsSet("dns") {
			cfg.DNS = userCfg.DNS
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return scoutConfig{}, fmt.Errorf("failed to read user config path %q: %w", userCfgPath, err)
	}

	return cfg, nil
}
