# Dehydrated API Go

A RESTful API service that wraps the [dehydrated](https://github.com/dehydrated-io/dehydrated) ACME client to enable remote management of SSL/TLS certificates. This project solves the operational challenges of managing Let's Encrypt certificates at scale by providing a modern API interface to dehydrated's domain configuration.

## Problem Statement

While dehydrated is an excellent tool for managing Let's Encrypt certificates, it becomes challenging to operate at scale:

- Operations teams need direct server access to modify the `domains.txt` configuration
- Manual configuration changes are error-prone and lack audit trails
- No programmatic way to manage certificates across multiple environments
- Difficult to integrate with modern infrastructure-as-code tools

## Solution

Dehydrated API Go transforms dehydrated into a modern, API-driven service that:

- Exposes a REST API for managing domain configurations
- Enables remote certificate management without direct server access
- Integrates seamlessly with infrastructure-as-code tools like Terraform
- Provides a web interface (React Admin) for certificate management
- Maintains the simplicity and reliability of dehydrated while adding enterprise-grade management capabilities

## Key Features

- **Domain Management**: Create, read, update, and delete domain entries
- **Plugin Support**: 
  - Built-in OpenSSL plugin for certificate analysis
  - Custom plugin support via gRPC
- **Security**:
  - JWT-based authentication
  - Azure AD integration
  - Role-based access control
- **Monitoring**:
  - Health check endpoint
  - Structured logging
  - File watching for configuration changes
- **Containerization**:
  - Docker support
  - Environment-based configuration
  - Health checks

## Getting Started

### Prerequisites

- Docker
- Go 1.23 or later (for development)

### Quick Start with Docker

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/certs:/data/dehydrated \
  schumann-it/dehydrated-api-go
```

### Configuration

The API can be configured using environment variables or a YAML configuration file. Key configuration options include:

- `PORT`: API server port (default: 3000)
- `BASE_DIR`: Base directory for dehydrated data
- `ENABLE_WATCHER`: Enable file watching for automatic updates
- `ENABLE_OPENSSL_PLUGIN`: Enable the built-in OpenSSL plugin
- `EXTERNAL_PLUGINS`: Configure external plugins
- `LOGGING`: Configure logging behavior

## API Documentation

The API documentation is available via Swagger UI at `/swagger` when the server is running.

### Main Endpoints

- `GET /api/v1/domains`: List all configured domains
- `GET /api/v1/domains/{domain}`: Get details for a specific domain
- `POST /api/v1/domains`: Create a new domain entry
- `PUT /api/v1/domains/{domain}`: Update a domain entry
- `DELETE /api/v1/domains/{domain}`: Delete a domain entry

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/schumann-it/dehydrated-api-go.git
cd dehydrated-api-go

# Build the binary
make build

# Run tests
make test
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 