package main

import (
	"context"
	"log"

	"tina.io/tinacloud-cli/internal/cli"
)

var version = "undefined"

func main() {
	ctx := context.Background()

	state, err := cli.New(ctx)
	if err != nil {
		log.Fatalf("failed to create cli state: %v", err)
	}

	rootErr := state.Root(version)

	if err := state.Close(); err != nil {
		log.Fatalf("failed to destroy cli state: %v", err)
	}

	if rootErr != nil {
		log.Fatal(rootErr)
	}
}
