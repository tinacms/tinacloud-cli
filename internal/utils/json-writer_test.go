package utils

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestWriteJSONResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := struct {
			Name string `json:"name"`
		}{Name: "Acme"}

		var err error
		got := captureOutput(t, &os.Stdout, func() {
			err = WriteJSONResult(v, nil)
		})
		if err != nil {
			t.Fatalf("WriteJSONResult() = %v, want nil error", err)
		}
		if want := "{\"name\":\"Acme\"}\n"; got != want {
			t.Errorf("WriteJSONResult() wrote %q, want %q", got, want)
		}
	})

	t.Run("input error", func(t *testing.T) {
		wantErr := errors.New("boom")

		var err error
		got := captureOutput(t, &os.Stderr, func() {
			err = WriteJSONResult(nil, wantErr)
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("WriteJSONResult() = %v, want %v", err, wantErr)
		}
		var outputErr *OutputError
		if !errors.As(err, &outputErr) {
			t.Errorf("WriteJSONResult() error = %T, want *OutputError", err)
		}
		if want := "{\"error\":\"boom\"}\n"; got != want {
			t.Errorf("WriteJSONResult() wrote %q, want %q", got, want)
		}
	})

	t.Run("unmarshalable value", func(t *testing.T) {
		err := WriteJSONResult(make(chan int), nil)
		if err == nil {
			t.Fatal("WriteJSONResult() = nil error, want marshal error")
		}
	})
}

func captureOutput(t *testing.T, output **os.File, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %v", err)
	}

	original := *output
	*output = writer
	fn()
	*output = original

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("Close() = %v", err)
	}
	return string(data)
}
