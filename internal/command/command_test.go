package command

import (
	"bytes"
	"flag"
	"slices"
	"strings"
	"testing"
)

func TestDispatch(t *testing.T) {
	var called []string

	run := func(name string) func([]string) error {
		return func(args []string) error {
			called = append(called, name+"/"+strings.Join(args, ","))
			return nil
		}
	}

	flags := flag.NewFlagSet("tinacloud", flag.ContinueOnError)
	flags.String("name", "", "project name")

	root := &Command{
		Name:  "tinacloud",
		Flags: flags,
		SubCommands: []*Command{
			{Name: "org", SubCommands: []*Command{
				{Name: "list", Run: run("list")},
			}},
			{Name: "bare", Run: run("bare")},
			{Name: "empty"},
		},
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantLog []string
	}{
		{name: "runs leaf subcommand", args: []string{"org", "list"}, wantLog: []string{"list/"}},
		{name: "flags parsed before subcommand dispatch", args: []string{"--name", "foo", "org", "list"}, wantLog: []string{"list/"}},
		{name: "help flag is handled gracefully", args: []string{"--help"}},
		{name: "help arg is handled gracefully", args: []string{"bare", "--help"}},
		{name: "command with no run prints usage", args: []string{"empty"}},
		{name: "unknown subcommand", args: []string{"bogus"}, wantErr: true},
		{name: "invalid flag", args: []string{"--bogus"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called = nil
			err := root.Dispatch(tc.args, nil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Dispatch(%v) = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if !slices.Equal(called, tc.wantLog) {
				t.Errorf("Dispatch(%v) called %v, want %v", tc.args, called, tc.wantLog)
			}
		})
	}
}

func TestDispatchDoesNotRunOnTrailingHelpOrArguments(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "trailing help", args: []string{"--project-id", "project-1", "unexpected", "--help"}},
		{name: "unexpected arguments", args: []string{"--project-id", "project-1", "unexpected", "--bogus"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("configure", flag.ContinueOnError)
			flags.String("project-id", "", "project ID")
			called := false
			cmd := &Command{
				Name:  "configure",
				Flags: flags,
				Run: func([]string) error {
					called = true
					return nil
				},
			}

			err := cmd.Dispatch(tc.args, []string{"configure"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("Dispatch(%v) = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if called {
				t.Errorf("Dispatch(%v) ran the command, want no execution", tc.args)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	t.Run("command with subcommands", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &Command{
			Name:        "project",
			Description: "Manage projects",
			SubCommands: []*Command{
				{Name: "create", Description: "Create a project"},
				{Name: "list", Description: "List projects"},
			},
		}

		cmd.PrintUsage(&buf, []string{"project"})

		want := "Manage projects\n\nUsage:\n  tinacloud project [command]\n\nAvailable Commands:\n  create  Create a project\n  list    List projects\n"
		if got := buf.String(); got != want {
			t.Errorf("PrintUsage() = %q, want %q", got, want)
		}
	})

	t.Run("command with long description and flags", func(t *testing.T) {
		var buf bytes.Buffer
		flags := flag.NewFlagSet("configure", flag.ContinueOnError)
		flags.String("tina-dir", "tina", "Path to tina folder")
		cmd := &Command{
			Name:            "configure",
			Description:     "Configure a local Tina project",
			LongDescription: "Long help text.",
			Flags:           flags,
		}

		cmd.PrintUsage(&buf, []string{"configure"})

		got := buf.String()
		for _, want := range []string{
			"Configure a local Tina project\n\nUsage:\n  tinacloud configure\n\nLong help text.\n",
			"Flags:\n",
			"-tina-dir",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("PrintUsage() output missing %q:\n%s", want, got)
			}
		}
	})
}

// genTestTree is the command tree shared by the completion generator tests.
// The project description contains an apostrophe to exercise per-shell
// escaping in the generated scripts.
func genTestTree() *Command {
	flags := flag.NewFlagSet("tinacloud", flag.ContinueOnError)
	flags.String("tina-dir", "tina", "Path to tina folder")

	return &Command{
		Name:        "tinacloud",
		Description: "Tina Cloud CLI",
		Flags:       flags,
		SubCommands: []*Command{
			{
				Name:        "org",
				Description: "Interact with organizations",
				SubCommands: []*Command{
					{Name: "list", Description: "List organizations", Run: func([]string) error { return nil }},
				},
			},
			{
				Name:        "project",
				Description: "Manage user's projects",
				SubCommands: []*Command{
					{Name: "create", Description: "Create a project", Run: func([]string) error { return nil }},
				},
			},
		},
	}
}

func TestGenBash(t *testing.T) {
	var buf bytes.Buffer
	GenBash(genTestTree(), &buf)

	want := `_tinacloud_completions() {
    local cur cmd_path
    cur="${COMP_WORDS[COMP_CWORD]}"
    cmd_path="${COMP_WORDS[*]:1:COMP_CWORD-1}"
    case "$cmd_path" in
        "")
            COMPREPLY=( $(compgen -W "org project" -- "$cur") )
            ;;
        "org")
            COMPREPLY=( $(compgen -W "list" -- "$cur") )
            ;;
        "project")
            COMPREPLY=( $(compgen -W "create" -- "$cur") )
            ;;
    esac
}
complete -F _tinacloud_completions tinacloud
`
	if got := buf.String(); got != want {
		t.Errorf("GenBash() = %q, want %q", got, want)
	}
}

func TestGenFish(t *testing.T) {
	var buf bytes.Buffer
	GenFish(genTestTree(), &buf)

	want := `function __tinacloud_using_command
    set -l cmd (commandline -opc)
    set -e cmd[1]
    test "$cmd" = "$argv"
end

complete -c tinacloud -n '__tinacloud_using_command' -a 'org' -d 'Interact with organizations'
complete -c tinacloud -n '__tinacloud_using_command' -a 'project' -d 'Manage user\'s projects'
complete -c tinacloud -n '__tinacloud_using_command' -l tina-dir -d 'Path to tina folder'
complete -c tinacloud -n '__tinacloud_using_command org' -a 'list' -d 'List organizations'
complete -c tinacloud -n '__tinacloud_using_command project' -a 'create' -d 'Create a project'
`
	if got := buf.String(); got != want {
		t.Errorf("GenFish() = %q, want %q", got, want)
	}
}

func TestGenZsh(t *testing.T) {
	var buf bytes.Buffer
	GenZsh(genTestTree(), &buf)

	want := `#compdef tinacloud

_tinacloud() {
    local -a subcmds
    subcmds=(
        'org:Interact with organizations'
        'project:Manage user'\''s projects'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' subcmds
        return
    fi

    local cmd="${words[2]}"
    shift words
    (( CURRENT-- ))

    case "$cmd" in
        org) _tinacloud_org ;;
        project) _tinacloud_project ;;
    esac
}

_tinacloud_org() {
    local -a subcmds
    subcmds=(
        'list:List organizations'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' subcmds
        return
    fi

    local cmd="${words[2]}"
    shift words
    (( CURRENT-- ))

    case "$cmd" in
    esac
}

_tinacloud_project() {
    local -a subcmds
    subcmds=(
        'create:Create a project'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' subcmds
        return
    fi

    local cmd="${words[2]}"
    shift words
    (( CURRENT-- ))

    case "$cmd" in
    esac
}

_tinacloud
`
	if got := buf.String(); got != want {
		t.Errorf("GenZsh() = %q, want %q", got, want)
	}
}

func TestGenPowerShell(t *testing.T) {
	var buf bytes.Buffer
	GenPowerShell(genTestTree(), &buf)

	want := `Register-ArgumentCompleter -Native -CommandName tinacloud -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $tokens = $commandAst.CommandElements | ForEach-Object { $_.ToString() }
    $elements = @($tokens | Select-Object -Skip 1)
    if ($wordToComplete) { $elements = @($elements | Select-Object -SkipLast 1) }
    $path = $elements -join ' '

    switch ($path) {
        '' {
            @(
                [System.Management.Automation.CompletionResult]::new('org', 'org', 'ParameterValue', 'Interact with organizations')
                [System.Management.Automation.CompletionResult]::new('project', 'project', 'ParameterValue', 'Manage user''s projects')
            ) | Where-Object { $_.CompletionText -like "$wordToComplete*" }
        }
        'org' {
            @(
                [System.Management.Automation.CompletionResult]::new('list', 'list', 'ParameterValue', 'List organizations')
            ) | Where-Object { $_.CompletionText -like "$wordToComplete*" }
        }
        'project' {
            @(
                [System.Management.Automation.CompletionResult]::new('create', 'create', 'ParameterValue', 'Create a project')
            ) | Where-Object { $_.CompletionText -like "$wordToComplete*" }
        }
    }
}
`
	if got := buf.String(); got != want {
		t.Errorf("GenPowerShell() = %q, want %q", got, want)
	}
}
