package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
)

type State struct {
	config        config.Config
	requestClient *requests.Client
	posthog       posthog.Sender
	distinctID    string
}

func New(ctx context.Context, dev string) (*State, error) {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if strings.TrimSpace(dev) != "" {
		if isDev, err := strconv.ParseBool(dev); err == nil {
			cfg.Core.Dev = isDev
		}
	}

	state := &State{
		config:        cfg,
		requestClient: requests.NewClient(),
	}

	// Resolve the distinct ID only when telemetry is on; it is unused otherwise.
	if !cfg.Core.Dev {
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

// No-op if already initialised.
func (s *State) InitialisePostHog() error {
	if s.posthog == nil {
		posthogTokens, err := s.requestClient.GetPostHogToken(s.config.TinaCloud.Endpoint)
		if err != nil {
			return fmt.Errorf("failed to get posthog tokens from tinacloud: %w", err)
		}

		s.posthog = posthog.NewClient(posthogTokens.APIKey, posthogTokens.Endpoint)
	}

	return nil
}

// Track enqueues a telemetry event with the distinct ID. No-op when telemetry is off; never errors so it can't affect command flow.
func (s *State) Track(event posthog.Capture) {
	if s.config.Core.Dev {
		return
	}

	if err := s.InitialisePostHog(); err != nil {
		return
	}

	event.DistinctId = s.distinctID
	s.posthog.Enqueue(event)
}

// Identify links the anonymous distinct ID to the authenticated user so pre-login events are attributed to them.
func (s *State) Identify(userID string) {
	if s.config.Core.Dev || userID == "" || userID == s.distinctID {
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
