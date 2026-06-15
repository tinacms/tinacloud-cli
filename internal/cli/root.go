package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"tina.io/tinacloud-cli/internal/cli/auth"
	"tina.io/tinacloud-cli/internal/cli/configure"
	"tina.io/tinacloud-cli/internal/cli/org"
	"tina.io/tinacloud-cli/internal/cli/project"
	"tina.io/tinacloud-cli/internal/command"
	events "tina.io/tinacloud-cli/internal/posthog"
)

func (s *State) Root(version string) error {
	flagSet := flag.NewFlagSet("tinacloud-cli", flag.ExitOnError)
	flagSet.Usage = func() {
		command.PrintSubCommands(os.Stderr, "tinacloud-cli", []command.SubCommand{
			{Name: "auth", Description: "Authenticate with TinaCloud"},
			{Name: "org", Description: "Interact with organizations"},
			{Name: "project", Description: "Interact with projects"},
			{Name: "configure", Description: "Configure a local Tina project"},
		})
		fmt.Fprint(os.Stderr, "\nFlags:\n")
		flagSet.PrintDefaults()
	}

	versionFlag := flagSet.Bool("version", false, "Print the current installed version")
	flagSet.BoolVar(versionFlag, "v", false, "Alias of -version")
	_ = flagSet.Parse(os.Args[1:])

	if *versionFlag {
		fmt.Println(version)
		return nil
	}

	args := flagSet.Args()
	s.Track(events.CLIStarted(commandPath(args)))
	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	cmd, subargs := args[0], args[1:]
	var err error
	switch cmd {
	case "auth":
		err = auth.Auth(subargs, s.config, s.Track, s.Identify)
	case "org":
		err = org.Org(subargs, s.config)
	case "project":
		err = project.Project(subargs, s.config, s.Track)
	case "configure":
		err = configure.Configure(subargs, s.config, s.Track)
	default:
		err = fmt.Errorf("unrecognised command: %q", cmd)
	}

	if err != nil {
		s.Track(events.CommandFailed(commandPath(args), err.Error(), args))
	}

	return err
}

// commandPath reduces the invocation to its command and subcommand
// (e.g. "project list") so the property stays low-cardinality; flag
// values are deliberately excluded here (full argv is reported
// separately on command failures).
func commandPath(args []string) string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return ""
	}
	if len(args) == 1 || strings.HasPrefix(args[1], "-") {
		return args[0]
	}
	return args[0] + " " + args[1]
}
