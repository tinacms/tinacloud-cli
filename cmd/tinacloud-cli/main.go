package main

import (
	"context"
	"log"

	"tina.io/tinacloud-cli/internal/cli"
)

// Injected at release time via ldflags; empty for local/source builds.
var (
	version       = "undefined"
	posthogAPIKey string
	devMode       string
)

func main() {
	ctx := context.Background()

	state, err := cli.New(ctx, posthogAPIKey, devMode)
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
