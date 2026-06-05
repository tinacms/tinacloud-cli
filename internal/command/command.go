package command

import (
	"encoding/json"
	"fmt"
	"io"
)

type SubCommand struct {
	Name        string
	Description string
}

func (sc SubCommand) fprint(w io.Writer) {
	fmt.Fprintf(w, "  %-12s %s\n", sc.Name, sc.Description)
}

func PrintSubCommands(w io.Writer, parent string, cmds []SubCommand) {
	fmt.Fprintf(w, "Usage: %s <command> [flags]\n\nCommands:\n", parent)
	for _, cmd := range cmds {
		cmd.fprint(w)
	}
}

func WriteJSONResult(w io.Writer, v any, err error) error {
	if err != nil {
		data, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{err.Error()})

		fmt.Fprintln(w, string(data))
		return err
	}

	data, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal output: %w", marshalErr)
	}

	fmt.Fprintln(w, string(data))
	return nil
}
