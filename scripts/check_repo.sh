#!/usr/bin/env bash
# Checks the repository against the course rules that linters do not cover:
#   С1–С5 required files and folders, С6 Conventional Commits,
#   Г1–Г5 no generated files, Р1 files <= 1000 lines, Р6 lines <= 120 chars.
# Inside a git repository it checks the files that are (or would be)
# committed: tracked files plus untracked files not ignored by .gitignore.
set -u
cd "$(dirname "$0")/.." || exit 1

MAX_FILE_LINES=1000
MAX_LINE_LENGTH=120
COMMIT_PATTERN='^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9._/-]+\))?!?: .+'
failed=0

fail() {
	echo "FAIL [$1] $2"
	failed=1
}

list_files() {
	if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
		git ls-files --cached --others --exclude-standard
	else
		find . -type f -not -path './.git/*' | sed 's|^\./||'
	fi
}

check_structure() {
	[ -f README.md ] || fail С1 "README.md is missing"
	[ -f .gitignore ] || fail С2 ".gitignore is missing"
	[ -d src ] || fail С3 "src/ folder is missing"
	[ -d tests ] || fail С4 "tests/ folder is missing"
	[ -f run.sh ] || [ -f run.bat ] || [ -f Makefile ] ||
		fail С5 "run.sh, run.bat or Makefile is missing"
}

check_commits() {
	git rev-parse --verify HEAD >/dev/null 2>&1 || return 0
	git log --format='%h %s' | while read -r hash subject; do
		case "$subject" in Merge\ * | Revert\ *) continue ;; esac
		printf '%s\n' "$subject" | grep -Eq "$COMMIT_PATTERN" ||
			echo "FAIL [С6] commit $hash: $subject"
	done | grep . && failed=1
	return 0
}

check_generated() {
	local file=$1
	case "$file" in
	*.zip | *.tar | *.gz | *.tgz | *.rar | *.7z | *.bz2 | *.xz | *.jar)
		fail Г1 "archive: $file" ;;
	*.exe | *.dll | *.so | *.dylib | *.o | *.a | *.bin)
		fail Г2 "binary: $file" ;;
	*__pycache__/* | *.pyc | *.pyo | .venv/* | venv/* | *.egg-info/*)
		fail Г3 "Python service file: $file" ;;
	bin/* | build/* | dist/* | out/* | *.test | *.out | coverage.*)
		fail Г4 "build artifact: $file" ;;
	.idea/* | .vscode/* | .vs/* | *.iml | *.swp | *~)
		fail Г5 "editor settings: $file" ;;
	esac
}

check_text() {
	local file=$1 lines long
	if ! grep -Iq . "$file" 2>/dev/null; then
		[ -s "$file" ] && fail Г2 "binary content: $file"
		return
	fi
	lines=$(wc -l <"$file")
	[ "$lines" -gt "$MAX_FILE_LINES" ] &&
		fail Р1 "$file has $lines lines (> $MAX_FILE_LINES)"
	long=$(awk -v max="$MAX_LINE_LENGTH" \
		'length($0) > max { printf "%d ", NR }' "$file")
	[ -n "$long" ] &&
		fail Р6 "$file: lines longer than $MAX_LINE_LENGTH chars: $long"
}

check_structure
check_commits
while IFS= read -r file; do
	[ -f "$file" ] || continue
	check_generated "$file"
	check_text "$file"
done < <(list_files)

if [ "$failed" -ne 0 ]; then
	echo "Repository check failed."
	exit 1
fi
echo "Repository check passed."
