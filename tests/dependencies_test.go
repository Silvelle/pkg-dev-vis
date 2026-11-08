package tests

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Silvelle/pkg-dep-vis/src/app"
)

type testTransport func(*http.Request) (*http.Response, error)

func (transport testTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func mockRepository(t *testing.T, path, body string, status int) string {
	t.Helper()
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = testTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host != "example.org" || request.URL.Path != path {
			t.Errorf("unexpected request URL: %s", request.URL)
		}
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	return "https://example.org"
}

func TestLoadDependencies(t *testing.T) {
	cases := []struct {
		name, body, errorText string
		status                int
		want                  []string
	}{
		{"requirements", `{"info":{"requires_dist":["child>=1; python_version < '3.12'",` +
			`"optional; extra == 'test'"]}}`, "", http.StatusOK,
			[]string{"child>=1; python_version < '3.12'", "optional; extra == 'test'"}},
		{"empty", `{"info":{"requires_dist":[]}}`, "", http.StatusOK, []string{}},
		{"null", `{"info":{"requires_dist":null}}`, "", http.StatusOK, nil},
		{"missing package", "", "404", http.StatusNotFound, nil},
		{"invalid JSON", "not JSON", "decode dependencies", http.StatusOK, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repository := mockRepository(t, "/mirror/pypi/parent/json", tc.body, tc.status)
			got, err := app.LoadDependencies(repository+"/mirror/", "parent")
			if tc.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errorText) {
					t.Fatalf("error=%v, want %q", err, tc.errorText)
				}
			} else if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("dependencies=%v error=%v, want %v", got, err, tc.want)
			}
		})
	}
}
