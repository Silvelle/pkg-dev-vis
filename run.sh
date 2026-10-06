#!/bin/sh
# Builds the program and runs it with the given arguments.
# Usage: ./run.sh [arguments...]
cd "$(dirname "$0")" || exit 1
mkdir -p bin
CGO_ENABLED=0 go build -o bin/pkg-dep-vis ./src || exit 1
exec ./bin/pkg-dep-vis "$@"
