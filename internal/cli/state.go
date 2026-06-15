package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
)

type State struct {
	config     config.Config
	posthog    posthog.Client
	distinctID string
}

func New(ctx context.Context, posthogAPIKey, devMode string) (*State, error) {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// ldflags-embedded release values override the runtime env; empty for local builds.
	if posthogAPIKey != "" {
		cfg.Posthog.APIKey = posthogAPIKey
	}
	if devMode != "" {
		if dev, err := strconv.ParseBool(devMode); err == nil {
			cfg.Core.Dev = dev
		}
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

	// Resolve the distinct ID only when telemetry is on; it is unused otherwise.
	if state.telemetryEnabled() {
		state.distinctID = resolveDistinctID()
	}

	return state, nil
}

// resolveDistinctID returns the logged-in user's TinaCloud ID, or a random per-invocation ID when anonymous.
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

func (s *State) telemetryEnabled() bool {
	return s.posthog != nil && !s.config.Core.Dev
}

// Track enqueues a telemetry event with the distinct ID. No-op when telemetry is off; never errors so it can't affect command flow.
func (s *State) Track(event posthog.Capture) {
	if !s.telemetryEnabled() {
		return
	}

	event.DistinctId = s.distinctID
	_ = s.posthog.Enqueue(event)
}

// Identify links the anonymous distinct ID to the authenticated user so pre-login events are attributed to them.
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
