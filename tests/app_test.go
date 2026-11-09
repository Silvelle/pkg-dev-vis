// Package tests verifies the application's public configuration interface.
package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

const validINI = "[settings]\npackage=requests\nrepository=https://pypi.org\ntest_mode=false\n"

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.ini")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertRun(t *testing.T, path, wantOutput string, wantCode int, errorText string) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	code := app.Run(path, &out, &diagnostics)
	if code != wantCode || out.String() != wantOutput {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), diagnostics.String())
	}
	if errorText == "" && diagnostics.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", diagnostics.String())
	}
	if errorText != "" && !strings.Contains(diagnostics.String(), errorText) {
		t.Fatalf("stderr=%q, want substring %q", diagnostics.String(), errorText)
	}
}

func TestRunPrintsAllSettings(t *testing.T) {
	repository := mockGraphRepository(t, map[string]string{
		"/pypi/requests/json": `{"info":{"requires_dist":["urllib3<3,>=1.21.1"]}}`,
		"/pypi/urllib3/json":  `{"info":{"requires_dist":[]}}`,
	})
	config := strings.Replace(validINI, "https://pypi.org", repository, 1)
	want := "package=requests\nrepository=" + repository + "\ntest_mode=false\nrequests: urllib3\nurllib3: \n"
	assertRun(t, writeConfig(t, config), want, app.ExitOK, "")
}

func TestRunTestRepository(t *testing.T) {
	repository := writeConfig(t, "# graph contents are not parsed in stage 1\n")
	text := fmt.Sprintf("[settings]\npackage=A\nrepository=%s\ntest_mode=true\n", repository)
	want := fmt.Sprintf("package=A\nrepository=%s\ntest_mode=true\n", repository)
	assertRun(t, writeConfig(t, text), want, app.ExitOK, "")
}

func TestLoadConfigParsesWhitespaceAndComments(t *testing.T) {
	text := "; comment\r\n\r\n [settings] \r\n# comment\r\n" +
		" package = Requests \r\n repository = https://example.org/simple?q=a=b \r\n test_mode = false \r\n"
	parameters, err := app.LoadConfig(writeConfig(t, text))
	if err != nil {
		t.Fatal(err)
	}
	if parameters["package"] != "Requests" || parameters["repository"] != "https://example.org/simple?q=a=b" {
		t.Fatalf("unexpected settings: %#v", parameters)
	}
}

func TestRunRejectsInvalidINI(t *testing.T) {
	cases := []struct{ name, text, diagnostic string }{
		{"empty", "", "section"},
		{"no section", "package=requests\n", "line 1"},
		{"wrong section", "[other]\n", "line 1"},
		{"broken section", "[settings\n", "line 1"},
		{"empty section", "[]\n", "line 1"},
		{"repeated section", "[settings]\n[settings]\n", "repeated"},
		{"no equals", "[settings]\npackage requests\n", "line 2"},
		{"empty key", "[settings]\n=requests\n", "empty"},
		{"empty value", "[settings]\npackage=\n", "empty"},
		{"unknown key", "[settings]\nunknown=value\n", "unknown setting"},
		{"duplicate", validINI + "package=other\n", "duplicate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertRun(t, writeConfig(t, tc.text), "", app.ExitError, tc.diagnostic)
		})
	}
}

func TestRunRequiresEverySetting(t *testing.T) {
	lines := strings.Split(validINI, "\n")
	for _, key := range []string{"package", "repository", "test_mode"} {
		t.Run(key, func(t *testing.T) {
			var kept []string
			for _, line := range lines {
				if !strings.HasPrefix(line, key+"=") {
					kept = append(kept, line)
				}
			}
			assertRun(t, writeConfig(t, strings.Join(kept, "\n")), "", app.ExitError, key)
		})
	}
}

func TestRunRejectsInvalidParameters(t *testing.T) {
	cases := []struct{ old, replacement, diagnostic string }{
		{"package=requests", "package=-requests", "package:"},
		{"package=requests", "package=requests-", "package:"},
		{"package=requests", "package=request s", "package:"},
		{"package=requests", "package=пакет", "package:"},
		{"test_mode=false", "test_mode=yes", "test_mode:"},
		{"repository=https://pypi.org", "repository=local.txt", "repository:"},
		{"repository=https://pypi.org", "repository=ftp://example.org", "repository:"},
		{"repository=https://pypi.org", "repository=https:///simple", "repository:"},
		{"repository=https://pypi.org", "repository=https://example.org/%zz", "repository:"},
	}
	for _, tc := range cases {
		t.Run(tc.replacement, func(t *testing.T) {
			text := strings.Replace(validINI, tc.old, tc.replacement, 1)
			assertRun(t, writeConfig(t, text), "", app.ExitError, tc.diagnostic)
		})
	}
}

func TestLoadConfigAcceptsDistributionNames(t *testing.T) {
	for _, name := range []string{"a", "1", "3project", "my-package", "My_Package.v2", "class"} {
		t.Run(name, func(t *testing.T) {
			text := strings.Replace(validINI, "requests", name, 1)
			parameters, err := app.LoadConfig(writeConfig(t, text))
			if err != nil || parameters["package"] != name {
				t.Fatalf("settings=%v error=%v", parameters, err)
			}
		})
	}
}

func TestRunRejectsInvalidTestRepository(t *testing.T) {
	cases := []struct{ name, path string }{
		{"missing file", filepath.Join(t.TempDir(), "missing.txt")},
		{"directory", t.TempDir()},
		{"URL instead of file", "https://pypi.org"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := fmt.Sprintf("[settings]\npackage=A\nrepository=%s\ntest_mode=true\n", tc.path)
			assertRun(t, writeConfig(t, text), "", app.ExitError, "repository:")
		})
	}
}

func TestRunRejectsInvalidTestPackage(t *testing.T) {
	repository := writeConfig(t, "# fixture\n")
	text := fmt.Sprintf("[settings]\npackage=requests\nrepository=%s\ntest_mode=true\n", repository)
	assertRun(t, writeConfig(t, text), "", app.ExitError, "package:")
}

func TestRunReportsMissingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.ini")
	assertRun(t, path, "", app.ExitError, "open config:")
	_, err := app.LoadConfig(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error=%v, want wrapped os.ErrNotExist", err)
	}
}

func TestRunReportsScannerFailure(t *testing.T) {
	const oversizedLine = 70 * 1024
	text := "[settings]\npackage=" + strings.Repeat("a", oversizedLine)
	assertRun(t, writeConfig(t, text), "", app.ExitError, "read config:")
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("output unavailable")
}

func TestRunReportsOutputFailure(t *testing.T) {
	var diagnostics bytes.Buffer
	code := app.Run(writeConfig(t, validINI), failingWriter{}, &diagnostics)
	if code != app.ExitError || !strings.Contains(diagnostics.String(), "output unavailable") {
		t.Fatalf("code=%d stderr=%q", code, diagnostics.String())
	}
}
