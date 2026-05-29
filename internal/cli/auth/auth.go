package auth

import (
	"flag"
	"fmt"

	"tina.io/tinacloud-cli/internal/cli/auth/login"
	"tina.io/tinacloud-cli/internal/command"
)

func Auth(args []string) error {
	flagSet := flag.NewFlagSet("auth", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands("tinacloud-cli auth", []command.SubCommand{
			{Name: "login", Description: "Log in to a TinaCloud account"},
			{Name: "logout", Description: "Log out of a TinaCloud account"},
			{Name: "switch", Description: "Switch active TinaCloud account"},
			{Name: "status", Description: "Display active account and authentication status on each know TinaCloud account"},
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
	case "login":
		return login.Login(subargs)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
