.PHONY: all clean build build-api build-plugin test proto

# Default target
all: build

# Build everything
build: build-api build-plugin

# Build the API server
build-api:
	@echo "Building API server..."
	@mkdir -p bin
	@go build -o bin/api cmd/api/main.go

# Build the certs plugin
build-plugin:
	@echo "Building certs plugin..."
	@mkdir -p bin
	@go build -o bin/certs-plugin cmd/certs-plugin/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Generate gRPC code
proto:
	@echo "Generating gRPC code..."
	@protoc --go_out=. --go-grpc_out=. internal/dehydrated/plugin/rpc/plugin.proto

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ 