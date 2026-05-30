.DEFAULT_GOAL := help

BIN := ./bin/inject
PKG := ./cmd/inject

.PHONY: help build test lint tidy clean release

help:  ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build:  ## Build ./bin/inject for the host platform.
	go build -o $(BIN) $(PKG)

test:  ## Run the full Go test suite locally.
	go test ./...

lint:  ## Run golangci-lint against the codebase.
	golangci-lint run ./...

tidy:  ## Run go mod tidy.
	go mod tidy

clean:  ## Remove build artifacts (./bin, ./dist).
	rm -rf ./bin ./dist

release:  ## Build cross-platform archives and publish to GitHub Releases (requires GoReleaser + GITHUB_TOKEN).
	goreleaser release --clean
