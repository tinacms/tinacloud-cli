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
		config:  cfg,
		posthog: ph,
	}

	// Resolve the distinct ID only when telemetry will be sent; in dev mode or
	// without an API key it is never used, so skip the token lookup entirely.
	if state.telemetryEnabled() {
		state.distinctID = resolveDistinctID()
	}

	return state, nil
}

// resolveDistinctID returns the authenticated user's TinaCloud ID when
// logged in; otherwise a random per-invocation ID, so unauthenticated
// usage is not trackable across runs.
func resolveDistinctID() string {
	if tokens, err := auth.PeekTokens(); err == nil {
		if claims, err := auth.DecodeAccessToken(tokens.AccessToken); err == nil {
			if claims.TinaCloudUserID != "" {
				return claims.TinaCloudUserID
			}
		}
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "anonymous"
	}
	return "anon-" + hex.EncodeToString(b)
}

// telemetryEnabled reports whether events should actually be sent: a PostHog
// client must be configured and dev mode must be off.
func (s *State) telemetryEnabled() bool {
	return s.posthog != nil && !s.config.Core.Dev
}

// Track enqueues a telemetry event, attaching the persistent distinct
// ID. It is a no-op in dev mode or when telemetry is disabled (no
// PostHog API key) and never returns an error: telemetry must not
// affect command flow.
func (s *State) Track(event posthog.Capture) {
	if !s.telemetryEnabled() {
		return
	}

	event.DistinctId = s.distinctID
	_ = s.posthog.Enqueue(event)
}

// Identify links the current anonymous distinct ID to the authenticated user,
// so events emitted before login (e.g. sign-in) are attributed to them. Like
// Track, it is a no-op in dev mode or when telemetry is disabled and never
// returns an error.
func (s *State) Identify(userID string) {
	if !s.telemetryEnabled() || userID == "" || userID == s.distinctID {
		return
	}

	_ = s.posthog.Enqueue(posthog.Alias{
		DistinctId: userID,
		Alias:      s.distinctID,
	})
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
