package posthog

const (
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

// ConfigMethod describes how a project was chosen during configure.
type ConfigMethod string

const (
	ConfigMethodManualSelect   ConfigMethod = "manual-select"
	ConfigMethodAutoDetect     ConfigMethod = "auto-detect"
	ConfigMethodOnlyOneProject ConfigMethod = "only-one-project"
)

func CommandFailed(command, errMsg string, args []string) Capture {
	return Capture{
		Event: CommandFailedEvent,
		Properties: map[string]any{
			"command": command,
			"error":   errMsg,
			"args":    args,
		},
	}
}

func UserSignInInitiated() Capture {
	return Capture{Event: UserSignInInitiatedEvent}
}

func UserSignInSuccess() Capture {
	return Capture{Event: UserSignInSuccessEvent}
}

func UserLoggedOut() Capture {
	return Capture{Event: UserLoggedOutEvent}
}

func UserConfigured(method ConfigMethod) Capture {
	return Capture{
		Event:      UserConfiguredEvent,
		Properties: map[string]any{"configuration_method": method},
	}
}

func ProjectCreated(orgID, projectID, name, repoURL, siteURL string) Capture {
	return Capture{
		Event: ProjectCreatedEvent,
		Properties: map[string]any{
			"org_id":     orgID,
			"project_id": projectID,
			"name":       name,
			"repo_url":   repoURL,
			"site_url":   siteURL,
		},
	}
}

func SiteURLAdded(urls []string) Capture {
	return Capture{
		Event:      SiteURLAddedEvent,
		Properties: map[string]any{"urls": urls},
	}
}

func SiteURLRemoved(urls []string) Capture {
	return Capture{
		Event:      SiteURLRemovedEvent,
		Properties: map[string]any{"urls": urls},
	}
}

func ProtectedBranchAdded(branches []string) Capture {
	return Capture{
		Event:      ProtectedBranchAddedEvent,
		Properties: map[string]any{"branches": branches},
	}
}

func ProtectedBranchRemoved(branches []string) Capture {
	return Capture{
		Event:      ProtectedBranchRemovedEvent,
		Properties: map[string]any{"branches": branches},
	}
}
