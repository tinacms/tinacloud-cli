package posthog

type Capture struct {
	Event      string
	DistinctId string
	Properties map[string]any
}

type Tracker func(Capture)

// Sender is the subset of *Client that State depends on, so tests can substitute a fake.
type Sender interface {
	Enqueue(Capture)
	Alias(distinctID, alias string)
	Close()
}

// NoopTracker discards events; useful in tests.
func NoopTracker(Capture) {}

// Identifier links the anonymous distinct ID to an authenticated user so pre-login events are attributed to them.
type Identifier func(userID string)

// NoopIdentifier discards identify calls; useful in tests.
func NoopIdentifier(string) {}
