package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type OutputError struct {
	err error
}

func (e *OutputError) Error() string { return e.err.Error() }
func (e *OutputError) Unwrap() error { return e.err }

func WriteJSONResult(v any, err error) error {
	if err != nil {
		data, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{err.Error()})

		fmt.Fprintln(os.Stderr, string(data))
		return &OutputError{err: err}
	}

	data, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal output: %w", marshalErr)
	}

	fmt.Fprintln(os.Stdout, string(data))
	return nil
}
