package posthog

import (
	ph "github.com/posthog/posthog-go"
)

// Tracker enqueues a telemetry event, decoupling commands from the PostHog client.
type Tracker func(ph.Capture)

// NoopTracker discards events; useful in tests.
func NoopTracker(ph.Capture) {}

// Identifier links the anonymous distinct ID to an authenticated user so pre-login events are attributed to them.
type Identifier func(userID string)

// NoopIdentifier discards identify calls; useful in tests.
func NoopIdentifier(string) {}
