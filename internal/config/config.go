package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type CoreConfig struct {
	Dev bool `env:"DEV, default=true"`
}

type TinaCloudConfig struct {
	Endpoint         string `env:"ENDPOINT, default=https://identity-v2.tinajs.io"`
	FrontEndEndpoint string `env:"FRONT_END_ENDPOINT, default=https://app.tina.io"`
}

type Config struct {
	Core      CoreConfig      `env:""`
	TinaCloud TinaCloudConfig `env:",prefix=TINACLOUD_"`
}

func LoadConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to process env config: %w", err)
	}

	return cfg, nil
}
