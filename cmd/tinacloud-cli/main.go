package main

import (
	"context"
	"log"

	"tina.io/tinacloud-cli/internal/cli"
)

func main() {
	ctx := context.Background()

	state, err := cli.Make(ctx)
	if err != nil {
		log.Fatal("failed to create cli state:", err)
	}
	defer func() {
		if err := state.Close(); err != nil {
			log.Fatalf("failed to destroy cli state: %v", err)
		}
	}()

	if err := state.Root(); err != nil {
		log.Fatal(err)
	}
}
