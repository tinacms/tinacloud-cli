package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Command struct {
	Name            string
	Description     string
	LongDescription string
	Flags           *flag.FlagSet
	SubCommands     []*Command
	Run             func(args []string) error
}

func (c *Command) Dispatch(args []string, path []string) error {
	if c.Flags != nil {
		c.Flags.Usage = func() {
			c.PrintUsage(os.Stdout, path)
		}
		if err := c.Flags.Parse(args); err != nil {
			if err == flag.ErrHelp {
				return nil
			}
			return err
		}
		args = c.Flags.Args()
	}

	if len(args) == 0 {
		if c.Run == nil {
			c.PrintUsage(os.Stdout, path)
			return nil
		}
		return c.Run(args)
	}

	name, rest := args[0], args[1:]

	if name == "-h" || name == "--help" {
		c.PrintUsage(os.Stdout, path)
		return nil
	}

	for _, sub := range c.SubCommands {
		if sub.Name == name {
			return sub.Dispatch(rest, append(path, sub.Name))
		}
	}

	if len(c.SubCommands) == 0 && c.Run != nil {
		return c.Run(args)
	}

	fmt.Fprintf(os.Stderr, "unrecognized command %q\n\n", name)
	c.PrintUsage(os.Stdout, path)
	return fmt.Errorf("unrecognized command %q", name)
}

func (c *Command) PrintUsage(w io.Writer, path []string) {
	fullPath := strings.Join(append([]string{"tinacloud"}, path...), " ")

	if c.Description != "" {
		fmt.Fprintf(w, "%s\n\n", c.Description)
	}
	fmt.Fprintf(w, "Usage:\n  %s", fullPath)
	if len(c.SubCommands) > 0 {
		fmt.Fprint(w, " [command]")
	}
	fmt.Fprintln(w)

	if c.LongDescription != "" {
		fmt.Fprintf(w, "\n%s\n", c.LongDescription)
	}

	if len(c.SubCommands) > 0 {
		fmt.Fprint(w, "\nAvailable Commands:\n")
		maxLen := 0
		for _, s := range c.SubCommands {
			if len(s.Name) > maxLen {
				maxLen = len(s.Name)
			}
		}
		for _, s := range c.SubCommands {
			fmt.Fprintf(w, "  %-*s  %s\n", maxLen, s.Name, s.Description)
		}
	}

	if c.Flags != nil {
		fmt.Fprint(w, "\nFlags:\n")
		c.Flags.SetOutput(w)
		c.Flags.PrintDefaults()
	}
}
