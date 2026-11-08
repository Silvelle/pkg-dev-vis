// Package app loads configuration and retrieves Python package dependencies.
package app

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Name identifies the application.
const Name = "pkg-dep-vis"

// Process exit codes returned by Run.
const (
	// ExitOK indicates that configuration loading and output succeeded.
	ExitOK = 0
	// ExitError indicates a configuration, file, or output error.
	ExitError = 1
)

// settingKeys lists required settings in their output order.
var settingKeys = []string{"package", "repository", "test_mode"}

// packageName matches valid Python distribution names without normalization.
var packageName = regexp.MustCompile(`\A[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?\z`)

// testPackageName matches names consisting of uppercase ASCII letters.
var testPackageName = regexp.MustCompile(`\A[A-Z]+\z`)

// Run prints settings and, in online mode, direct dependency declarations.
// It writes errors to stderr and returns the process exit code.
func Run(configPath string, stdout, stderr io.Writer) int {
	parameters, err := LoadConfig(configPath)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "error:", err)
		return ExitError
	}
	if _, err := fmt.Fprint(stdout, stringConfig(parameters)); err != nil {
		_, _ = fmt.Fprintln(stderr, "error:", err)
		return ExitError
	}
	if parameters["test_mode"] == "false" {
		dependencies, err := LoadDependencies(parameters["repository"], parameters["package"])
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "error:", err)
			return ExitError
		}
		for _, dependency := range dependencies {
			if _, err := fmt.Fprintln(stdout, dependency); err != nil {
				_, _ = fmt.Fprintln(stderr, "error:", err)
				return ExitError
			}
		}
	}
	return ExitOK
}

// stringConfig formats validated settings in a fixed order with a final newline.
func stringConfig(parameters map[string]string) string {
	var result strings.Builder
	for _, key := range settingKeys {
		fmt.Fprintf(&result, "%s=%s\n", key, parameters[key])
	}
	return result.String()
}

// LoadConfig parses and validates an INI file.
// It returns settings on success or an error for file, syntax, or parameter failures.
// Relative repository paths are resolved from the process working directory.
func LoadConfig(configPath string) (map[string]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = file.Close() }()
	parameters, err := parseConfig(file)
	if err != nil {
		return nil, err
	}
	if err := validateConfig(parameters); err != nil {
		return nil, err
	}
	return parameters, nil
}

// parseConfig reads a single settings section and returns its key-value pairs.
// Syntax errors include line numbers; parameter validation is left to validateConfig.
func parseConfig(reader io.Reader) (map[string]string, error) {
	parameters := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	inSettings := false
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		var err error
		inSettings, err = parseConfigLine(parameters, line, inSettings)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if !inSettings {
		return nil, fmt.Errorf("missing [settings] section")
	}
	return parameters, nil
}

// parseConfigLine handles a trimmed INI line and updates parameters for a setting.
// It returns whether the settings section has started and any syntax error.
func parseConfigLine(parameters map[string]string, line string, inSettings bool) (bool, error) {
	if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
		return inSettings, nil
	}
	if line == "[settings]" {
		if inSettings {
			return inSettings, fmt.Errorf("repeated [settings] section")
		}
		return true, nil
	}
	if strings.HasPrefix(line, "[") || !inSettings {
		return inSettings, fmt.Errorf("expected [settings] section")
	}
	return inSettings, addSetting(parameters, line)
}

// addSetting splits a key-value pair at the first equals sign and stores it.
// It rejects empty keys or values, unknown settings, and duplicate keys.
func addSetting(parameters map[string]string, line string) error {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return fmt.Errorf("expected key=value")
	}
	key, value = strings.TrimSpace(key), strings.TrimSpace(value)
	if key == "" || value == "" {
		return fmt.Errorf("key and value must not be empty")
	}
	switch key {
	case "package", "repository", "test_mode":
	default:
		return fmt.Errorf("unknown setting %q", key)
	}
	if _, exists := parameters[key]; exists {
		return fmt.Errorf("duplicate setting %q", key)
	}
	parameters[key] = value
	return nil
}

// validateConfig checks required settings and applies rules for the selected mode.
// It returns the first missing or invalid parameter error.
func validateConfig(parameters map[string]string) error {
	for _, key := range settingKeys {
		if parameters[key] == "" {
			return fmt.Errorf("missing required setting %q", key)
		}
	}
	switch parameters["test_mode"] {
	case "true":
		if !testPackageName.MatchString(parameters["package"]) {
			return fmt.Errorf("package: test names must contain uppercase ASCII letters only")
		}
		return validateTestRepository(parameters["repository"])
	case "false":
		if !packageName.MatchString(parameters["package"]) {
			return fmt.Errorf("package: invalid Python distribution name")
		}
		return validateRepositoryURL(parameters["repository"])
	default:
		return fmt.Errorf("test_mode: expected true or false")
	}
}

// validateRepositoryURL checks for an HTTP or HTTPS URL with a host.
// It validates syntax without making network requests.
func validateRepositoryURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("repository: invalid URL: %w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return fmt.Errorf("repository: expected an HTTP or HTTPS URL with a host")
	}
	return nil
}

// validateTestRepository checks that path opens for reading as a regular file.
// It does not parse the file's graph contents.
func validateTestRepository(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("repository: cannot open test file: %w", err)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("repository: cannot inspect test file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("repository: expected a regular test file")
	}
	return nil
}
