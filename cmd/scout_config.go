package cmd

type scoutConfig struct {
	DNS []string `mapstructure:"dns"`
}

var defaultScoutConfig = scoutConfig{
	DNS: []string{"223.5.5.5", "8.8.8.8"},
}
