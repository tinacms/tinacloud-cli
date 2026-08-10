package command

import (
	"fmt"
	"io"
	"strings"
)

func GenBash(root *Command, w io.Writer) {
	fmt.Fprintf(w, "_%s_completions() {\n", root.Name)
	fmt.Fprintf(w, "    local cur cmd_path\n")
	fmt.Fprintf(w, "    cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	fmt.Fprintf(w, "    cmd_path=\"${COMP_WORDS[*]:1:COMP_CWORD-1}\"\n")
	fmt.Fprintf(w, "    case \"$cmd_path\" in\n")
	genBashCases(root, nil, w)
	fmt.Fprintf(w, "    esac\n}\n")
	fmt.Fprintf(w, "complete -F _%s_completions %s\n", root.Name, root.Name)
}

func genBashCases(c *Command, path []string, w io.Writer) {
	if len(c.SubCommands) > 0 {
		names := make([]string, len(c.SubCommands))
		for i, s := range c.SubCommands {
			names[i] = s.Name
		}
		fmt.Fprintf(w, "        \"%s\")\n", strings.Join(path, " "))
		fmt.Fprintf(w, "            COMPREPLY=( $(compgen -W \"%s\" -- \"$cur\") )\n", strings.Join(names, " "))
		fmt.Fprintf(w, "            ;;\n")
		for _, s := range c.SubCommands {
			genBashCases(s, append(path, s.Name), w)
		}
	}
}
