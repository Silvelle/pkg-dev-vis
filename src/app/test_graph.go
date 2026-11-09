package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadTestGraph reads a test repository and traverses it using BFS.
// Each package, including leaves, must have a line such as "A: B C" or "B:".
func LoadTestGraph(path, root string) (map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open test graph: %w", err)
	}
	defer func() { _ = file.Close() }()
	graph, err := parseTestGraph(file)
	if err != nil {
		return nil, err
	}
	if err := validateTestGraph(graph); err != nil {
		return nil, err
	}
	return BuildGraph(root, func(name string) ([]string, error) {
		dependencies, found := graph[name]
		if !found {
			return nil, fmt.Errorf("package %q is not defined in test graph", name)
		}
		return dependencies, nil
	})
}

// parseTestGraph reads package definitions, skipping blank lines and comments.
func parseTestGraph(reader io.Reader) (map[string][]string, error) {
	graph := make(map[string][]string)
	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if err := addTestPackage(graph, line); err != nil {
			return nil, fmt.Errorf("test graph line %d: %w", lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read test graph: %w", err)
	}
	return graph, nil
}

// addTestPackage validates and stores a single package definition.
func addTestPackage(graph map[string][]string, line string) error {
	name, value, found := strings.Cut(line, ":")
	name = strings.TrimSpace(name)
	if !found {
		return fmt.Errorf("expected PACKAGE: DEPENDENCIES")
	}
	if !testPackageName.MatchString(name) {
		return fmt.Errorf("invalid test package name %q: expected A-Z letters", name)
	}
	if _, exists := graph[name]; exists {
		return fmt.Errorf("duplicate package %q", name)
	}
	dependencies := strings.Fields(value)
	for _, dependency := range dependencies {
		if !testPackageName.MatchString(dependency) {
			return fmt.Errorf("invalid test dependency %q: expected A-Z letters", dependency)
		}
	}
	graph[name] = dependencies
	return nil
}

// validateTestGraph checks that every referenced package has a definition.
func validateTestGraph(graph map[string][]string) error {
	for name, dependencies := range graph {
		for _, dependency := range dependencies {
			if _, found := graph[dependency]; !found {
				return fmt.Errorf("test graph: package %q references undefined package %q", name, dependency)
			}
		}
	}
	return nil
}
