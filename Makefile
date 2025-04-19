# Main
MAIN_FILE=cmd/api/main.go
BINARY_NAME=dehydrated-api-go
# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.Commit=$(shell git rev-parse --short HEAD) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Test
GRPC_TEST_PLUGIN_PATH=internal/plugin/grpc/testdata/test-plugin
COVERAGE_FILE=coverage.out

# Docker
DOCKER_IMAGE=schumann-it/dehydrated-api-go
DOCKER_CONTAINER=dehydrated-api-go

.PHONY: build run test test-scripts test-coverage clean lint generate release help docker-build docker-run docker-stop docker-logs docker-shell docker-clean check-tools

#
# Build and run
#

build: generate $(BINARY_NAME) ## Build binary

generate: ## Generate code
	@go generate ./...

run: $(BINARY_NAME) ## Run the binary with example config
	@./$(BINARY_NAME) -config examples/config.yaml

release: ## Create a release with goreleaser
	@goreleaser release --snapshot --clean

#
# Test
#

test: $(COVERAGE_FILE) test-scripts ## Run all tests

test-scripts: ## Run script tests
	@./scripts/test-update-config.sh

test-coverage: $(COVERAGE_FILE) ## Show coverage report
	@go tool cover -html=$(COVERAGE_FILE)

lint: ## Run linter
	@golangci-lint run

#
# Docker
#

docker-build: ## Build Docker image
	@docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@docker run -d --name $(DOCKER_CONTAINER) -p 3000:3000 $(DOCKER_IMAGE)

docker-stop: ## Stop Docker container
	@docker stop $(DOCKER_CONTAINER)
	@docker rm $(DOCKER_CONTAINER)

docker-logs: ## View Docker container logs
	@docker logs $(DOCKER_CONTAINER)

docker-shell: ## Open shell in Docker container
	@docker exec -it $(DOCKER_CONTAINER) /bin/sh

#
# Cleanup
#

clean-all: clean clean-test clean-dist clean-docker ## Cleanup everything

clean: ## Clean build artifacts
	@go clean
	@rm -f proto/plugin/*.pb.go
	@rm -f $(BINARY_NAME)

clean-test: ## Clean test
	@rm -f $(COVERAGE_FILE)
	@rm -f $(GRPC_TEST_PLUGIN_PATH)/test-plugin

clean-dist: ## Clean dist
	@rm -rf dist

clean-docker: ## Remove Docker container and image
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

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
	if [ $$MISSING_TOOLS -eq 0 ]; then \
		echo "✅ All required tools are installed"; \
	fi

#
# Files
#

$(BINARY_NAME): ## Build the binary
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

$(COVERAGE_FILE): $(GRPC_TEST_PLUGIN_PATH)/test-plugin ## Build coverage profile
	@go test -v -race -coverprofile=$(COVERAGE_FILE) ./...

$(GRPC_TEST_PLUGIN_PATH)/test-plugin: generate ## Build test plugin
	@go build -o $(GRPC_TEST_PLUGIN_PATH)/test-plugin $(GRPC_TEST_PLUGIN_PATH)/main.go
