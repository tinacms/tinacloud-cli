package auth

import (
	"tina.io/tinacloud-cli/internal/cli/auth/login"
	"tina.io/tinacloud-cli/internal/cli/auth/logout"
	"tina.io/tinacloud-cli/internal/cli/auth/status"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Command(cfg config.Config) *command.Command {
	return &command.Command{
		Name:        "auth",
		Description: "Authenticate with TinaCloud",
		SubCommands: []*command.Command{
			login.Command(cfg),
			logout.Command(cfg),
			status.Command(cfg),
		},
	}
}
