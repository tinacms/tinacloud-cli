package posthog

import (
	ph "github.com/posthog/posthog-go"
)

const (
	CLIStartedEvent             string = "cli-started"
	UserSignInInitiatedEvent    string = "user-sign-in-initiated"
	UserSignInSuccessEvent      string = "user-sign-in-success"
	UserLoggedOutEvent          string = "user-logged-out"
	UserConfiguredEvent         string = "user-configured"
	ProjectCreatedEvent         string = "project-created"
	SiteURLAddedEvent           string = "site-url-added"
	SiteURLRemovedEvent         string = "site-url-removed"
	ProtectedBranchAddedEvent   string = "protected-branch-added"
	ProtectedBranchRemovedEvent string = "protected-branch-removed"
)

const (
	ConfigMethodManualSelect   string = "manual-select"
	ConfigMethodAutoDetect     string = "auto-detect"
	ConfigMethodOnlyOneProject string = "only-one-project"
)

func CLIStarted(command string) ph.Capture {
	return ph.Capture{
		Event: CLIStartedEvent,
		Properties: ph.NewProperties().
			Set("command", command),
	}
}

func UserSignInInitiated() ph.Capture {
	return ph.Capture{
		Event: UserSignInInitiatedEvent,
	}
}

func UserSignInSuccess() ph.Capture {
	return ph.Capture{
		Event: UserSignInSuccessEvent,
	}
}

func UserLoggedOut() ph.Capture {
	return ph.Capture{
		Event: UserLoggedOutEvent,
	}
}

func UserConfigured(method string) ph.Capture {
	return ph.Capture{
		Event: UserConfiguredEvent,
		Properties: ph.NewProperties().
			Set("configuration_method", method),
	}
}

func ProjectCreated(orgID, projectID, name, repoURL, siteURL string) ph.Capture {
	return ph.Capture{
		Event: ProjectCreatedEvent,
		Properties: ph.NewProperties().
			Set("org_id", orgID).
			Set("project_id", projectID).
			Set("name", name).
			Set("repo_url", repoURL).
			Set("site_url", siteURL),
	}
}

func SiteURLAdded(urls []string) ph.Capture {
	return ph.Capture{
		Event: SiteURLAddedEvent,
		Properties: ph.NewProperties().
			Set("urls", urls),
	}
}

func SiteURLRemoved(urls []string) ph.Capture {
	return ph.Capture{
		Event: SiteURLRemovedEvent,
		Properties: ph.NewProperties().
			Set("urls", urls),
	}
}

func ProtectedBranchAdded(branches []string) ph.Capture {
	return ph.Capture{
		Event: ProtectedBranchAddedEvent,
		Properties: ph.NewProperties().
			Set("branches", branches),
	}
}

func ProtectedBranchRemoved(branches []string) ph.Capture {
	return ph.Capture{
		Event: ProtectedBranchRemovedEvent,
		Properties: ph.NewProperties().
			Set("branches", branches),
	}
}
