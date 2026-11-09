#!/bin/sh
# Run from the project root: sh scripts/demo.sh
make build

echo '--- requests: dependency graph ---'
./bin/pkg-dep-vis -config configs/online.ini

echo '--- idna: dependency declarations ---'
./bin/pkg-dep-vis -config configs/no-dependencies.ini

echo '--- test mode: dependency graph from a file ---'
./bin/pkg-dep-vis -config configs/test.ini
