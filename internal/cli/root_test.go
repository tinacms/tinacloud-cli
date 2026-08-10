package cli

import (
	"slices"
	"testing"

	"tina.io/tinacloud-cli/internal/command"
)

func TestCommandPath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "bare invocation", args: nil, want: ""},
		{name: "top-level command", args: []string{"auth"}, want: "auth"},
		{name: "subcommand", args: []string{"project", "list"}, want: "project list"},
		{name: "deep subcommand truncated", args: []string{"project", "update", "site-url", "add"}, want: "project update"},
		{name: "flag values excluded", args: []string{"project", "create", "--name", "secret"}, want: "project create"},
		{name: "flag directly after command", args: []string{"project", "--json"}, want: "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandPath(tt.args); got != tt.want {
				t.Errorf("commandPath(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestCompletionCommand(t *testing.T) {
	root := &command.Command{Name: "tinacloud"}
	cc := completionCommand(root)

	if cc.Name != "completion" {
		t.Errorf("completion command name = %q, want %q", cc.Name, "completion")
	}
	if cc.Description == "" {
		t.Error("completion command missing description")
	}

	var names []string
	for _, sub := range cc.SubCommands {
		names = append(names, sub.Name)
	}
	if want := []string{"bash", "zsh", "fish", "powershell"}; !slices.Equal(names, want) {
		t.Errorf("completion subcommands = %v, want %v", names, want)
	}

	for _, sub := range cc.SubCommands {
		if sub.Description == "" {
			t.Errorf("completion subcommand %q missing description", sub.Name)
		}
		if sub.Run == nil {
			t.Errorf("completion subcommand %q missing Run", sub.Name)
		}
	}
}
