package auth

import (
	"tina.io/tinacloud-cli/internal/cli/auth/login"
	"tina.io/tinacloud-cli/internal/cli/auth/logout"
	"tina.io/tinacloud-cli/internal/cli/auth/status"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Command(cfg config.Config, track posthog.Tracker, identify posthog.Identifier) *command.Command {
	return &command.Command{
		Name:        "auth",
		Description: "Authenticate with TinaCloud",
		SubCommands: []*command.Command{
			login.Command(cfg, track, identify),
			logout.Command(cfg, track),
			status.Command(cfg),
		},
	}
}
