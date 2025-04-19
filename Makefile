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

# Tools
GOLANGCI_LINT_BIN=/opt/homebrew/bin/golangci-lint
GORELEASER_BIN=/opt/homebrew/bin/goreleaser
PROTOC_GEN_GO_BIN=/opt/homebrew/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC_BIN=/opt/homebrew/bin/protoc-gen-go-grpc

.PHONY: build run test test-scripts test-coverage clean lint generate release help docker-build docker-run docker-stop docker-logs docker-shell docker-clean

build: generate $(BINARY_NAME) ## Build binary

run: $(BINARY_NAME) ## Run the binary with example config
	@./$(BINARY_NAME) -config examples/config.yaml

test: $(COVERAGE_FILE) test-scripts ## Run all tests

test-scripts: ## Run script tests
	@./scripts/test-update-config.sh

test-coverage: $(COVERAGE_FILE) ## Show coverage report
	@go tool cover -html=$(COVERAGE_FILE)

# Linting
lint: ## Run linter
	@$(GOLANGCI_LINT_BIN) run

# Generate code using go generate
generate: ## Generate code
	@go generate ./...

# Release with goreleaser
release: ## Create a release with goreleaser
	@$(GORELEASER_BIN) release --snapshot --clean

# Docker targets
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

# Show help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "For more information, see the README.md file."


$(BINARY_NAME): ## Build the binary
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

$(COVERAGE_FILE): $(GRPC_TEST_PLUGIN_PATH)/test-plugin ## Build coverage profile
	@go test -v -race -coverprofile=$(COVERAGE_FILE) ./...

$(GRPC_TEST_PLUGIN_PATH)/test-plugin: ## Build test plugin
	@go build -o $(GRPC_TEST_PLUGIN_PATH)/test-plugin $(GRPC_TEST_PLUGIN_PATH)/main.go
