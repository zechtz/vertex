.PHONY: build build-release version install clean help

# Version information
# Try to get version from git tag, fallback to dev
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)
ifneq ($(GIT_TAG),)
    # If we're on a tag, use it (strip 'v' prefix if present)
    VERSION ?= $(shell echo $(GIT_TAG) | sed 's/^v//')
else
    # Otherwise, try to get the latest tag + commit count
    GIT_DESCRIBE := $(shell git describe --tags --always 2>/dev/null)
    ifneq ($(GIT_DESCRIBE),)
        VERSION ?= $(GIT_DESCRIBE)
    else
        VERSION ?= dev
    endif
endif

COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build vertex with version information
	@echo "Building Vertex $(VERSION) ($(COMMIT)) ..."
	@go build -ldflags="$(LDFLAGS)" -o vertex .
	@echo "✓ Build complete: ./vertex"

build-release: ## Build release version (set VERSION=x.x.x)
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: VERSION must be set for release builds"; \
		echo "Usage: make build-release VERSION=1.0.0"; \
		exit 1; \
	fi
	@echo "Building Vertex $(VERSION) ($(COMMIT)) ..."
	@go build -ldflags="$(LDFLAGS)" -o vertex .
	@echo "✓ Release build complete: ./vertex"

version: build ## Build and show version information
	@./vertex version

install: build ## Build and install vertex
	@./vertex install

clean: ## Remove build artifacts
	@rm -f vertex vertex-*
	@echo "✓ Build artifacts removed"

# Quick development build (alias)
dev: build ## Alias for build
