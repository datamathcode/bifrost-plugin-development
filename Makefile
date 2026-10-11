.PHONY: all build dev clean install help deps

# Plugin name
PLUGIN_NAME = bifrost-plugin-development
OUTPUT_DIR = build

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
 PLUGIN_EXT = .so
 PLATFORM = linux
 HOST_OS = linux
endif
ifeq ($(UNAME_S),Darwin)
 PLUGIN_EXT = .so
 PLATFORM = darwin
 HOST_OS = darwin
endif

# Architecture detection
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
 ARCH = amd64
 HOST_ARCH = amd64
endif
ifeq ($(UNAME_M),arm64)
 ARCH = arm64
 HOST_ARCH = arm64
endif
ifeq ($(UNAME_M),aarch64)
 ARCH = arm64
 HOST_ARCH = arm64
endif

# Build configuration (can be overridden via command line)
# Example: make build GOOS=linux GOARCH=amd64
GOOS ?=
GOARCH ?=

# Output file
OUTPUT = $(OUTPUT_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)

help: ## Show this help message
	@echo 'Usage: make [target] [GOOS=...] [GOARCH=...]'
	@echo ''
	@echo 'Examples:'
	@echo ' make dev          # Build for development (fast, no optimizations)'
	@echo ' make build        # Build for current platform (production)'
	@echo ' make build GOOS=linux GOARCH=amd64  # Build for Linux AMD64'
	@echo ' make build GOOS=darwin GOARCH=arm64 # Build for macOS ARM64'
	@echo ''
	@echo 'Host System:'
	@echo ' OS: $(HOST_OS)'
	@echo ' Architecture: $(HOST_ARCH)'
	@echo ''
	@echo 'Build Notes:'
	@echo ' - Native builds: Uses local Go compiler with CGO enabled'
	@echo ' - Cross-compilation: Uses Docker (Linux targets only)'
	@echo ' - macOS cross-compilation: Only works on macOS hosts'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf " %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

_clean_build_dir:
	@echo "Cleaning build directory..."
	@rm -rf $(OUTPUT_DIR)
	@echo "Build directory cleaned"

dev: _clean_build_dir ## Build the plugin for development (fast, no optimizations)
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building development plugin for $(HOST_OS)/$(HOST_ARCH)..."
	CGO_ENABLED=1 go build -buildmode=plugin -o $(OUTPUT) main.go
	@echo "Plugin built: $(OUTPUT)"

build: _clean_build_dir ## Build the plugin for production (with optimizations)
	@mkdir -p $(OUTPUT_DIR)
	@TARGET_OS="$(GOOS)"; \
	TARGET_ARCH="$(GOARCH)"; \
	ACTUAL_OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	ACTUAL_ARCH=$$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/arm64/arm64/'); \
	if [ -z "$$TARGET_OS" ]; then TARGET_OS=$$ACTUAL_OS; fi; \
	if [ -z "$$TARGET_ARCH" ]; then TARGET_ARCH=$$ACTUAL_ARCH; fi; \
	HOST_OS=$$ACTUAL_OS; \
	HOST_ARCH=$$ACTUAL_ARCH; \
	echo "Host: $$HOST_OS/$$HOST_ARCH"; \
	echo "Target: $$TARGET_OS/$$TARGET_ARCH"; \
	if [ "$$TARGET_OS" = "$$HOST_OS" ] && [ "$$TARGET_ARCH" = "$$HOST_ARCH" ]; then \
		echo "Native build..."; \
		CGO_ENABLED=1 go build -buildmode=plugin -ldflags="-w -s" -o $(OUTPUT) main.go; \
	elif [ "$$TARGET_OS" = "linux" ] && [ "$$HOST_OS" != "linux" ]; then \
		echo "Cross-compiling to Linux via Docker..."; \
		docker run --rm -v $(PWD):/src -w /src golang:1.26-bookworm \
			bash -c "CGO_ENABLED=1 GOOS=$$TARGET_OS GOARCH=$$TARGET_ARCH go build -buildmode=plugin -ldflags='-w -s' -o $(OUTPUT) main.go"; \
	elif [ "$$TARGET_OS" = "darwin" ] && [ "$$HOST_OS" = "darwin" ]; then \
		echo "macOS cross-arch build..."; \
		CGO_ENABLED=1 GOOS=$$TARGET_OS GOARCH=$$TARGET_ARCH go build -buildmode=plugin -ldflags="-w -s" -o $(OUTPUT) main.go; \
	else \
		echo "Unsupported cross-compilation: $$HOST_OS/$$HOST_ARCH -> $$TARGET_OS/$$TARGET_ARCH"; \
		exit 1; \
	fi
	@echo "Plugin built: $(OUTPUT)"

clean: ## Clean build artifacts
	@rm -rf $(OUTPUT_DIR)
	@echo "Build artifacts cleaned"

deps: ## Download and verify dependencies
	@go mod tidy
	@go mod verify
	@echo "Dependencies verified"

install: build ## Build and copy to Bifrost plugins directory (set BIFROST_PLUGIN_DIR)
	@if [ -z "$(BIFROST_PLUGIN_DIR)" ]; then \
		echo "Error: BIFROST_PLUGIN_DIR not set"; \
		echo "Usage: make install BIFROST_PLUGIN_DIR=/path/to/bifrost/plugins"; \
		exit 1; \
	fi
	@cp $(OUTPUT) $(BIFROST_PLUGIN_DIR)/
	@echo "Plugin installed to $(BIFROST_PLUGIN_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)"

test: ## Run tests (if any)
	@go test ./...

vet: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...