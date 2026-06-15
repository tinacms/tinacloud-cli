package cli

import (
	"testing"

	"github.com/posthog/posthog-go"
	"tina.io/tinacloud-cli/internal/config"
)

// captureClient records enqueued events. It embeds posthog.Client so it only
// has to implement Enqueue; Track calls nothing else.
type captureClient struct {
	posthog.Client
	events []posthog.Capture
}

func (c *captureClient) Enqueue(m posthog.Message) error {
	if event, ok := m.(posthog.Capture); ok {
		c.events = append(c.events, event)
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
		s := &State{distinctID: "user-1"}

		// Must not panic when no PostHog client is configured.
		s.Track(event)
	})
}
