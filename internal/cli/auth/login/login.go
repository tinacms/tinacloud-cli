package login

import (
	"flag"
	"fmt"
	"os"
)

func Login(args []string) error {
	flag := flag.NewFlagSet("login", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth login [flags]")
		fmt.Fprintln(os.Stderr, "\nLog in to a TinaCloud account")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
	var (
		help           = flag.Bool("h", false, "Show help information and exit")
		guess          = flag.Bool("guess", true, "Attempts to write a TinaCloud configuration from information found within the current project")
		projectID      = flag.String("project-id", "", "The project ID")
		organisationID = flag.String("org-id", "", "The organisation ID")
	)
	err := flag.Parse(args)
	if err != nil {
		return fmt.Errorf("Failed to parse arguments: %w", err)
	}

	if *help {
		flag.Usage()
		return nil
	}

	fmt.Println(*guess, *projectID, *organisationID)

	return nil
}
