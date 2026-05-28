package cli

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/command"
)

func Root() error {
	helpFlag := flag.Bool("h", false, "Show help information and exit")
	flag.Usage = func() {
		command.PrintSubCommands("tinacloud-cli", []command.SubCommand{
			{Name: "auth", Description: "Authenticate with TinaCloud"},
		})
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	if *helpFlag || len(args) == 0 {
		flag.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	_ = subargs

	switch cmd {
	default:
		return fmt.Errorf("unrecognized command %q", cmd)
	}
}
