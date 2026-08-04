package siteurl

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/add"
	"tina.io/tinacloud-cli/internal/cli/project/update/site-url/remove"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func SiteURL(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("projectupdatesiteurl", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud project update site-url")
		fmt.Fprintln(os.Stderr, "\nUpdate a projects site URLs")
		command.PrintSubCommands(os.Stderr, "tinacloud project update", []command.SubCommand{
			{Name: "Add", Description: "Add site URLs"},
			{Name: "Remove", Description: "Remove site URLs"},
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
		return add.Add(subargs, cfg, track)
	case "remove":
		return remove.Remove(subargs, cfg, track)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
