BINARY_NAME=rem
HELPER_NAME=reminders-helper
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

.PHONY: all build build-helper install test clean lint fmt help completions

all: build

build: build-helper ## Build the binary and Swift helper
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rem/

build-helper: ## Build the Swift EventKit helper for fast reads
	@mkdir -p bin
	swiftc -O -o bin/$(HELPER_NAME) internal/swift/helper.swift

install: build ## Install the binary and helper to $GOPATH/bin
	go install $(LDFLAGS) ./cmd/rem/
	cp bin/$(HELPER_NAME) $(shell go env GOPATH)/bin/$(HELPER_NAME)

test: ## Run tests
	go test ./... -v

test-short: ## Run tests without integration tests
	go test ./... -short -v

lint: ## Run linter
	go vet ./...

fmt: ## Format code
	go fmt ./...

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
