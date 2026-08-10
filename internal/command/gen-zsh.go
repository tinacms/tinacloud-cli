package command

import (
	"fmt"
	"io"
	"strings"
)

func GenZsh(root *Command, w io.Writer) {
	fmt.Fprintf(w, "#compdef %s\n\n", root.Name)
	genZshFunc(root, []string{root.Name}, w)
	fmt.Fprintf(w, "_%s\n", root.Name)
}

func genZshFunc(c *Command, path []string, w io.Writer) {
	funcName := "_" + strings.Join(path, "_")

	fmt.Fprintf(w, "%s() {\n", funcName)
	fmt.Fprintf(w, "    local -a subcmds\n")
	fmt.Fprintf(w, "    subcmds=(\n")
	for _, s := range c.SubCommands {
		desc := strings.ReplaceAll(s.Description, "'", "'\\''")
		fmt.Fprintf(w, "        '%s:%s'\n", s.Name, desc)
	}
	fmt.Fprintf(w, "    )\n\n")

	if len(c.SubCommands) > 0 {
		fmt.Fprintf(w, "    if (( CURRENT == 2 )); then\n")
		fmt.Fprintf(w, "        _describe 'command' subcmds\n")
		fmt.Fprintf(w, "        return\n")
		fmt.Fprintf(w, "    fi\n\n")
		fmt.Fprintf(w, "    local cmd=\"${words[2]}\"\n")
		fmt.Fprintf(w, "    shift words\n")
		fmt.Fprintf(w, "    (( CURRENT-- ))\n\n")
		fmt.Fprintf(w, "    case \"$cmd\" in\n")
		for _, s := range c.SubCommands {
			if len(s.SubCommands) > 0 {
				subFunc := funcName + "_" + s.Name
				fmt.Fprintf(w, "        %s) %s ;;\n", s.Name, subFunc)
			}
		}
		fmt.Fprintf(w, "    esac\n")
	}
	fmt.Fprintf(w, "}\n\n")

	for _, s := range c.SubCommands {
		if len(s.SubCommands) > 0 {
			genZshFunc(s, append(path, s.Name), w)
		}
	}
}
