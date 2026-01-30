# ffmpeg-converter Makefile
# Run 'make help' for usage information

# ==============================================================================
# Variables
# ==============================================================================

BINARY_NAME := ffmpeg-converter
BUILD_DIR := build

# Version information (from git or defaults)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Linker flags for embedding version info and stripping debug symbols
LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)

# Go commands
GO := go
GOFLAGS := -trimpath

# ==============================================================================
# Default target
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_\/-]+:.*##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ==============================================================================
# Build targets
# ==============================================================================

.PHONY: build
build: ## Build binary for current platform
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY_NAME) .

.PHONY: run
run: build ## Build and run with default arguments
	./$(BINARY_NAME) -h

.PHONY: install
install: ## Install to $GOPATH/bin
	$(GO) install $(GOFLAGS) -ldflags '$(LDFLAGS)' .

# ==============================================================================
# Cross-compilation targets
# ==============================================================================

.PHONY: build/darwin-amd64
build/darwin-amd64: ## Build for macOS Intel
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

.PHONY: build/darwin-arm64
build/darwin-arm64: ## Build for macOS Apple Silicon
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

.PHONY: build/linux-amd64
build/linux-amd64: ## Build for Linux Intel
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

.PHONY: build/linux-arm64
build/linux-arm64: ## Build for Linux ARM64
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .

.PHONY: build/windows-amd64
build/windows-amd64: ## Build for Windows Intel
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

.PHONY: build/all
build/all: build/darwin-amd64 build/darwin-arm64 build/linux-amd64 build/linux-arm64 build/windows-amd64 ## Build for all platforms
	@echo "Built binaries:"
	@ls -lh $(BUILD_DIR)/

# ==============================================================================
# Cleanup targets
# ==============================================================================

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# ==============================================================================
# Info targets
# ==============================================================================

.PHONY: version
version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
