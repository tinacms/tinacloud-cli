package cli

import (
	"testing"

	"tina.io/tinacloud-cli/internal/posthog"
)

// aliasCall records the arguments of one Alias call.
type aliasCall struct {
	DistinctId string
	Alias      string
}

// captureClient records enqueued messages; it implements posthog.Sender.
type captureClient struct {
	events  []posthog.Capture
	aliases []aliasCall
}

func (c *captureClient) Enqueue(event posthog.Capture) {
	c.events = append(c.events, event)
}

func (c *captureClient) Alias(distinctID, alias string) {
	c.aliases = append(c.aliases, aliasCall{DistinctId: distinctID, Alias: alias})
}

func (c *captureClient) Close() {}

func TestTrack(t *testing.T) {
	event := posthog.Capture{Event: "test-event"}

	t.Run("enqueues with the distinct id when enabled", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "user-1", telemetryEnabled: true}

		s.Track(event)

		if len(client.events) != 1 {
			t.Fatalf("enqueued %d events, want 1", len(client.events))
		}
		if got := client.events[0]; got.Event != "test-event" || got.DistinctId != "user-1" {
			t.Errorf("enqueued %+v, want event=test-event distinctID=user-1", got)
		}
	})

	t.Run("suppresses telemetry when not opted in", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "user-1"}

		s.Track(event)

		if len(client.events) != 0 {
			t.Errorf("enqueued %d events without opt-in, want 0", len(client.events))
		}
	})

	t.Run("is a no-op when posthog cannot be initialised", func(t *testing.T) {
		// No endpoint, so lazy init fails and Track must no-op without panicking.
		s := &State{distinctID: "user-1"}

		s.Track(event)
	})
}

func TestTelemetryEnabled(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		development bool
		want        bool
	}{
		{name: "disabled by default"},
		{name: "enabled", value: "true", want: true},
		{name: "disabled", value: "false"},
		{name: "invalid", value: "yes"},
		{name: "disabled in development", value: "true", development: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TINACLOUD_TELEMETRY", tc.value)
			if got := telemetryEnabled(tc.development); got != tc.want {
				t.Errorf("telemetryEnabled(%t) = %t, want %t", tc.development, got, tc.want)
			}
		})
	}
}

func TestIdentify(t *testing.T) {
	t.Run("aliases the anonymous id to the user when enabled", func(t *testing.T) {
		client := &captureClient{}
		s := &State{posthog: client, distinctID: "anon-1", telemetryEnabled: true}

		s.Identify("user-1")

		if len(client.aliases) != 1 {
			t.Fatalf("enqueued %d aliases, want 1", len(client.aliases))
		}
		if got := client.aliases[0]; got.DistinctId != "user-1" || got.Alias != "anon-1" {
			t.Errorf("alias = %+v, want DistinctId=user-1 Alias=anon-1", got)
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
