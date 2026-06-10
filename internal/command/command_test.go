package command

import (
	"bytes"
	"errors"
	"testing"
)

func TestPrintSubCommands(t *testing.T) {
	var buf bytes.Buffer
	PrintSubCommands(&buf, "tinacloud-cli org", []SubCommand{
		{Name: "list", Description: "List organizations"},
	})

	want := "Usage: tinacloud-cli org <command> [flags]\n\nCommands:\n  list         List organizations\n"
	if got := buf.String(); got != want {
		t.Errorf("PrintSubCommands() = %q, want %q", got, want)
	}
}

func TestWriteJSONResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		v := struct {
			Name string `json:"name"`
		}{Name: "Acme"}

		if err := WriteJSONResult(&buf, v, nil); err != nil {
			t.Fatalf("WriteJSONResult() = %v, want nil error", err)
		}
		if want := "{\"name\":\"Acme\"}\n"; buf.String() != want {
			t.Errorf("WriteJSONResult() wrote %q, want %q", buf.String(), want)
		}
	})

	t.Run("input error", func(t *testing.T) {
		var buf bytes.Buffer
		wantErr := errors.New("boom")

		err := WriteJSONResult(&buf, nil, wantErr)
		if !errors.Is(err, wantErr) {
			t.Fatalf("WriteJSONResult() = %v, want %v", err, wantErr)
		}
		if want := "{\"error\":\"boom\"}\n"; buf.String() != want {
			t.Errorf("WriteJSONResult() wrote %q, want %q", buf.String(), want)
		}
	})

	t.Run("unmarshalable value", func(t *testing.T) {
		var buf bytes.Buffer

		err := WriteJSONResult(&buf, make(chan int), nil)
		if err == nil {
			t.Fatal("WriteJSONResult() = nil error, want marshal error")
		}
		if buf.Len() != 0 {
			t.Errorf("WriteJSONResult() wrote %q on marshal failure, want nothing", buf.String())
		}
	})
}
