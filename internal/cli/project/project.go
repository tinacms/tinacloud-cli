package project

import (
	"flag"
	"fmt"

	"tina.io/tinacloud-cli/internal/cli/project/create"
	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/cli/project/status"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Project(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("project", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands("tinacloud-cli project", []command.SubCommand{
			{Name: "list", Description: "List projects"},
			{Name: "create", Description: "Create a new project"},
			{Name: "status", Description: "List a projects status"},
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
		return create.Create(subargs, cfg)
	case "status":
		return status.Status(subargs, cfg)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
