package cli

import (
	"context"
	"fmt"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/config"
)

type State struct {
	config  config.Config
	posthog posthog.Client
}

func New(ctx context.Context) (*State, error) {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var ph posthog.Client
	if cfg.Posthog.APIKey != "" {
		ph, err = posthog.NewWithConfig(
			cfg.Posthog.APIKey,
			posthog.Config{Endpoint: cfg.Posthog.Endpoint},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create posthog client: %w", err)
		}
	}

	state := &State{
		config:  cfg,
		posthog: ph,
	}

	return state, nil
}

func (s *State) Close() error {

	if s.posthog == nil {
		return nil
	}
	err := s.posthog.Close()
	if err != nil {
		return fmt.Errorf("failed to close posthog connection: %w", err)
	}

	return nil
}
