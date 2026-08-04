package protectedbranch

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/project/update/protected-branch/add"
	"tina.io/tinacloud-cli/internal/cli/project/update/protected-branch/remove"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func ProtectedBranch(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("projectupdateprotectedbranch", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud project update protected-branch")
		fmt.Fprintln(os.Stderr, "\nUpdate a projects protected branches. This command is only available if your project has access to the editorial workflow feature.")
		command.PrintSubCommands(os.Stderr, "tinacloud project update", []command.SubCommand{
			{Name: "Add", Description: "Add protected branches"},
			{Name: "Remove", Description: "Remove protected branches"},
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
