package command

import (
	"fmt"
	"os"
)

type SubCommand struct {
	Name        string
	Description string
}

func (sc SubCommand) Print() {
	fmt.Fprintf(os.Stderr, "  %-12s %s\n", sc.Name, sc.Description)
}

func PrintSubCommands(parent string, cmds []SubCommand) {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\nCommands:\n", parent)
	for _, cmd := range cmds {
		cmd.Print()
	}
}
