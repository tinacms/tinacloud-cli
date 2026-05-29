package login

import (
	"flag"
	"fmt"
	"os"
)

func Login(args []string) error {
	flagSet := flag.NewFlagSet("login", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth login [flags]")
		fmt.Fprintln(os.Stderr, "\nLog in to a TinaCloud account")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flagSet.PrintDefaults()
	}

	var (
		help           = flagSet.Bool("h", false, "Show help information and exit")
		guess          = flagSet.Bool("guess", true, "Attempts to write a TinaCloud configuration from information found within the current project")
		projectID      = flagSet.String("project-id", "", "The project ID")
		organisationID = flagSet.String("org-id", "", "The organisation ID")
	)

	err := flagSet.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	if *help {
		flagSet.Usage()
		return nil
	}

	fmt.Println(*guess, *projectID, *organisationID)

	return nil
}
