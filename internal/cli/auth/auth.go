package auth

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/auth/login"
	"tina.io/tinacloud-cli/internal/cli/auth/logout"
	"tina.io/tinacloud-cli/internal/cli/auth/status"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Auth(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("auth", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands(os.Stderr, "tinacloud-cli auth", []command.SubCommand{
			{Name: "login", Description: "Log in to a TinaCloud account"},
			{Name: "logout", Description: "Log out of a TinaCloud account"},
			{Name: "status", Description: "Display authentication status and TinaCloud account information"},
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
		return login.Login(subargs, cfg)
	case "status":
		return status.Status(subargs, cfg)
	case "logout":
		return logout.Logout(subargs, cfg)
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
