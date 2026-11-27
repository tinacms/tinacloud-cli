package leaderboard

import (
	shared "tina.io/tinacloud-cli/internal/cli/leaderboard"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/requests"
)

func Command(cfg config.Config) *command.Command {
	return shared.Command(cfg, requests.Project)
}
