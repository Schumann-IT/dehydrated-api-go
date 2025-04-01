# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOWORK=$(GOCMD) work
BINARY_NAME=dehydrated-api-go
MAIN_FILE=cmd/api/main.go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

# Tools
GOLANGCI_LINT_VERSION=v1.55.2
GOLANGCI_LINT_BIN=$(shell go env GOPATH)/bin/golangci-lint
MOCKGEN_VERSION=v1.6.0
MOCKGEN_BIN=$(shell go env GOPATH)/bin/mockgen
PROTOC_GEN_GO_VERSION=v1.31.0
PROTOC_GEN_GO_BIN=$(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC_VERSION=v1.3.0
PROTOC_GEN_GO_GRPC_BIN=$(shell go env GOPATH)/bin/protoc-gen-go-grpc
GORELEASER_VERSION=v1.22.1
GORELEASER_BIN=$(shell go env GOPATH)/bin/goreleaser

.PHONY: all build test clean run deps lint mock generate release

all: deps test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOTEST) -v -race -coverprofile=coverage.out ./internal/plugin/registry/...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)
	./$(BINARY_NAME)

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Development tools installation
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

# Linting
lint:
	$(GOLANGCI_LINT_BIN) run

# Generate mocks
mock:
	$(MOCKGEN_BIN) -source=internal/plugin/interface/plugin.go -destination=internal/plugin/interface/mock_plugin.go

# Generate code using go generate
generate:
	$(GOCMD) generate ./...

# Release with goreleaser
release:
	$(GORELEASER_BIN) release --snapshot --rm-dist

# Development setup
dev-setup: install-tools deps generate mock lint

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Run deps, test, and build"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run the binary"
	@echo "  deps         - Download dependencies"
	@echo "  install-tools - Install development tools"
	@echo "  lint         - Run linter"
	@echo "  mock         - Generate mocks"
	@echo "  generate     - Generate code using go generate"
	@echo "  release      - Create a release with goreleaser"
	@echo "  dev-setup    - Setup development environment"
	@echo "  help         - Show this help message"