package cli

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli/auth"
	"tina.io/tinacloud-cli/internal/command"
)

func (s *State) Root() error {
	flag.Usage = func() {
		command.PrintSubCommands("tinacloud-cli", []command.SubCommand{
			{Name: "auth", Description: "Authenticate with TinaCloud"},
		})
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}

	help := flag.Bool("h", false, "Show help information and exit")
	flag.Parse()

	args := flag.Args()
	if *help || len(args) == 0 {
		flag.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "auth":
		return auth.Auth(subargs)
	default:
		return fmt.Errorf("unrecognised command: %q", cmd)
	}
}
