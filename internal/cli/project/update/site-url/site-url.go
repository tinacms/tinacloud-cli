package siteurl

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/add"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func SiteURL(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdatesiteurl", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update site-url")
		fmt.Fprintln(os.Stderr, "\nUpdate a projects site URLs")
		command.PrintSubCommands("tinacloud-cli project update", []command.SubCommand{
			{Name: "Add", Description: "Add site URLs"},
		})
	}

	_ = flagSet.Parse(args)
	args = flagSet.Args()
	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "add":
		return add.Add(subargs, cfg)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
