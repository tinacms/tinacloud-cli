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
	root := s.rootCommand(version)
	args := os.Args[1:]

	err := root.Dispatch(args, nil)
	if err != nil {
		s.Track(events.CommandFailed(commandPath(args), err.Error(), args))
	}
	return err
}

func (s *State) rootCommand(version string) *command.Command {
	flagSet := flag.NewFlagSet("tinacloud", flag.ContinueOnError)
	versionFlag := flagSet.Bool("version", false, "Print the current installed version")
	flagSet.BoolVar(versionFlag, "v", false, "Alias of -version")

	root := &command.Command{
		Name:  "tinacloud",
		Flags: flagSet,
		SubCommands: []*command.Command{
			auth.Command(s.config, s.Track, s.Identify),
			org.Command(s.config),
			project.Command(s.config, s.Track),
			configure.Command(s.config, s.Track),
		},
	}

	root.Run = func(args []string) error {
		if *versionFlag {
			fmt.Println(version)
			return nil
		}
		root.PrintUsage(os.Stderr, nil)
		return nil
	}

	root.SubCommands = append(root.SubCommands, completionCommand(root))

	return root
}

// commandPath returns the command and subcommand (e.g. "project list"),
// dropping flag values so events group by command rather than by arbitrary
// user input.
func commandPath(args []string) string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return ""
	}
	if len(args) == 1 || strings.HasPrefix(args[1], "-") {
		return args[0]
	}
	return args[0] + " " + args[1]
}

func completionCommand(root *command.Command) *command.Command {
	return &command.Command{
		Name:        "completion",
		Description: "Generate shell completion scripts",
		LongDescription: `If you need to set up completions manually, pick your shell and follow the
instructions. Restart your shell afterward to confirm it worked.`,
		SubCommands: []*command.Command{
			{
				Name:        "bash",
				Description: "Generate the bash completion script",
				LongDescription: `First, ensure "bash-completion" is installed via your package manager.

Try it for the current session:
  source <(tinacloud completion bash)

Persist it by adding this to your ~/.bash_profile:
  eval "$(tinacloud completion bash)"`,
				Run: func(args []string) error { command.GenBash(root, os.Stdout); return nil },
			},
			{
				Name:        "zsh",
				Description: "Generate the zsh completion script",
				LongDescription: `Generate a "_tinacloud" completion script and put it somewhere in your $fpath:
  tinacloud completion zsh > "${fpath[1]}/_tinacloud"

Ensure the following is present in your ~/.zshrc:
  autoload -U compinit
  compinit -i

Zsh version 5.7 or later is recommended.`,
				Run: func(args []string) error { command.GenZsh(root, os.Stdout); return nil },
			},
			{
				Name:        "fish",
				Description: "Generate the fish completion script",
				LongDescription: `Generate a "tinacloud.fish" completion script:
  tinacloud completion fish > ~/.config/fish/completions/tinacloud.fish`,
				Run: func(args []string) error { command.GenFish(root, os.Stdout); return nil },
			},
			{
				Name:        "powershell",
				Description: "Generate the PowerShell completion script",
				LongDescription: `Open your profile script, creating it if necessary:
  mkdir -Path (Split-Path -Parent $profile) -ErrorAction SilentlyContinue
  notepad $profile

Add this line and save the file:
  Invoke-Expression -Command $(tinacloud completion powershell | Out-String)`,
				Run: func(args []string) error { command.GenPowerShell(root, os.Stdout); return nil },
			},
		},
	}
}
