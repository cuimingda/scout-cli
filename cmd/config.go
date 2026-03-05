package cmd

import (
	"fmt"

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

	return cfg, nil
}
