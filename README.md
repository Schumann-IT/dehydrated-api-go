# Dehydrated API Go

A REST API service for managing domains in the Dehydrated ACME client. This service provides a programmatic interface to manage SSL/TLS certificates through Dehydrated.

## Features

- RESTful API for domain management
- Plugin system for extensibility
- File-based domain configuration
- Real-time file watching for configuration changes
- Structured logging with Zap
- Graceful shutdown handling
- YAML-based configuration

## Installation

1. Clone the repository:
```bash
git clone https://github.com/schumann-it/dehydrated-api-go.git
cd dehydrated-api-go
```

2. Build the project:
```bash
go build -o dehydrated-api ./cmd/api
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
  example-plugin:
    enabled: true
    path: "/path/to/plugin"
    config:
      api_key: "your-api-key"
```

4. Run the service:
```bash
./dehydrated-api -config config.yaml
```

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
│   ├── handler/       # HTTP request handlers
│   ├── logger/        # Logging configuration
│   ├── model/         # Data models
│   └── service/       # Business logic
├── pkg/
│   └── dehydrated/    # Dehydrated client integration
└── plugin/
    └── registry/      # Plugin management
```

### Building

```bash
# Build the main application
go build -o dehydrated-api ./cmd/api

# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/api
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