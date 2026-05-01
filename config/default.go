package config

import (
	_ "embed"
	"fmt"

	"go.yaml.in/yaml/v3"
)

//go:embed default.yml
var defaultConfig []byte

func loadDefault() (*Config, error) {
	var cfg Config

	if err := yaml.Unmarshal(defaultConfig, &cfg); err != nil {
		return nil, fmt.Errorf("loading default config: %w", err)
	}

	return &cfg, nil
}
