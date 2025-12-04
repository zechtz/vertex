.PHONY: build build-release build-frontend version install clean clean-all help

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

build-frontend: ## Build the frontend web application
	@echo "Building frontend..."
	@cd web && yarn install --frozen-lockfile && yarn build
	@echo "✓ Frontend built: web/dist/"

build: build-frontend ## Build vertex with version information (includes frontend)
	@echo "Building Vertex $(VERSION) ($(COMMIT)) ..."
	@go build -ldflags="$(LDFLAGS)" -o vertex .
	@echo "✓ Build complete: ./vertex"

build-release: build-frontend ## Build release version (set VERSION=x.x.x)
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

clean: ## Remove Go build artifacts
	@rm -f vertex vertex-*
	@echo "✓ Build artifacts removed"

clean-all: clean ## Remove all build artifacts (including frontend)
	@rm -rf web/dist web/node_modules
	@echo "✓ All build artifacts removed"

# Quick development build (alias)
dev: build ## Alias for build
