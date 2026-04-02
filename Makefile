BINARY      := ldctl
MODULE      := github.com/rodmhgl/ldctl
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "\
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)"

.PHONY: build test test-integration vet lint tidy clean help

## build: Compile the binary with version info injected via ldflags.
build:
	go build $(LDFLAGS) -o $(BINARY) .

## test: Run tests with race detector enabled.
test:
	go test -race ./...

## test-integration: Build binary then run integration tests (uses build tag).
test-integration: build
	go test -race -tags integration ./...

## vet: Run go vet across all packages.
vet:
	go vet ./...

## lint: Run golangci-lint.
lint:
	golangci-lint run ./...

## tidy: Tidy and verify go modules.
tidy:
	go mod tidy
	go mod verify

## clean: Remove the compiled binary.
clean:
	rm -f $(BINARY)

## help: Show this help message.
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
