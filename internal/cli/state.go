package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
)

type State struct {
	config     config.Config
	posthog    posthog.Client
	distinctID string
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
		config:     cfg,
		posthog:    ph,
		distinctID: distinctID(),
	}

	return state, nil
}

// distinctID returns the authenticated user's TinaCloud ID when
// logged in; otherwise a random per-invocation ID, so unauthenticated
// usage is not trackable across runs.
func distinctID() string {
	if tokens, err := auth.PeekTokens(); err == nil {
		if claims, err := auth.DecodeAccessToken(tokens.AccessToken); err == nil {
			if claims.TinaCloudClientID != "" {
				return claims.TinaCloudClientID
			}
			if claims.Subject != "" {
				return claims.Subject
			}
		}
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "anonymous"
	}
	return "anon-" + hex.EncodeToString(b)
}

// Track enqueues a telemetry event, attaching the persistent distinct
// ID. It is a no-op when telemetry is disabled (no PostHog API key)
// and never returns an error: telemetry must not affect command flow.
func (s *State) Track(event posthog.Capture) {
	if s.posthog == nil {
		return
	}

	event.DistinctId = s.distinctID
	_ = s.posthog.Enqueue(event)
}

func (s *State) Close() error {
	if s.posthog != nil {
		err := s.posthog.Close()
		if err != nil {
			return fmt.Errorf("failed to close posthog connection: %w", err)
		}
	}

	return nil
}
