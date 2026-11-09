#!/bin/sh
# Run from the project root: sh scripts/demo-test.sh
make build

echo '--- Transitive dependencies and a shared dependency ---'
./bin/pkg-dep-vis -config configs/test.ini

echo '--- Cycle: A -> B -> C -> A ---'
./bin/pkg-dep-vis -config configs/test-cycle.ini

echo '--- Package without dependencies ---'
./bin/pkg-dep-vis -config configs/test-leaf.ini
