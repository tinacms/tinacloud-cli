package siteurl

import (
	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/add"
	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/remove"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Command(cfg config.Config) *command.Command {
	return &command.Command{
		Name:        "site-url",
		Description: "Update a projects site URLs",
		SubCommands: []*command.Command{
			add.Command(cfg),
			remove.Command(cfg),
		},
	}
}
