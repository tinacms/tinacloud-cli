package main

import (
	"context"
	"log"

	"tina.io/tinacloud-cli/internal/cli"
)

var (
	version = "undefined"
	dev     string
)

func main() {
	ctx := context.Background()

	state, err := cli.New(ctx, dev)
	if err != nil {
		log.Fatalf("failed to create cli state: %v", err)
	}

	rootErr := state.Root(version)

	// Don't let a telemetry-flush failure mask the command's own error.
	if err := state.Close(); err != nil {
		log.Printf("warning: failed to destroy cli state: %v", err)
	}

	if rootErr != nil {
		log.Fatal(rootErr)
	}
}
