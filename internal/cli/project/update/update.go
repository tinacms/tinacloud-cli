package update

import (
	"flag"
	"fmt"
	"os"

	protectedbranch "tina.io/tinacloud-cli/internal/cli/project/update/protected-branch"
	siteurl "tina.io/tinacloud-cli/internal/cli/project/update/site-url"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Update(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("projectupdate", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands(os.Stderr, "tinacloud-cli project update", []command.SubCommand{
			{Name: "site-url", Description: "Update a projects site URLs"},
			{Name: "protected-branch", Description: "Update a projects protected branches. This command is only available if your project has access to the editorial workflow feature."},
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
		return siteurl.SiteURL(subargs, cfg, track)
	case "protected-branch":
		return protectedbranch.ProtectedBranch(subargs, cfg, track)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
