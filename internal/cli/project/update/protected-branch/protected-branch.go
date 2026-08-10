package protectedbranch

import (
	"tina.io/tinacloud-cli/internal/cli/project/update/protected-branch/add"
	"tina.io/tinacloud-cli/internal/cli/project/update/protected-branch/remove"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Command(cfg config.Config, track posthog.Tracker, identify posthog.Identifier) *command.Command {
	return &command.Command{
		Name:        "protected-branch",
		Description: "Update a projects protected branches. This command is only available if your project has access to the editorial workflow feature.",
		SubCommands: []*command.Command{
			add.Command(cfg, track, identify),
			remove.Command(cfg, track, identify),
		},
	}
}
