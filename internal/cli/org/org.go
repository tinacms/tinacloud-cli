package org

import (
	"tina.io/tinacloud-cli/internal/cli/org/leaderboard"
	"tina.io/tinacloud-cli/internal/cli/org/list"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Command(cfg config.Config) *command.Command {
	return &command.Command{
		Name:        "org",
		Description: "Interact with organizations",
		SubCommands: []*command.Command{
			list.Command(cfg),
			leaderboard.Command(cfg),
		},
	}
}
