package posthog

import (
	ph "github.com/posthog/posthog-go"
)

// Event names. Past-tense, snake_case.
const (
	EventCLIStarted     = "cli_started"
	EventUserLoggedIn   = "user_logged_in"
	EventUserLoggedOut  = "user_logged_out"
	EventProjectCreated = "project_created"
	EventCommandFailed  = "command_failed"
	EventProtectedBranchAdded = "protected_branch_added"
	EventProtectedBranchRemoved = "protected_branch_removed"
)

func CLIStarted(command, version string) ph.Capture {
	return ph.Capture{
		Event: EventCLIStarted,
		Properties: ph.NewProperties().
			Set("command", command).
			Set("version", version),
	}
}

func UserLoggedIn() ph.Capture {
	return ph.Capture{
		Event: EventUserLoggedIn,
	}
}

func UserLoggedOut() ph.Capture {
	return ph.Capture{
		Event: EventUserLoggedOut,
	}
}

func ProjectCreated(orgID string) ph.Capture {
	return ph.Capture{
		Event: EventProjectCreated,
		Properties: ph.NewProperties().
			Set("org_id", orgID),
	}
}

func CommandFailed(command, errorKind string) ph.Capture {
	return ph.Capture{
		Event: EventCommandFailed,
		Properties: ph.NewProperties().
			Set("command", command).
			Set("error_kind", errorKind),
	}
}
