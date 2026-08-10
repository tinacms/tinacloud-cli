package siteurl

import (
	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/add"
	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/remove"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Command(cfg config.Config, track posthog.Tracker) *command.Command {
	return &command.Command{
		Name:        "site-url",
		Description: "Update a projects site URLs",
		SubCommands: []*command.Command{
			add.Command(cfg, track),
			remove.Command(cfg, track),
		},
	}
}
