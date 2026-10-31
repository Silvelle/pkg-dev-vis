package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

const validINI = `[settings]
package = requests
repository = https://pypi.org
test_mode = false
`

func TestRunPrintsName(t *testing.T) {
	var out, errOut bytes.Buffer
	code := app.Run(writeConfig(t, validINI), &out, &errOut)
	if code != app.ExitOK {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), app.Name) {
		t.Fatalf("output %q does not mention %q", out.String(), app.Name)
	}
}

func TestRunPrintsAllSettings(t *testing.T) {
	var out, diagnostics bytes.Buffer

	code := app.Run(writeConfig(t, validINI), &out, &diagnostics)
	want := "package=requests\nrepository=https://pypi.org\ntest_mode=false\n"

	if code != 0 || out.String() != want || diagnostics.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q",
			code, out.String(), diagnostics.String())
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.ini")

	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	return path
}
