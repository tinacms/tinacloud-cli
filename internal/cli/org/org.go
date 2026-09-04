package org

import (
	"tina.io/tinacloud-cli/internal/cli/org/list"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/requests"
)

func Command(cfg config.Config, requestClient *requests.Client) *command.Command {
	return &command.Command{
		Name:        "org",
		Description: "Interact with organizations",
		SubCommands: []*command.Command{
			list.Command(cfg, requestClient),
		},
	}
}
