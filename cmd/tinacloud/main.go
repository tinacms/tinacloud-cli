package main

import (
	"errors"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/cli"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/httpclient"
	"tina.io/tinacloud-cli/internal/utils"
)

var (
	version           = "undefined"
	dev               string
	identity_endpoint string
	content_endpoint  string
)

func main() {
	httpclient.SetVersion(version)

	cfg := config.Config{
		Core: config.CoreConfig{
			Dev: dev != "false",
		},
		TinaCloud: config.TinaCloudConfig{
			IdentityEndpoint: identity_endpoint,
			ContentEndpoint:  content_endpoint,
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
