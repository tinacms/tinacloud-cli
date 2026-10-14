package utils

import (
	"os"
	"testing"
)

func TestIsTerminalPipe(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %v", err)
	}
	t.Cleanup(func() {
		reader.Close()
		writer.Close()
	})

	if IsTerminal(reader) {
		t.Error("IsTerminal(pipe) = true, want false")
	}
}
