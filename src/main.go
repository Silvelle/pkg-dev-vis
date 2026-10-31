// Command pkg-dep-vis visualizes the dependency graph of a pip package.
package main

import (
	"flag"
	"os"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func main() {
	configPath := flag.String(
		"config",
		"config.ini",
		"Path to INI file",
	)
	os.Exit(app.Run(*configPath, os.Stdout, os.Stderr))
}
