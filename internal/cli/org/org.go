package org

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/org/list"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Org(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("org", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands(os.Stderr, "tinacloud-cli org", []command.SubCommand{
			{Name: "list", Description: "List organizations"},
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
