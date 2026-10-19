package posthog

const (
	UserConfiguredEvent string = "tinacloud-cli-user-configured"
	CommandFailedEvent  string = "tinacloud-cli-command-failed"
)

// ConfigMethod describes how a project was chosen during configure.
type ConfigMethod string

const (
	ConfigMethodManualSelect   ConfigMethod = "manual-select"
	ConfigMethodAutoDetect     ConfigMethod = "auto-detect"
	ConfigMethodOnlyOneProject ConfigMethod = "only-one-project"
)

func UserConfigured(method ConfigMethod) Capture {
	return Capture{
		Event:      UserConfiguredEvent,
		Properties: map[string]any{"configuration_method": method},
	}
}
func CommandFailed(command, errMsg string, args []string) Capture {
	return Capture{Event: CommandFailedEvent, Properties: map[string]any{"command": command, "error": errMsg, "args": args}}
}
