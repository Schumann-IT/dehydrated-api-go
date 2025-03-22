# Dehydrated API Go

A Go implementation of a gRPC-based API for managing dehydrated certificates. This project provides a robust and extensible way to interact with dehydrated certificates through a plugin system.

## Overview

Dehydrated API Go is designed to provide a standardized interface for managing dehydrated certificates through a gRPC-based API. It allows for easy integration with various certificate management systems and provides a plugin architecture for extending functionality.

### Key Features

- **gRPC-based API**: High-performance, type-safe communication protocol
- **Plugin System**: Extensible architecture for adding new certificate management capabilities
- **Domain Management**: CRUD operations for certificate domains
- **Metadata Support**: Rich metadata handling for domains and certificates
- **File System Integration**: Automatic file watching and updates
- **Configuration Management**: Flexible configuration system with environment support

## Architecture

The project is structured into several key components:

- **API Layer**: Handles HTTP/gRPC requests and responses
- **Service Layer**: Implements business logic and domain management
- **Plugin System**: Provides extensibility through gRPC plugins
- **File System**: Manages certificate files and configurations
- **Configuration**: Handles application settings and environment variables

### Plugin System

The plugin system allows for extending the functionality of the API through gRPC plugins. Plugins can:
- Provide custom certificate management logic
- Implement specific validation rules
- Add new metadata fields
- Integrate with external systems

## Getting Started

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (protoc)
- gRPC tools

### Installation

1. Clone the repository:
```bash
git clone https://github.com/schumann-it/dehydrated-api-go.git
cd dehydrated-api-go
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build ./...
```

### Configuration

The application can be configured through environment variables or a configuration file:

```yaml
# config.yaml
base_dir: "/path/to/dehydrated"
domains_file: "domains.txt"
```

Environment variables:
```bash
export DEHYDRATED_BASE_DIR="/path/to/dehydrated"
export DEHYDRATED_DOMAINS_FILE="domains.txt"
```

## Development

### Project Structure

```
.
├── cmd/
│   └── api/           # Main application entry point
├── internal/
│   ├── handler/       # HTTP/gRPC request handlers
│   ├── model/         # Data models and validation
│   └── service/       # Business logic implementation
├── pkg/
│   └── dehydrated/    # Dehydrated integration package
├── plugin/
│   ├── grpc/         # gRPC plugin implementation
│   ├── interface/    # Plugin interface definitions
│   └── registry/     # Plugin registry and management
└── proto/            # Protocol Buffer definitions
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Building Plugins

Plugins are implemented as separate gRPC servers. Example plugin structure:

```go
package main

import (
    "context"
    pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

type pluginServer struct {
    pb.UnimplementedPluginServer
}

func (s *pluginServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
    // Implementation
}

func (s *pluginServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
    // Implementation
}

func (s *pluginServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
    // Implementation
}
```

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

1. **Code Style**
   - Follow Go standard formatting (`go fmt`)
   - Use meaningful variable and function names
   - Add comments for complex logic
   - Keep functions focused and small

2. **Testing**
   - Write unit tests for new features
   - Maintain or improve test coverage
   - Include integration tests for API changes
   - Test error cases and edge conditions

3. **Documentation**
   - Update README.md for significant changes
   - Document new features and APIs
   - Include examples for new functionality
   - Update API documentation

4. **Commit Messages**
   - Use clear and descriptive commit messages
   - Reference issues when applicable
   - Keep commits focused and atomic

### Pull Request Process

1. Update the README.md with details of changes if needed
2. Update the documentation for any new features
3. Ensure all tests pass
4. Request review from maintainers

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [dehydrated](https://github.com/dehydrated-io/dehydrated) - The original dehydrated project
- [gRPC](https://grpc.io/) - The gRPC framework
- [Protocol Buffers](https://developers.google.com/protocol-buffers) - The serialization format 