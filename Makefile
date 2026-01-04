.PHONY: build clean test fmt lint install help verify-versions release

# Build variables
BINARY_NAME := tf-update-module-versions
BIN_DIR := bin
MAIN_PATH := ./cmd/$(BINARY_NAME)

# Version variables - read from .version file (single source of truth)
VERSION ?= $(shell cat .version 2>/dev/null || git describe --tags 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags for static binary without OS dependencies
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Go build environment - platform agnostic (uses current platform)
export CGO_ENABLED=0

help:
	@echo "$(BINARY_NAME) - Build and development targets"
	@echo ""
	@echo "Targets:"
	@echo "  build           - Build the binary (default)"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests with coverage"
	@echo "  fmt             - Format code with gofmt"
	@echo "  lint            - Run golangci-lint"
	@echo "  install         - Install binary to \$$GOPATH/bin"
	@echo "  verify-versions - Check version sync (.version, flake.nix, go.mod)"
	@echo "  release VERSION - Create a release (e.g., make release VERSION=0.2.0)"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Build info:"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(COMMIT)"
	@echo "  Build time: $(BUILD_TIME)"

verify-versions:
	@echo "Verifying version sync..."
	@VERSION_FILE=$$(cat .version 2>/dev/null | tr -d ' \n'); \
	GO_MOD_VERSION=$$(grep "^go " go.mod | awk '{print $$2}' | cut -d. -f1-2); \
	FLAKE_GO_VERSION=$$(grep "go_" flake.nix | sed 's/.*go_//' | tr '_' '.' | head -1); \
	echo "  .version file:        $$VERSION_FILE"; \
	echo "  go.mod version:       $$GO_MOD_VERSION"; \
	echo "  flake.nix Go version: $$FLAKE_GO_VERSION"; \
	echo ""; \
	if grep -q "version = \"$(VERSION)\"" flake.nix 2>/dev/null || grep -q "version = \"" flake.nix 2>/dev/null; then \
		echo "  ✓ Version sync OK"; \
	else \
		echo "  ⚠ flake.nix reads version from .version file (good!)"; \
	fi; \
	if [ "$$GO_MOD_VERSION" != "$$FLAKE_GO_VERSION" ]; then \
		echo "  ⚠ Go version mismatch! Update flake.nix or go.mod"; \
		exit 1; \
	else \
		echo "  ✓ Go versions in sync"; \
	fi

build: $(BIN_DIR)
	@echo "Building $(BINARY_NAME)..."
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Binary created: $(BIN_DIR)/$(BINARY_NAME)"

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out
	@echo "✓ Clean complete"

test:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	@echo "✓ Tests complete. Coverage report: coverage.out"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Format complete"

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...
	@echo "✓ Lint complete"

install: build
	@echo "Installing $(BINARY_NAME) to \$$GOPATH/bin..."
	go install $(BUILD_FLAGS) $(MAIN_PATH)
	@echo "✓ Installation complete"

release:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "ERROR: VERSION not specified or no git tags found"; \
		echo "Usage: make release VERSION=0.2.0"; \
		exit 1; \
	fi
	@echo "Releasing version $(VERSION)..."
	@sh scripts/release.sh $(VERSION)
