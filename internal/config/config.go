package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type CoreConfig struct {
	Dev bool `env:"DEV, default=true"`
}

type Config struct {
	Core CoreConfig `env:""`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
