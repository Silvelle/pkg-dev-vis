package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// LoadDependencies retrieves direct dependency declarations from a PyPI-compatible repository.
func LoadDependencies(repository, name string) ([]string, error) {
	address, err := url.JoinPath(repository, "pypi", name, "json")
	if err != nil {
		return nil, fmt.Errorf("dependency URL: %w", err)
	}
	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Get(address)
	if err != nil {
		return nil, fmt.Errorf("request dependencies: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request dependencies: %s", response.Status)
	}
	var data struct {
		Info struct {
			RequiresDist []string `json:"requires_dist"`
		} `json:"info"`
	}
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode dependencies: %w", err)
	}
	return data.Info.RequiresDist, nil
}
