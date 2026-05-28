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
	flag := flag.NewFlagSet("auth", flag.ExitOnError)
	flag.Usage = func() {
		command.PrintSubCommands("tinacloud-cli auth", []command.SubCommand{
			{Name: "login", Description: "Log in to a TinaCloud account"},
			{Name: "logout", Description: "Log out of a TinaCloud account"},
			{Name: "switch", Description: "Switch active TinaCloud account"},
			{Name: "status", Description: "Display active account and authentication status on each know TinaCloud account"},
		})
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}

	helpFlag := flag.Bool("h", false, "Show help information and exit")

	err := flag.Parse(args)
	if err != nil {
		return fmt.Errorf("Failed to parse arguments: %w", err)
	}

	if *helpFlag || len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	args = flag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a subcommand.")
	}

	cmd, subargs := args[0], args[1:]
	switch cmd {
	case "login":
		err = login.Login(subargs)
	default:
		log.Fatalf("Unrecognized command %q.", cmd)
	}
	if err != nil {
		log.Fatalf("Failed to execute %q: %v", cmd, err)
	}

	return nil
}
