package main

import (
	"context"
	"errors"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "failed to create cli state: %v\n", err)
		os.Exit(1)
	}

	rootErr := state.Root(version)

	state.Close()

	if rootErr == nil {
		return
	}

	var outputErr *utils.OutputError
	if !errors.As(rootErr, &outputErr) {
		fmt.Fprintln(os.Stderr, rootErr)
	}
	os.Exit(1)
}
