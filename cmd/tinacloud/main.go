package main

import (
	"errors"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/utils"
)

var (
	version  = "undefined"
	dev      string
	endpoint string
)

func main() {
	cfg := config.Config{
		Core: config.CoreConfig{
			Dev: dev != "false",
		},
		TinaCloud: config.TinaCloudConfig{
			Endpoint: endpoint,
		},
	}

	state := cli.New(cfg)

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
