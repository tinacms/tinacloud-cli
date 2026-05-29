package cli

import (
	"context"
	"fmt"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/config"
)

type State struct {
	config  *config.Config
	posthog *posthog.Client
}

func Make(ctx context.Context) (*State, error) {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	posthogClient := new(posthog.Client)
	if cfg.Posthog.ApiKey != "" {
		ph, err := posthog.NewWithConfig(
			cfg.Posthog.ApiKey,
			posthog.Config{Endpoint: cfg.Posthog.Endpoint},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create posthog client: %w", err)
		}
		posthogClient = &ph
	}

	state := &State{
		config:  cfg,
		posthog: posthogClient,
	}

	return state, nil
}

func (s *State) Close() error {
	if s.posthog != nil {
		err := (*s.posthog).Close()
		if err != nil {
			return err
		}
	}
	return nil
}
