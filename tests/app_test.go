package tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func TestRunPrintsName(t *testing.T) {
	var out, errOut bytes.Buffer
	code := app.Run(nil, &out, &errOut)
	if code != app.ExitOK {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), app.Name) {
		t.Fatalf("output %q does not mention %q", out.String(), app.Name)
	}
}
