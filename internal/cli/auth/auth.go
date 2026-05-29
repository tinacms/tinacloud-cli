package auth

import (
	"flag"
	"fmt"
	"log"
	"os"

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
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	help := flagSet.Bool("h", false, "Show help information and exit")
	err := flagSet.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	if *help || len(args) == 0 {
		flagSet.Usage()
		os.Exit(1)
	}

	args = flagSet.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a subcommand.")
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "login":
		err = login.Login(subargs)
	default:
		log.Fatalf("unrecognized command %q.", cmd)
	}
	if err != nil {
		return fmt.Errorf("failed to execute %q: %w", cmd, err)
	}

	return nil
}
