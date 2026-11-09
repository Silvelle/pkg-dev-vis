package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

func TestBuildGraphBFS(t *testing.T) {
	want := map[string][]string{
		"A": {"B", "C"}, "B": {"D"}, "C": {"D"}, "D": {},
	}
	var order []string
	graph, err := app.BuildGraph("A", func(name string) ([]string, error) {
		order = append(order, name)
		return want[name], nil
	})
	if err != nil || !reflect.DeepEqual(graph, want) {
		t.Fatalf("graph=%v error=%v, want %v", graph, err, want)
	}
	if !reflect.DeepEqual(order, []string{"A", "B", "C", "D"}) {
		t.Fatalf("unexpected BFS order: %v", order)
	}
}

func TestBuildGraphError(t *testing.T) {
	want := errors.New("repository unavailable")
	graph, err := app.BuildGraph("A", func(string) ([]string, error) {
		return nil, want
	})
	if graph != nil || !errors.Is(err, want) || !strings.Contains(err.Error(), "A") {
		t.Fatalf("graph=%v error=%v", graph, err)
	}
}

func mockGraphRepository(t *testing.T, bodies map[string]string) string {
	t.Helper()
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = testTransport(func(request *http.Request) (*http.Response, error) {
		body, found := bodies[request.URL.Path]
		if !found {
			return nil, fmt.Errorf("unexpected request: %s", request.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	return "https://example.org"
}

func TestLoadGraph(t *testing.T) {
	repository := mockGraphRepository(t, map[string]string{
		"/pypi/parent/json":    `{"info":{"requires_dist":["Child_Pkg>=1", "child.pkg[test]; extra == 'test'"]}}`,
		"/pypi/child-pkg/json": `{"info":{"requires_dist":["leaf"]}}`,
		"/pypi/leaf/json":      `{"info":{"requires_dist":null}}`,
	})
	graph, err := app.LoadGraph(repository, "Parent")
	want := map[string][]string{"parent": {"child-pkg"}, "child-pkg": {"leaf"}, "leaf": {}}
	if err != nil || !reflect.DeepEqual(graph, want) {
		t.Fatalf("graph=%v error=%v, want %v", graph, err, want)
	}
}

func TestPrintGraph(t *testing.T) {
	graph := map[string][]string{"B": {}, "A": {"B"}}
	var output bytes.Buffer
	if err := app.PrintGraph(&output, graph); err != nil || output.String() != "A: B\nB: \n" {
		t.Fatalf("output=%q error=%v", output.String(), err)
	}
	if err := app.PrintGraph(failingWriter{}, graph); err == nil {
		t.Fatal("expected output error")
	}
}
