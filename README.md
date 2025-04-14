# Dehydrated API Go

A REST API service for managing domains in the Dehydrated ACME client. This service provides a programmatic interface to
manage SSL/TLS certificates through Dehydrated.

## Features

- RESTful API for domain management
- Plugin system for extensibility
- File-based domain configuration
- Real-time file watching for configuration changes
- Structured logging with Zap
- Graceful shutdown handling
- YAML-based configuration

## Prerequisites

- Go 1.21 or later
- Make
- Docker (optional, for containerized deployment)
- OpenSSL (for certificate operations)
- dcron (for scheduled renewals)

## Installation

### From Source

1. Clone the repository:

```bash
git clone https://github.com/schumann-it/dehydrated-api-go.git
cd dehydrated-api-go
```

2. Build the project:

```bash
make build
```

3. Create a configuration file (config.yaml):

```yaml
port: 3000
dehydrated_base_dir: "/path/to/dehydrated"
enable_watcher: true

logging:
  level: "info"
  encoding: "console"
  output_path: ""

plugins:
  openssl:
    enabled: true
#  example-external-plugin:
#    enabled: true
#    path: "/path/to/plugin"
#    config:
#      api_key: "your-api-key"
```

4. Run the service:

```bash
make run
```

### Using Docker

1. Build the Docker image:

```bash
make docker-build
```

2. Run the container:

```bash
make docker-run
```

## Configuration

The application can be configured using environment variables:

- `PORT`: API server port (default: 3000)
- `ENABLE_WATCHER`: Enable file system watcher (default: false)
- `ENABLE_OPENSSL_PLUGIN`: Enable OpenSSL plugin (default: true)
- `CRON_SCHEDULE`: Cron schedule for certificate renewal (format: "0 */12 * * *")
- `EXTERNAL_PLUGINS`: JSON configuration for external plugins
- `DEHYDRATED_*`: Any dehydrated configuration setting prefixed with DEHYDRATED_

### External Plugins Configuration

External plugins can be configured using the `EXTERNAL_PLUGINS` environment variable in JSON format:

```json
{
  "plugin_name": {
    "enabled": true,
    "path": "/path/to/plugin",
    "config": {
      "key": "value"
    }
  }
}
```

See [external-plugins.md](docs/external-plugins.md) for detailed documentation.

## API Usage

### List Domains

```bash
curl http://localhost:3000/api/v1/domains
```

### Get Domain

```bash
curl http://localhost:3000/api/v1/domains/example.com
```

### Create Domain

```bash
curl -X POST http://localhost:3000/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "example.com",
    "alternative_names": ["www.example.com"],
    "alias": "example",
    "enabled": true,
    "comment": "Production domain"
  }'
```

### Update Domain

```bash
curl -X PUT http://localhost:3000/api/v1/domains/example.com \
  -H "Content-Type: application/json" \
  -d '{
    "alternative_names": ["www.example.com", "api.example.com"],
    "enabled": true,
    "comment": "Updated production domain"
  }'
```

### Delete Domain

```bash
curl -X DELETE http://localhost:3000/api/v1/domains/example.com
```

## Plugin Development

### Plugin Interface

To create a new plugin, implement the following interface:

```go
type Plugin interface {
    // GetMetadata retrieves metadata for a domain
    GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error)
    
    // Close cleans up plugin resources
    Close(ctx context.Context) error
}
```

### Example Plugin

Here's a simple example of a plugin implementation:

```go
package main

import (
    "context"
    "time"
)

type ExamplePlugin struct {
    config map[string]interface{}
}

func NewExamplePlugin(config map[string]interface{}) (*ExamplePlugin, error) {
    return &ExamplePlugin{
        config: config,
    }, nil
}

func (p *ExamplePlugin) GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error) {
    return map[string]interface{}{
        "status": "active",
        "last_check": time.Now().Unix(),
    }, nil
}

func (p *ExamplePlugin) Close(ctx context.Context) error {
    return nil
}
```

### Plugin Registration

Plugins are automatically loaded based on the configuration in `config.yaml`. Each plugin needs:

1. A unique name in the configuration
2. The `enabled` flag set to `true`
3. The `path` to the plugin binary (must be absolute)
4. Any additional configuration needed by the plugin

Example configuration:

```yaml
plugins:
  example-plugin:
    enabled: true
    path: "/path/to/example-plugin"
    config:
      api_key: "your-api-key"
      timeout: 30
```

## Development

### Project Structure

```
dehydrated-api-go/
├── cmd/
│   └── api/           # Main application entry point
├── internal/
│   ├── api/           # API handlers and routes
│   ├── config/        # Configuration management
│   ├── dehydrated/    # Dehydrated integration
│   └── watcher/       # File system watcher
├── scripts/           # Utility scripts
│   ├── configure-cron.sh
│   ├── generate-config.sh
│   ├── healthcheck.sh
│   ├── renew-certs.sh
│   ├── start-api.sh
│   ├── start-crond.sh
│   ├── test-configure-cron.sh
│   ├── update-api-config.sh
│   └── update-dehydrated-config.sh
├── examples/          # Example configurations
│   └── config/
│       ├── config.yaml
│       └── dehydrated
├── docs/             # Documentation
│   └── external-plugins.md
├── Dockerfile        # Container definition
├── Makefile         # Build and development tasks
└── README.md        # Project documentation
```

### Development Commands

```bash
# Build the application
make build

# Run tests
make test

# Run linters
make lint

# Clean build artifacts
make clean

# Run with race detector
make run-race

# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Stop Docker container
make docker-stop

# View Docker logs
make docker-logs

# Open shell in Docker container
make docker-shell

# Clean Docker artifacts
make docker-clean
```

### Adding New Features

1. Create new models in `internal/model/`
2. Add business logic in `internal/service/`
3. Create handlers in `internal/handler/`
4. Register new routes in the main application

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 