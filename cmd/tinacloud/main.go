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

	state.Close()

	if rootErr != nil {
		log.Fatal(rootErr)
	}
}
