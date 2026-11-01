# Dependency graph visualizer for pip packages

Coursework, variant 19. Stage 1 implements a configurable CLI and displays its settings.
Later stages will collect dependencies, traverse them with non-recursive BFS,
find reverse dependencies, and produce PlantUML. Those stages are not implemented yet.

## Build and test

Requires Go 1.27 or newer. Uses the standard library only.

```bash
make build
make test
make check
```

`make check` also requires golangci-lint. To use a writable temporary Go cache:

```bash
GOCACHE=/tmp/pkg-dev-vis-go-cache make test
```

## Configuration

The INI file must contain exactly one `[settings]` section and all three keys:

| Key | Meaning |
| --- | --- |
| `package` | Python distribution name; uppercase ASCII letters only in test mode |
| `repository` | HTTP/HTTPS URL in normal mode; readable regular file in test mode |
| `test_mode` | Exactly `true` or `false`, lowercase |

Normal package names follow the Python Packaging name specification:
ASCII letters, digits, period, underscore and hyphen, starting and ending with a letter or digit.
See https://packaging.python.org/en/latest/specifications/name-normalization/.

Whitespace around lines, keys and values is ignored. Blank lines and full-line comments
starting with `#` or `;` are supported. Values are unquoted; inline comments are not supported.
The first `=` separates key from value. Unknown keys, duplicates, repeated sections,
missing settings and malformed lines are errors. Input uses bufio.Scanner's default token limit;
oversized lines produce a read error.

Relative paths resolve from the process working directory, including repository paths.
Normal mode validates URL syntax and scheme only; it does not contact the repository.
Test mode checks file accessibility and type, without interpreting graph contents.
The test fixture is a placeholder; graph syntax will be defined in stage 3.

## Usage and error demonstrations

Run from the project root:

```bash
go run ./src -config configs/online.ini
go run ./src -config configs/test.ini
```

Normal output:

```text
package=requests
repository=https://pypi.org
test_mode=false
```

Output order is always package, repository, test_mode, with a final newline.
Success returns exit code 0. File, INI, parameter, and output errors return exit code 1
and print a diagnostic to stderr. No settings are printed when configuration loading fails.

To demonstrate errors, copy configs/online.ini to a temporary file and try each change separately:

| Change | Expected diagnostic |
| --- | --- |
| `package=-requests` | invalid Python distribution name |
| `repository=not-a-url` | expected an HTTP or HTTPS URL with a host |
| `test_mode=yes` | expected true or false |
| Delete any required key | missing required setting |
| Use `-config missing.ini` | open config error |

In configs/test.ini, try a missing repository path or a directory to demonstrate test-file errors.
The tests exercise successful configurations in both modes and these failure categories.

## Functions

| Function | Responsibility |
| --- | --- |
| `main` | Parse the optional config-file flag and exit with Run's result |
| `Run` | Load settings, write output or diagnostics, return an exit code |
| `LoadConfig` | Open and close the INI file, parse and validate settings |
| `parseConfig` | Scan sections and lines; report syntax and reading errors |
| `parseConfigLine` | Handle comments, section placement and individual setting lines |
| `addSetting` | Split one pair; reject empty, unknown and duplicate settings |
| `validateConfig` | Check required keys and mode-specific parameter rules |
| `validateRepositoryURL` | Check URL parsing, HTTP/HTTPS scheme and host |
| `validateTestRepository` | Check that a test path opens as a regular file |
| `stringConfig` | Format settings in a stable order |

The config-file flag defaults to config.ini. User settings themselves come from INI.
Stage results should be saved in a public repository with a descriptive Conventional Commit.
