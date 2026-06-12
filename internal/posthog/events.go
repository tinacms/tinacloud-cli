package posthog

import (
	ph "github.com/posthog/posthog-go"
)

const (
	CLIStartedEvent             string = "tinacloud-cli-started"
	UserSignInInitiatedEvent    string = "tinacloud-cli-user-sign-in-initiated"
	UserSignInSuccessEvent      string = "tinacloud-cli-user-sign-in-success"
	UserLoggedOutEvent          string = "tinacloud-cli-user-logged-out"
	UserConfiguredEvent         string = "tinacloud-cli-user-configured"
	ProjectCreatedEvent         string = "tinacloud-cli-project-created"
	SiteURLAddedEvent           string = "tinacloud-cli-site-url-added"
	SiteURLRemovedEvent         string = "tinacloud-cli-site-url-removed"
	ProtectedBranchAddedEvent   string = "tinacloud-cli-protected-branch-added"
	ProtectedBranchRemovedEvent string = "tinacloud-cli-protected-branch-removed"
	CommandFailedEvent          string = "tinacloud-cli-command-failed"
)

const (
	ConfigMethodManualSelect   string = "manual-select"
	ConfigMethodAutoDetect     string = "auto-detect"
	ConfigMethodOnlyOneProject string = "only-one-project"
)

func CommandFailed(command, errMsg string, args []string) ph.Capture {
	return ph.Capture{
		Event: CommandFailedEvent,
		Properties: ph.NewProperties().
			Set("command", command).
			Set("error", errMsg).
			Set("args", args),
	}
}

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
