package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

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
