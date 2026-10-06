// Command pkg-dep-vis visualizes the dependency graph of a pip package.
package main

import (
	"os"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
