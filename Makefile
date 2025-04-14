# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOWORK=$(GOCMD) work
GOVET=$(GOCMD) vet
BINARY_NAME=dehydrated-api-go
COVERAGE_FILE=coverage.out
DOCKER_IMAGE=schumann-it/dehydrated-api-go
DOCKER_CONTAINER=dehydrated-api-go-container
MAIN_FILE=cmd/api/main.go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

# Tools
GOLANGCI_LINT_BIN=/opt/homebrew/bin/golangci-lint
GORELEASER_BIN=/opt/homebrew/bin/goreleaser
PROTOC_GEN_GO_BIN=/opt/homebrew/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC_BIN=/opt/homebrew/bin/protoc-gen-go-grpc

.PHONY: all build test clean run deps lint generate release help docker-build docker-run docker-stop docker-logs docker-shell docker-clean proto

all: deps test build ## Run deps, test, and build

build: generate ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

test: test-app test-scripts

test-app: generate-test
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE)

test-scripts:
	./scripts/test-update-api-config.sh
	./scripts/test-update-dehydrated-config.sh
	./scripts/test-configure-cron.sh

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f proto/plugin/*.pb.go
	rm -f internal/plugin/grpc/testdata/test-plugin/test-plugin
	rm -rf dist

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOGET) -v -t -d ./...

# Linting
lint: ## Run linter
	$(GOLANGCI_LINT_BIN) run

# Generate code using go generate
generate: ## Generate code using go generate
	$(GOCMD) generate ./...

# Generate code for tests using go generate
generate-test: ## Generate code for tests using go generate
	$(GOCMD) generate -tags test ./...

# Release with goreleaser
release: ## Create a release with goreleaser
	$(GORELEASER_BIN) release --snapshot --clean

# Development setup
dev-setup: deps generate-test lint

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -d --name $(DOCKER_CONTAINER) -p 3000:3000 $(DOCKER_IMAGE)

docker-stop: ## Stop Docker container
	docker stop $(DOCKER_CONTAINER)
	docker rm $(DOCKER_CONTAINER)

docker-logs: ## View Docker container logs
	docker logs $(DOCKER_CONTAINER)

docker-shell: ## Open shell in Docker container
	docker exec -it $(DOCKER_CONTAINER) /bin/sh

docker-clean: ## Remove Docker container and image
	docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# Show help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "For more information, see the README.md file."