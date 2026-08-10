package command

import (
	"fmt"
	"io"
	"strings"
)

func GenPowerShell(root *Command, w io.Writer) {
	fmt.Fprintf(w, "Register-ArgumentCompleter -Native -CommandName %s -ScriptBlock {\n", root.Name)
	fmt.Fprintf(w, "    param($wordToComplete, $commandAst, $cursorPosition)\n\n")
	fmt.Fprintf(w, "    $tokens = $commandAst.CommandElements | ForEach-Object { $_.ToString() }\n")
	fmt.Fprintf(w, "    $path = ($tokens | Select-Object -Skip 1 | Select-Object -SkipLast 1) -join ' '\n\n")
	fmt.Fprintf(w, "    switch ($path) {\n")
	genPowerShellCases(root, nil, w)
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "}\n")
}

func genPowerShellCases(c *Command, path []string, w io.Writer) {
	if len(c.SubCommands) > 0 {
		pathStr := strings.Join(path, " ")
		fmt.Fprintf(w, "        '%s' {\n", pathStr)
		fmt.Fprintf(w, "            @(\n")
		for _, s := range c.SubCommands {
			desc := strings.ReplaceAll(s.Description, "'", "''")
			fmt.Fprintf(w, "                [System.Management.Automation.CompletionResult]::new('%s', '%s', 'ParameterValue', '%s')\n",
				s.Name, s.Name, desc)
		}
		fmt.Fprintf(w, "            ) | Where-Object { $_.CompletionText -like \"$wordToComplete*\" }\n")
		fmt.Fprintf(w, "        }\n")

		for _, s := range c.SubCommands {
			genPowerShellCases(s, append(path, s.Name), w)
		}
	}
}
