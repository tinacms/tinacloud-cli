package project

import (
	"flag"
	"fmt"

	"tina.io/tinacloud-cli/internal/cli/project/list"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Project(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("org", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands("tinacloud-cli project", []command.SubCommand{
			{Name: "list", Description: "List projects"},
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
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
