# Main
MAIN_FILE=cmd/api/main.go
BINARY_NAME=dehydrated-api-go
EXAMPLE_PLUGIN_DIR=examples/plugins
EXAMPLE_PLUGIN_NAME=simple

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.Commit=$(shell git rev-parse --short HEAD) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Test
COVERAGE_FILE=coverage.out

.PHONY: build run test test-scripts test-coverage clean lint generate release help check-tools docker-build docker-build-local docker-build-release docker-run docker-stop docker-logs

#
# Build and run
#

build: generate $(BINARY_NAME) ## Build binary

pre-commit: build clean ## Prepare for commit

generate: ## Generate code and documentation
	@go generate ./...

run: build $(EXAMPLE_PLUGIN_DIR)/$(EXAMPLE_PLUGIN_NAME)/$(EXAMPLE_PLUGIN_NAME) ## Run the binary with example config
	@./$(BINARY_NAME) -config examples/config.yaml

release: ## Create a release with goreleaser
	@goreleaser release --snapshot --clean

#
# Test
#

test-all: test test-scripts ## Run all tests

test: $(EXAMPLE_PLUGIN_DIR)/$(EXAMPLE_PLUGIN_NAME)/$(EXAMPLE_PLUGIN_NAME) ## Run unit tests
	@go test -v ./...

test-scripts: ## Run script tests
	@./scripts/test-update-config.sh

test-coverage: $(COVERAGE_FILE) ## Show coverage report
	@go test -v -coverprofile=$(COVERAGE_FILE) -coverpkg=./cmd/...,./internal/...,./plugin/proto/config.go,./plugin/proto/metadata.go,./plugin/server/...
	@go tool cover -html=$(COVERAGE_FILE)

lint: ## Run linter
	@golangci-lint run

lint-fix: ## Run linter (and fix issues if possible)
	@golangci-lint run --fix

#
# Docker
#

docker-build: release ## Build Docker image using release artifacts
	@docker build -t dehydrated-api-go .

#
# Cleanup
#

clean-all: clean clean-test clean-dist clean-gen generate ## Cleanup everything

clean: ## Clean build artifacts
	@go clean
	@rm -f $(BINARY_NAME)
	@rm -rf .dehydrated-api-go

clean-test: ## Clean test
	@rm -f $(COVERAGE_FILE)
	@rm -f $(EXAMPLE_PLUGIN_DIR)/$(EXAMPLE_PLUGIN_NAME)/$(EXAMPLE_PLUGIN_NAME)
	@for f in $(shell find . -name 'domains.txt' | grep -v "examples/data"); do \
		rm -f $$f; \
	done
	@for d in $(shell find . -type d -name '.dehydrated-api-go' | grep -v "examples/data"); do \
		rm -rf $$d; \
	done

clean-dist: ## Clean dist
	@rm -rf dist

clean-gen: ## Clean generated files
	@rm -rf plugin/plugin*.go
	@rm -rf docs

#
# Help
#

help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "----------------------------------------"
	@$(MAKE) check-tools
	@echo "----------------------------------------"
	@echo ""
	@echo "For more information, see the README.md file."

check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@MISSING_TOOLS=0; \
	if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "❌ golangci-lint is missing. Install with: brew install golangci-lint"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "❌ goreleaser is missing. Install with: brew install goreleaser"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "❌ protoc-gen-go is missing. Install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		echo "❌ protoc-gen-go-grpc is missing. Install with: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v protoc >/dev/null 2>&1; then \
		echo "❌ protoc is missing. Install with: brew install protobuf"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v docker >/dev/null 2>&1; then \
		echo "❌ docker is missing. Install with: brew install --cask docker"; \
		MISSING_TOOLS=1; \
	fi; \
	if ! command -v $(GOPATH)/bin/swag >/dev/null 2>&1; then \
		echo "❌ swag is missing. Install with: make $(GOPATH)/bin/swag"; \
		MISSING_TOOLS=1; \
	fi; \
	if [ $$MISSING_TOOLS -eq 0 ]; then \
		echo "✅ All required tools are installed"; \
	fi

#
# Files
#

$(BINARY_NAME): ## Build the binary
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

$(EXAMPLE_PLUGIN_DIR)/$(EXAMPLE_PLUGIN_NAME)/$(EXAMPLE_PLUGIN_NAME): ## Build example plugin binary
	@cd $(EXAMPLE_PLUGIN_DIR)/$(EXAMPLE_PLUGIN_NAME) && go build -o $(EXAMPLE_PLUGIN_NAME)
