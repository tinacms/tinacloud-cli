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
		noGuess        = flagSet.Bool("no-guess", false, "Skips automated attempts to write a TinaCloud configuration from information found within the current project")
		projectID      = flagSet.String("project-id", "", "The project ID")
		organisationID = flagSet.String("org-id", "", "The organisation ID")
	)
	_ = flagSet.Parse(args)

	fmt.Println(*noGuess, *projectID, *organisationID)

	return nil
}
