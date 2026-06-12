package cli

import "testing"

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
