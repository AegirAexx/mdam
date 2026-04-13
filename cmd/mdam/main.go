package main

import (
	"fmt"
	"os"

	"github.com/AegirAexx/mdam/internal/cli"
)

// version is set at build time via:
//
//	go build -ldflags "-X main.version=v0.1.0" ./cmd/mdam
var version = "dev"

func main() {
	cli.SetVersion(version)
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
