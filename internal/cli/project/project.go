package project

import (
	"tina.io/tinacloud-cli/internal/cli/project/create"
	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/cli/project/status"
	"tina.io/tinacloud-cli/internal/cli/project/update"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Command(cfg config.Config) *command.Command {
	return &command.Command{
		Name:        "project",
		Description: "Interact with projects",
		SubCommands: []*command.Command{
			list.Command(cfg),
			create.Command(cfg),
			status.Command(cfg),
			update.Command(cfg),
		},
	}
}
