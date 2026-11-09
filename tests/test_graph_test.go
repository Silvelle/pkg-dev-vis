package tests

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func TestLoadTestGraph(t *testing.T) {
	cases := []struct {
		name, text string
		want       map[string][]string
	}{
		{"shared dependency", "# graph\n\n A: B C\r\nB: D\n; comment\nC: D\nD:\n",
			map[string][]string{"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": {}}},
		{"cycle", "A: B\nB: C\nC: A\n",
			map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"A"}}},
		{"self dependency", "A: A\n", map[string][]string{"A": {"A"}}},
		{"leaf", "A:\n", map[string][]string{"A": {}}},
		{"unreachable", "A:\nB: C\nC:\n", map[string][]string{"A": {}}},
		{"several letters", "A: BC\nBC:\n", map[string][]string{"A": {"BC"}, "BC": {}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, err := app.LoadTestGraph(writeConfig(t, tc.text), "A")
			if err != nil || !reflect.DeepEqual(graph, tc.want) {
				t.Fatalf("graph=%v error=%v, want %v", graph, err, tc.want)
			}
		})
	}
}

func TestLoadTestGraphRejectsInvalidData(t *testing.T) {
	cases := []struct{ name, text, diagnostic string }{
		{"empty", "", "not defined"},
		{"missing root", "B:\n", "not defined"},
		{"missing colon", "A B\n", "line 1"},
		{"empty name", ": B\n", "invalid test package"},
		{"lowercase name", "a:\n", "invalid test package"},
		{"lowercase dependency", "A: b\n", "invalid test dependency"},
		{"digits", "A: B1\n", "invalid test dependency"},
		{"duplicate package", "A:\nA:\n", "duplicate package"},
		{"undefined dependency", "A: B\n", "undefined package"},
		{"invalid unreachable package", "A:\nB: C\n", "undefined package"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, err := app.LoadTestGraph(writeConfig(t, tc.text), "A")
			if err == nil || graph != nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("graph=%v error=%v, want %q", graph, err, tc.diagnostic)
			}
		})
	}
}

func TestLoadTestGraphReadErrors(t *testing.T) {
	_, err := app.LoadTestGraph(filepath.Join(t.TempDir(), "missing.txt"), "A")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing file error, got %v", err)
	}
	const oversizedLine = 70 * 1024
	_, err = app.LoadTestGraph(writeConfig(t, "A: "+strings.Repeat("B", oversizedLine)), "A")
	if err == nil || !strings.Contains(err.Error(), "read test graph") {
		t.Fatalf("expected scanner error, got %v", err)
	}
}

func TestRunRejectsInvalidTestGraph(t *testing.T) {
	repository := writeConfig(t, "A: B\n")
	config := "[settings]\npackage=A\nrepository=" + repository + "\ntest_mode=true\n"
	want := "package=A\nrepository=" + repository + "\ntest_mode=true\n"
	assertRun(t, writeConfig(t, config), want, app.ExitError, "undefined package")
}
