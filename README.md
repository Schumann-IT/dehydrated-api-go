# Dehydrated API Go

A REST API server for managing domains with [dehydrated](https://github.com/dehydrated-io/dehydrated), the ACME client
for Let's Encrypt certificates. This API provides a clean interface for domain management with optional authentication
and a plugin system for extensibility.

## Key Features

- **RESTful API**: Full CRUD operations for domain management
- **Authentication**: Optional Azure AD integration with JWT validation
- **Plugin System**: Extensible architecture using gRPC plugins
- **File Watching**: Real-time domain configuration monitoring
- **Structured Logging**: Comprehensive logging with Zap
- **Swagger Documentation**: Auto-generated API documentation
- **Docker Support**: Containerized deployment with minimal runtime dependencies
- **Health Checks**: Built-in health monitoring

## Technology Stack

- **Framework**: [Fiber](https://gofiber.io/) - Fast HTTP framework
- **Authentication**: Azure AD with JWT tokens
- **Logging**: [Zap](https://github.com/uber-go/zap) - Structured logging
- **Plugin System**: [gRPC](https://grpc.io/) with [Hashicorp go-plugin](https://github.com/hashicorp/go-plugin)
- **Documentation**: [Swagger](https://swagger.io/) with [swaggo](https://github.com/swaggo/swag)
- **Configuration**: YAML-based configuration
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or later
- Docker (optional)
- Make (for build automation)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/schumann-it/dehydrated-api-go.git
   cd dehydrated-api-go
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the application**:
   ```bash
   make build
   ```

4. **Run with example configuration**:
   ```bash
   make run
   ```

The API will be available at `http://localhost:3000`

## 🐳 Docker

### Docker Overview

The Dockerfile is designed to use binary artifacts from goreleaser instead of building from source code. This approach
provides several benefits:

- **Faster builds**: No compilation time required
- **Consistent artifacts**: Uses the same binaries as official releases
- **Smaller images**: No build tools or source code included
- **Security**: Uses pre-built, tested binaries

### Build Scenarios

Build an image using a snapshot release:

```bash
make docker-build
```

For easy local development:

```bash
docker-compose up -d
```

### Running the Container

#### Basic Usage

```bash
docker run -d \
  --name dehydrated-api-go \
  -p 3000:3000 \
  -v /path/to/config.yaml:/app/config/config.yaml \
  -v /path/to/data:/data/dehydrated \
  dehydrated-api-go:latest
```

#### Using docker-compose

```bash
# Using docker-compose
docker-compose up -d

# Using Makefile
make docker-run
```

### Environment Variables

The container supports the following environment variables:

| Variable | Default | Description      |
|----------|---------|------------------|
| `PORT`   | 3000    | HTTP server port |

### Development Workflow

For development, you can use the snapshot build:

1. Make your changes
2. Build the local binary: `make build`
3. Build the Docker image: `docker build --build-arg VERSION=snapshot -t dehydrated-api-go:dev .`
4. Test your changes: `docker run dehydrated-api-go:dev`

## 📖 Configuration

### Basic Configuration

Create a `config.yaml` file:

```yaml
port: 3000
dehydratedBaseDir: ./data
enableWatcher: true
logging:
  level: debug
  encoding: console
  outputPath: ""
plugins:
  example:
    enabled: true
    registry:
      type: local
      config:
        path: ./examples/plugins/simple/simple
    config:
      name: example
```

### Authentication Configuration

For Azure AD authentication, add the `auth` section:

```yaml
auth:
  tenantId: "your-tenant-id"
  clientId: "00000003-0000-0000-c000-000000000000"
  authority: "https://login.microsoftonline.com/your-tenant-id"
  allowedAudiences:
    - "https://graph.microsoft.com"
    - "https://your-domain.com/your-client-id"
```

### Configuration Options

| Option               | Type   | Default   | Description                          |
|----------------------|--------|-----------|--------------------------------------|
| `port`               | int    | 3000      | HTTP server port                     |
| `dehydratedBaseDir`  | string | `./data`  | Base directory for dehydrated data   |
| `enableWatcher`      | bool   | false     | Enable file system watching          |
| `logging.level`      | string | `info`    | Log level (debug, info, warn, error) |
| `logging.encoding`   | string | `console` | Log encoding (console, json)         |
| `logging.outputPath` | string | `""`      | Log file path (empty for stdout)     |

## 🔌 Plugin System

The application supports a plugin system using gRPC for extensibility. Plugins can be used to:

- Customize domain validation
- Add custom business logic
- Integrate with external services
- Extend API functionality

### Plugin Configuration

```yaml
plugins:
  my-plugin:
    enabled: true
    registry:
      type: local
      config:
        path: /path/to/plugin/binary
    config:
      # Plugin-specific configuration
      apiKey: "your-api-key"
      endpoint: "https://api.example.com"
```

### Creating a Plugin

See the example plugin in `examples/plugins/simple/` for a complete implementation.

## 📚 API Documentation

### Swagger Documentation

The API includes auto-generated Swagger documentation available at:

- **Swagger UI**: `http://localhost:3000/docs/`
- **OpenAPI JSON**: `http://localhost:3000/docs/doc.json`
- **OpenAPI YAML**: `http://localhost:3000/docs/swagger.yaml`

### API Endpoints

#### Health Check

- `GET /health` - Health check endpoint

#### Domain Management

- `GET /api/v1/domains` - List all domains
- `GET /api/v1/domains/{domain}` - Get specific domain
- `POST /api/v1/domains` - Create new domain
- `PUT /api/v1/domains/{domain}` - Update domain
- `DELETE /api/v1/domains/{domain}` - Delete domain

### Authentication

When authentication is enabled, include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Example API Usage

```bash
# List all domains
curl -H "Authorization: Bearer <token>" http://localhost:3000/api/v1/domains

# Create a new domain
curl -X POST -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>" \
     -d '{"domain": "example.com", "aliases": ["www.example.com"]}' \
     http://localhost:3000/api/v1/domains

# Get specific domain
curl -H "Authorization: Bearer <token>" \
     http://localhost:3000/api/v1/domains/example.com
```

## 🛠️ Development

### Building

```bash
# Build binary
make build

# Build with example plugin
make build-example-plugin

# Generate code and documentation
make generate
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run linting
make lint
```

### Code Generation

The project uses several code generation tools:

```bash
# Generate Swagger documentation
make swag

# Generate protobuf code
go generate ./...
```

### Required Tools

Install the required development tools:

```bash
# Check required tools
make check-tools

# Install missing tools (macOS)
brew install golangci-lint goreleaser protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

## 📋 Makefile Commands

The project includes a comprehensive Makefile with the following commands:

### Build Commands

- `make build` - Build the binary
- `make generate` - Generate code and documentation
- `make swag` - Update Swagger documentation

### Test Commands

- `make test` - Run all tests
- `make test-coverage` - Show coverage report
- `make lint` - Run linter

### Docker Commands

- `make docker-build-local` - Build Docker image with snapshot artifacts
- `make docker-build-release` - Build Docker image for latest release
- `make docker-run` - Run Docker container
- `make docker-stop` - Stop Docker container
- `make docker-logs` - View Docker logs

### Utility Commands

- `make clean` - Clean build artifacts
- `make help` - Show help message
- `make check-tools` - Check required tools

## 🔧 Troubleshooting

### Common Issues

1. **Plugin not loading**: Check plugin path and permissions
2. **Authentication errors**: Verify Azure AD configuration
3. **File watcher issues**: Ensure proper file permissions
4. **Port conflicts**: Change port in configuration

### Logs

Enable debug logging for troubleshooting:

```yaml
logging:
  level: debug
  encoding: console
```

### Health Check

The application includes a health check endpoint at `/health` that returns:

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📞 Support

For support and questions:

- Create an issue on GitHub
- Review the [examples](examples/)