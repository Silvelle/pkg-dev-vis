BIN := bin/pkg-dep-vis
SRC := ./src

# Pure Go build: no cgo, no system libraries, runs without Docker.
export CGO_ENABLED := 0

.PHONY: build run test lint check setup clean

build:
	mkdir -p bin
	go build -o $(BIN) $(SRC)

run:
	go run $(SRC) $(ARGS)

test:
	go test -v ./...

lint:
	golangci-lint run ./...

check: test lint
	./scripts/check_repo.sh

setup:
	git config core.hooksPath .githooks

clean:
	rm -rf bin
