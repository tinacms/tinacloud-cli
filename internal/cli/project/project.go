package project

import (
	"tina.io/tinacloud-cli/internal/cli/project/create"
	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/cli/project/status"
	"tina.io/tinacloud-cli/internal/cli/project/update"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
	"tina.io/tinacloud-cli/internal/requests"
)

func Command(cfg config.Config, requestClient *requests.Client, track posthog.Tracker) *command.Command {
	return &command.Command{
		Name:        "project",
		Description: "Interact with projects",
		SubCommands: []*command.Command{
			list.Command(cfg, requestClient),
			create.Command(cfg, requestClient, track),
			status.Command(cfg, requestClient),
			update.Command(cfg, requestClient, track),
		},
	}
}
