package cli

import (
	"context"
	"fmt"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/config"
)

type State struct {
	config  *config.Config
}

func Make(ctx context.Context) (*State, error) {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	state := &State{
		config:  cfg,
	}

	return state, nil
}

func (s *State) Close() error {
	return nil
}
