package main

import (
	"context"
	"errors"
	"log"
	"os"

	"tina.io/tinacloud-cli/internal/cli"
	"tina.io/tinacloud-cli/internal/utils"
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
		var outputErr *utils.OutputError
		if errors.As(rootErr, &outputErr) {
			os.Exit(1)
		}
		log.Fatal(rootErr)
	}
}
