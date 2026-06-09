package protectedbranch

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/project/update/protected-branch/add"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func ProtectedBranch(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("projectupdateprotectedbranch", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli project update protected-branch")
		fmt.Fprintln(os.Stderr, "\nUpdate a projects protected branches. This command is only available if your project has access to the editorial workflow feature.")
		command.PrintSubCommands("tinacloud-cli project update", []command.SubCommand{
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
		return add.Add(subargs, cfg)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
