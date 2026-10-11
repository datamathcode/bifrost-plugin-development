.PHONY: build dev clean test vet fmt install help

PLUGIN_NAME = redactor-plugin
OUTPUT_DIR = build
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
 PLUGIN_EXT = .so
endif
ifeq ($(UNAME_S),Linux)
 PLUGIN_EXT = .so
endif
OUTPUT = $(OUTPUT_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf " %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Build (fast, no optimizations)
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 go build -buildmode=plugin -o $(OUTPUT) main.go
	@echo "Plugin built: $(OUTPUT)"

build: ## Build (optimized)
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 go build -buildmode=plugin -ldflags="-w -s" -o $(OUTPUT) main.go
	@echo "Plugin built: $(OUTPUT)"

test: ## Run tests
	@go test ./...

vet: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

clean: ## Remove build artifacts
	@rm -rf $(OUTPUT_DIR)

install: build ## Build and copy to ~/.bifrost/plugins
	@mkdir -p ~/.bifrost/plugins
	@cp $(OUTPUT) ~/.bifrost/plugins/
	@echo "Installed to ~/.bifrost/plugins/$(PLUGIN_NAME)$(PLUGIN_EXT)"
