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
  - Built-in OpenSSL plugin for certificate analysis and metadata extraction
  - Extensible plugin system for dehydrated hooks integration
  - Example use cases:
    - Check certificate deployment to load balancers (e.g., NetScaler)
    - Check Firewall certificate updates
    - Custom deployment scenarios
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

## Prerequisites

- Go 1.23 or later (for development)
- Docker (optional, for containerized deployment)

## Getting Started

### Quick Start with Docker

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/dehydrated/basedir:/data/dehydrated \
  schumann-it/dehydrated-api-go
```

### Quick Start with binary

1. Download the latest release from [GitHub Releases](https://github.com/schumann-it/dehydrated-api-go/releases)
2. Create a basic configuration file `config.yaml`:
   ```yaml
    port: 3000
    dehydratedBaseDir: /path/to/dehydrated/basedir
    enableWatcher: false
   ```
3. Run the binary:
   ```bash
   ./dehydrated-api-go -config config.yaml
   ```

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

### Plugin System

The API supports two types of plugins:

1. **Built-in Plugins**: Integrated directly into the API
   - OpenSSL Plugin: Parses and analyzes certificate data
   - More built-in plugins can be added in the future

2. **External Plugins**: Implemented as separate services
   - Communicate via gRPC
   - Can implement dehydrated hooks for:
     - DNS challenge automation
     - Certificate deployment
     - Custom deployment scenarios

#### Creating an External Plugin

1. Define your plugin interface using Protocol Buffers:
   ```protobuf
   service DehydratedPlugin {
     rpc DeployCertificate(DeployRequest) returns (DeployResponse);
     rpc CleanupCertificate(CleanupRequest) returns (CleanupResponse);
   }
   ```

2. Implement the plugin service in your preferred language
3. Configure the plugin in the API's configuration:
   ```yaml
   plugins:
     - name: "my-plugin"
       address: "localhost:50051"
       type: "deploy"
   ```

#### Example Plugin Use Cases

- **Load Balancer Integration**: Deploy certificates to NetScaler or other load balancers
- **Firewall Management**: Update firewall certificates automatically
- **DNS Challenge**: Implement DNS challenge automation for various providers
- **Custom Deployment**: Create custom deployment workflows for your infrastructure

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 