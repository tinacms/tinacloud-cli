package main

import (
	"log"

	"tina.io/tinacloud-cli/internal/cli"
)

func main() {
	if err := cli.Root(); err != nil {
		log.Fatal(err)
	}
}
