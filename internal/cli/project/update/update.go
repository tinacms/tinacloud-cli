package update

import (
	protectedbranch "tina.io/tinacloud-cli/internal/cli/project/update/protected-branch"
	siteurl "tina.io/tinacloud-cli/internal/cli/project/update/site-url"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	return &command.Command{
		Name:        "update",
		Description: "Update a projects information",
		SubCommands: []*command.Command{
			siteurl.Command(cfg, track),
			protectedbranch.Command(cfg, track),
		},
	}
}
