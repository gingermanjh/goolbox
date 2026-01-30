# Project Root Makefile
# Run 'make help' for usage information

# ==============================================================================
# Variables
# ==============================================================================

GO := go

# List of all apps (add new apps here)
APPS := ffmpeg-converter

# ==============================================================================
# Default target
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Project-wide targets:'
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_\/-]+:.*##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ''
	@echo 'To build a specific app:'
	@echo '  make -C cmd/<app-name> build'
	@echo '  cd cmd/<app-name> && make build'

# ==============================================================================
# Test targets
# ==============================================================================

.PHONY: test
test: ## Run tests with race detection
	$(GO) test -race ./...

.PHONY: test/cover
test/cover: ## Run tests with coverage report
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: test/verbose
test/verbose: ## Run tests with verbose output
	$(GO) test -race -v ./...

# ==============================================================================
# Quality targets
# ==============================================================================

.PHONY: tidy
tidy: ## Run go mod tidy and go fmt
	$(GO) mod tidy
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && $(GO) install honnef.co/go/tools/cmd/staticcheck@latest)
	$(shell $(GO) env GOPATH)/bin/staticcheck ./...

.PHONY: audit
audit: tidy vet lint test ## Run all quality checks

# ==============================================================================
# Build targets
# ==============================================================================

.PHONY: build/all
build/all: ## Build all apps for all platforms
	@for app in $(APPS); do \
		echo "Building $$app..."; \
		$(MAKE) -C cmd/$$app build/all; \
	done

.PHONY: build
build: ## Build all apps for current platform
	@for app in $(APPS); do \
		echo "Building $$app..."; \
		$(MAKE) -C cmd/$$app build; \
	done

# ==============================================================================
# Info targets
# ==============================================================================

.PHONY: apps
apps: ## List all available apps
	@echo "Available apps:"
	@for app in $(APPS); do \
		echo "  - $$app (cmd/$$app)"; \
	done

# ==============================================================================
# Cleanup targets
# ==============================================================================

.PHONY: clean
clean: ## Remove all build artifacts
	@for app in $(APPS); do \
		$(MAKE) -C cmd/$$app clean; \
	done
	rm -f coverage.out coverage.html
