package command

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

func GenFish(root *Command, w io.Writer) {
	fmt.Fprintf(w, "function __%s_using_command\n", root.Name)
	fmt.Fprintf(w, "    set -l cmd (commandline -opc)\n")
	fmt.Fprintf(w, "    set -e cmd[1]\n")
	fmt.Fprintf(w, "    test \"$cmd\" = \"$argv\"\n")
	fmt.Fprintf(w, "end\n\n")

	genFishLines(root, nil, root.Name, w)
}

func genFishLines(c *Command, path []string, rootName string, w io.Writer) {
	pathStr := strings.Join(path, " ")

	for _, s := range c.SubCommands {
		desc := strings.ReplaceAll(s.Description, "'", "\\'")
		if pathStr == "" {
			fmt.Fprintf(w, "complete -c %s -n '__%s_using_command' -a '%s' -d '%s'\n",
				rootName, rootName, s.Name, desc)
		} else {
			fmt.Fprintf(w, "complete -c %s -n '__%s_using_command %s' -a '%s' -d '%s'\n",
				rootName, rootName, pathStr, s.Name, desc)
		}
	}

	if c.Flags != nil {
		c.Flags.VisitAll(func(f *flag.Flag) {
			cond := pathStr
			if cond == "" {
				fmt.Fprintf(w, "complete -c %s -n '__%s_using_command' -l %s -d '%s'\n",
					rootName, rootName, f.Name, f.Usage)
			} else {
				fmt.Fprintf(w, "complete -c %s -n '__%s_using_command %s' -l %s -d '%s'\n",
					rootName, rootName, cond, f.Name, f.Usage)
			}
		})
	}

	for _, s := range c.SubCommands {
		genFishLines(s, append(path, s.Name), rootName, w)
	}
}
