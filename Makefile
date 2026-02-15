BINARY_NAME=rem
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: all build install test clean lint fmt help completions release

all: build

build: ## Build the binary (includes EventKit via cgo)
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rem/

install: build ## Install the binary to $GOPATH/bin
	go install $(LDFLAGS) ./cmd/rem/

test: ## Run tests
	go test ./... -v

test-short: ## Run tests without integration tests
	go test ./... -short -v

lint: ## Run linter
	go vet ./...

fmt: ## Format code
	go fmt ./...

release: ## Build release tarball for GitHub upload
	@mkdir -p bin
	GOARCH=arm64 go build $(LDFLAGS) -o bin/rem ./cmd/rem/
	chmod +x bin/rem
	tar -czf bin/rem-darwin-arm64.tar.gz -C bin rem
	@echo "Upload bin/rem-darwin-arm64.tar.gz to GitHub Releases"

clean: ## Remove built binaries
	rm -rf bin/

completions: build ## Generate shell completion scripts
	mkdir -p completions
	./bin/$(BINARY_NAME) completion bash > completions/rem.bash
	./bin/$(BINARY_NAME) completion zsh > completions/_rem
	./bin/$(BINARY_NAME) completion fish > completions/rem.fish

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
