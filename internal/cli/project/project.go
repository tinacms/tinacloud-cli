package project

import (
	"tina.io/tinacloud-cli/internal/cli/project/create"
	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/cli/project/status"
	"tina.io/tinacloud-cli/internal/cli/project/update"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Command(cfg config.Config, track posthog.Tracker, identify posthog.Identifier) *command.Command {
	return &command.Command{
		Name:        "project",
		Description: "Interact with projects",
		SubCommands: []*command.Command{
			list.Command(cfg),
			create.Command(cfg, track),
			status.Command(cfg),
			update.Command(cfg, track),
		},
	}
}
