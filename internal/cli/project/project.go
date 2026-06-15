package project

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/project/create"
	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/cli/project/status"
	"tina.io/tinacloud-cli/internal/cli/project/update"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Project(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("project", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands(os.Stderr, "tinacloud-cli project", []command.SubCommand{
			{Name: "list", Description: "List projects"},
			{Name: "create", Description: "Create a new project"},
			{Name: "status", Description: "List a project's status"},
			{Name: "update", Description: "Update a projects information"},
		})
		flagSet.PrintDefaults()
	}

	_ = flagSet.Parse(args)
	args = flagSet.Args()
	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "list":
		return list.List(subargs, cfg)
	case "create":
		return create.Create(subargs, cfg, track)
	case "status":
		return status.Status(subargs, cfg)
	case "update":
		return update.Update(subargs, cfg, track)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
