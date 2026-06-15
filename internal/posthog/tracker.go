package posthog

import (
	ph "github.com/posthog/posthog-go"
)

// Tracker enqueues a telemetry event. Commands receive one from the
// CLI state so they never touch the PostHog client directly.
type Tracker func(ph.Capture)

// NoopTracker discards events; useful in tests.
func NoopTracker(ph.Capture) {}

// Identifier links the anonymous distinct ID to an authenticated user ID so
// events emitted before login are attributed to them. Like Tracker, commands
// receive one from the CLI state instead of touching the PostHog client.
type Identifier func(userID string)

// NoopIdentifier discards identify calls; useful in tests.
func NoopIdentifier(string) {}
