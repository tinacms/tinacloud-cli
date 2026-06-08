package update

import (
	"flag"
	"fmt"

	siteurl "tina.io/tinacloud-cli/internal/cli/project/update/site-url"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Update(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdate", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands("tinacloud-cli project update", []command.SubCommand{
			{Name: "site-url", Description: "Update a projects site URLs"},
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
	case "site-url":
		return siteurl.SiteURL(subargs, cfg)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
