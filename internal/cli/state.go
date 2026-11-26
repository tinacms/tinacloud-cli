package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
)

type State struct {
	config           config.Config
	posthog          posthog.Sender
	distinctID       string
	telemetryEnabled bool
}

func New(cfg config.Config) *State {
	telemetryEnabled := telemetryEnabled(cfg.Core.Dev)

	state := &State{
		config:           cfg,
		telemetryEnabled: telemetryEnabled,
	}

	if telemetryEnabled {
		state.distinctID = resolveDistinctID()
	}

	return state
}

func telemetryEnabled(development bool) bool {
	enabled, err := strconv.ParseBool(os.Getenv("TINACLOUD_TELEMETRY"))
	return !development && err == nil && enabled
}

// Return the logged-in user's TinaCloud ID, or a random per-invocation ID when
// anonymous.
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

// No-op if already initialised.
func (s *State) InitialisePostHog() error {
	if s.posthog == nil {
		posthogTokens, err := requests.GetPostHogToken(s.config.TinaCloud.IdentityEndpoint)
		if err != nil {
			return fmt.Errorf("failed to get posthog tokens from tinacloud: %w", err)
		}

		s.posthog = posthog.NewClient(posthogTokens.APIKey, posthogTokens.Endpoint)
	}

	return nil
}

// Enqueue a telemetry event with the distinct ID. No-op when telemetry is off.
func (s *State) Track(event posthog.Capture) {
	if !s.telemetryEnabled {
		return
	}

	if err := s.InitialisePostHog(); err != nil {
		return
	}

	event.DistinctId = s.distinctID
	s.posthog.Enqueue(event)
}

// Link the anonymous distinct ID to the authenticated user so pre-login events
// are attributed to them.
func (s *State) Identify(userID string) {
	if !s.telemetryEnabled || userID == "" || userID == s.distinctID {
		return
	}

	if err := s.InitialisePostHog(); err != nil {
		return
	}

	s.posthog.Alias(userID, s.distinctID)
}

func (s *State) Close() {
	if s.posthog != nil {
		s.posthog.Close()
	}
}
