package cli

import (
	"testing"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/config"
)

// captureClient records enqueued messages; it embeds posthog.Client so only Enqueue needs implementing.
type captureClient struct {
	posthog.Client
	events  []posthog.Capture
	aliases []posthog.Alias
}

func (c *captureClient) Enqueue(m posthog.Message) error {
	switch msg := m.(type) {
	case posthog.Capture:
		c.events = append(c.events, msg)
	case posthog.Alias:
		c.aliases = append(c.aliases, msg)
	}
	return nil
}

func TestTrack(t *testing.T) {
	event := posthog.Capture{Event: "test-event"}

	t.Run("enqueues with the distinct id when enabled", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "user-1"}

		s.Track(event)

		if len(client.events) != 1 {
			t.Fatalf("enqueued %d events, want 1", len(client.events))
		}
		if got := client.events[0]; got.Event != "test-event" || got.DistinctId != "user-1" {
			t.Errorf("enqueued %+v, want event=test-event distinctID=user-1", got)
		}
	})

	t.Run("suppresses telemetry in dev mode", func(t *testing.T) {
		client := &captureClient{}
		s := &State{
			posthog:    client,
			distinctID: "user-1",
			config:     config.Config{Core: config.CoreConfig{Dev: true}},
		}

		s.Track(event)

		if len(client.events) != 0 {
			t.Errorf("enqueued %d events in dev mode, want 0", len(client.events))
		}
	})

	t.Run("is a no-op when telemetry is disabled", func(t *testing.T) {
		s := State{distinctID: "user-1"}

		// Must not panic when no PostHog client is configured.
		s.Track(event)
	})
}

func TestIdentify(t *testing.T) {
	t.Run("aliases the anonymous id to the user when enabled", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "anon-1"}

		s.Identify("user-1")

		if len(client.aliases) != 1 {
			t.Fatalf("enqueued %d aliases, want 1", len(client.aliases))
		}
		if got := client.aliases[0]; got.DistinctId != "user-1" || got.Alias != "anon-1" {
			t.Errorf("alias = %+v, want DistinctId=user-1 Alias=anon-1", got)
		}
	})

	t.Run("suppresses the alias in dev mode", func(t *testing.T) {
		client := &captureClient{}
		s := &State{
			posthog:    client,
			distinctID: "anon-1",
			config:     config.Config{Core: config.CoreConfig{Dev: true}},
		}

		s.Identify("user-1")

		if len(client.aliases) != 0 {
			t.Errorf("enqueued %d aliases in dev mode, want 0", len(client.aliases))
		}
	})

	t.Run("does not alias when already attributed to the user", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "user-1"}

		s.Identify("user-1")

		if len(client.aliases) != 0 {
			t.Errorf("enqueued %d aliases for a self-alias, want 0", len(client.aliases))
		}
	})

	t.Run("does not alias an empty user id", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "anon-1"}

		s.Identify("")

		if len(client.aliases) != 0 {
			t.Errorf("enqueued %d aliases for an empty user id, want 0", len(client.aliases))
		}
	})
}
