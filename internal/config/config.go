package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type CoreConfig struct {
	Dev bool `env:"DEV, default=true"`
}

type PosthogConfig struct {
	ApiKey   string `env:"API_KEY"`
	Endpoint string `env:"ENDPOINT, default=https://us.i.posthog.com"`
}

type Config struct {
	Core    CoreConfig    `env:""`
	Posthog PosthogConfig `env:",prefix=POSTHOG_"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
