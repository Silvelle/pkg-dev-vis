// Package app wires the command-line interface of pkg-dep-vis.
package app

import (
	"fmt"
	"io"
)

// Name is the program name shown in messages.
const Name = "pkg-dep-vis"

// Exit codes returned by Run.
const (
	ExitOK    = 0
	ExitError = 1
)

// Run executes the program with the given arguments and returns
// the process exit code.
func Run(configPath string, stdout, stderr io.Writer) int {
	_ = configPath
	_, err := fmt.Fprintf(stdout, "%s: dependency graph visualizer for pip packages\n", Name)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "error:", err)
		return ExitError
	}
	return ExitOK
}
