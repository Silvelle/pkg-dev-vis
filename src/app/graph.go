package app

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var requirementName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*`)
var nameSeparators = regexp.MustCompile(`[-_.]+`)

// BuildGraph loads all reachable dependencies using a queue, without recursion.
// The loader returns direct package names; each package is loaded only once.
func BuildGraph(root string, load func(string) ([]string, error)) (map[string][]string, error) {
	graph := make(map[string][]string)
	queue := []string{root}
	visited := map[string]bool{root: true}
	for index := 0; index < len(queue); index++ {
		name := queue[index]
		dependencies, err := load(name)
		if err != nil {
			return nil, fmt.Errorf("dependencies of %s: %w", name, err)
		}
		graph[name] = dependencies
		for _, dependency := range dependencies {
			if !visited[dependency] {
				visited[dependency] = true
				queue = append(queue, dependency)
			}
		}
	}
	return graph, nil
}

// LoadGraph builds a graph from PyPI declarations using normalized package names.
// Version constraints, extras and environment markers are not evaluated yet.
func LoadGraph(repository, root string) (map[string][]string, error) {
	root = normalizeName(root)
	return BuildGraph(root, func(name string) ([]string, error) {
		declarations, err := LoadDependencies(repository, name)
		if err != nil {
			return nil, err
		}
		dependencies := []string{}
		seen := make(map[string]bool)
		for _, declaration := range declarations {
			dependency := requirementName.FindString(strings.TrimSpace(declaration))
			if dependency == "" {
				return nil, fmt.Errorf("invalid dependency %q", declaration)
			}
			dependency = normalizeName(dependency)
			if !seen[dependency] {
				seen[dependency] = true
				dependencies = append(dependencies, dependency)
			}
		}
		return dependencies, nil
	})
}

// normalizeName merges equivalent Python distribution name spellings.
func normalizeName(name string) string {
	return nameSeparators.ReplaceAllString(strings.ToLower(name), "-")
}

// PrintGraph prints an adjacency list in a stable alphabetical order.
func PrintGraph(writer io.Writer, graph map[string][]string) error {
	names := make([]string, 0, len(graph))
	for name := range graph {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := fmt.Fprintf(writer, "%s: %s\n", name, strings.Join(graph[name], ", ")); err != nil {
			return err
		}
	}
	return nil
}
