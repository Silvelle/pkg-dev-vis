// Command pkg-dep-vis displays configuration for a pip dependency visualizer.
package main

import (
	"flag"
	"os"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func main() {
	configPath := flag.String("config", "config.ini", "Path to INI file")
	flag.Parse()
	os.Exit(app.Run(*configPath, os.Stdout, os.Stderr))
}
