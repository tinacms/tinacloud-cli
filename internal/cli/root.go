package cli

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/auth"
	"tina.io/tinacloud-cli/internal/command"
)

func (s *State) Root() error {
	flagSet := flag.NewFlagSet("tinacloud-cli", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands("tinacloud-cli", []command.SubCommand{
			{Name: "auth", Description: "Authenticate with TinaCloud"},
		})
		flagSet.PrintDefaults()
	}

	_ = flagSet.Parse(os.Args[1:])
	args := flagSet.Args()
	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "auth":
		return auth.Auth(subargs, s.config)
	default:
		return fmt.Errorf("unrecognised command: %q", cmd)
	}
}
