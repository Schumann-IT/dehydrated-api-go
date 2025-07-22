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
  # Enable JWT signature validation (recommended for production)
  enableSignatureValidation: true
  # Key cache TTL (e.g., "24h", "1h", "30m")
  keyCacheTTL: "24h"
```

#### JWT Signature Validation

The authentication system now supports **JWT signature validation** for enhanced security:

- **Enabled by default**: Signature validation is enabled by default for production use
- **Azure AD integration**: Automatically fetches and caches public keys from Azure AD
- **Configurable caching**: Keys are cached for 24 hours by default (configurable)
- **Fallback mode**: Can be disabled to use claim-based validation only

**Security Modes:**

1. **Full Validation (Recommended)**: `enableSignatureValidation: true`
   - Validates JWT signature using Azure AD public keys
   - Validates all claims (expiration, audience, issuer)
   - Provides maximum security

2. **Claim-Only Validation**: `enableSignatureValidation: false`
   - Validates claims only (expiration, audience, issuer)
   - Does not validate cryptographic signature
   - Suitable for development/testing environments

### Configuration Options

| Option               | Type   | Default   | Description                          |
|----------------------|--------|-----------|--------------------------------------|
| `port`               | int    | 3000      | HTTP server port                     |
| `dehydratedBaseDir`  | string | `./data`  | Base directory for dehydrated data   |
| `enableWatcher`      | bool   | false     | Enable file system watching          |
| `logging.level`      | string | `info`    | Log level (debug, info, warn, error) |
| `logging.encoding`   | string | `console` | Log encoding (console, json)         |
| `logging.outputPath` | string | `""`      | Log file path (empty for stdout)     |
| `auth.enableSignatureValidation` | bool | true | Enable JWT signature validation |
| `auth.keyCacheTTL`   | string | `24h`     | Key cache time-to-live |

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
      # Optional: Override log level for this plugin
      # logLevel: "debug"
```

#### Plugin Log Level Configuration

Plugins automatically inherit the main application's log level unless explicitly configured. You can override the log level for individual plugins by adding a `logLevel` field to the plugin's configuration:

```yaml
plugins:
  my-plugin:
    enabled: true
    registry:
      type: local
      config:
        path: /path/to/plugin/binary
    config:
      # Override log level for this specific plugin
      logLevel: "debug"
      # Other plugin-specific configuration
      apiKey: "your-api-key"
```

**Available log levels**: `debug`, `info`, `warn`, `error`

If no `logLevel` is specified in the plugin configuration, the plugin will use the same log level as the main application (configured in the `logging.level` section).

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

- `GET /api/v1/domains` - List all domains (with pagination)
- `GET /api/v1/domains/{domain}` - Get specific domain
- `POST /api/v1/domains` - Create new domain
- `PUT /api/v1/domains/{domain}` - Update domain
- `DELETE /api/v1/domains/{domain}` - Delete domain

### Pagination

The `ListDomains` endpoint supports pagination to efficiently handle large datasets. This implementation follows the **Hybrid Approach** with query parameters and rich response metadata.

#### Query Parameters

| Parameter | Type | Required | Default | Min | Max | Description |
|-----------|------|----------|---------|-----|-----|-------------|
| `page` | integer | No | 1 | 1 | - | Page number (1-based) |
| `per_page` | integer | No | 100 | 1 | 1000 | Number of items per page |
| `sort` | string | No | "" | - | - | Sort order for domain field ("asc" or "desc", optional - defaults to alphabetical order) |
| `search` | string | No | "" | - | - | Search term to filter domains by domain field (case-insensitive contains) |

#### Response Format

The response uses the `PaginatedDomainsResponse` structure:

```json
{
  "success": true,
  "data": [
    {
      "domain": "example.com",
      "alternative_names": ["www.example.com"],
      "alias": "",
      "enabled": true,
      "comment": "Production domain",
      "metadata": {}
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 100,
    "total": 150,
    "total_pages": 2,
    "has_next": true,
    "has_prev": false,
    "next_url": "/api/v1/domains?page=2&per_page=100",
    "prev_url": ""
  }
}
```

#### Pagination Metadata

| Field | Type | Description |
|-------|------|-------------|
| `current_page` | integer | Current page number (1-based) |
| `per_page` | integer | Number of items per page |
| `total` | integer | Total number of items across all pages |
| `total_pages` | integer | Total number of pages |
| `has_next` | boolean | Whether there is a next page |
| `has_prev` | boolean | Whether there is a previous page |
| `next_url` | string | URL for the next page (if available) |
| `prev_url` | string | URL for the previous page (if available) |

#### Examples

**Basic Usage:**
```bash
# Get first page with default settings (100 items per page)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains"
```

**Custom Page Size:**
```bash
# Get first page with 10 items per page
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?page=1&per_page=10"
```

**Navigation:**
```bash
# Get second page with 10 items per page
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?page=2&per_page=10"
```

**Using Generated URLs:**
```bash
# Get first page
response=$(curl -s -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?page=1&per_page=5")

# Extract next URL
next_url=$(echo "$response" | jq -r '.pagination.next_url')

# Navigate to next page
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000$next_url"
```

**Sorting:**
```bash
# Get domains in default order (alphabetical, same as sort=asc)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains"

# Sort domains in ascending order (alphabetical)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?sort=asc"

# Sort domains in descending order (reverse alphabetical)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?sort=desc"
```

**Searching:**
```bash
# Search for domains containing "example"
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?search=example"

# Case-insensitive search
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?search=EXAMPLE"
```

**Combined Features:**
```bash
# Search and sort with pagination
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?search=example&sort=desc&page=1&per_page=10"
```

#### Error Handling

**Invalid Page Number:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?page=0"
```

**Response:**
```json
{
  "success": false,
  "error": "page parameter must be at least 1"
}
```

**Large Page Size (Auto-capped):**
If `per_page` exceeds 1000, it will be automatically capped to 1000:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?per_page=2000"
```

**Response:**
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "current_page": 1,
    "per_page": 1000,  // Capped from 2000
    "total": 150,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

**Invalid Sort Parameter:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:3000/api/v1/domains?sort=invalid"
```

**Response:**
```json
{
  "success": false,
  "error": "sort parameter must be either 'asc' or 'desc'"
}
```

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