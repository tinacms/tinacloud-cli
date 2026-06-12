package posthog

import (
	ph "github.com/posthog/posthog-go"
)

// Tracker enqueues a telemetry event. Commands receive one from the
// CLI state so they never touch the PostHog client directly.
type Tracker func(ph.Capture)

// NoopTracker discards events; useful in tests.
func NoopTracker(ph.Capture) {}
