# Dehydrated API Go

A Go implementation of an API for managing dehydrated domains and certificates.

## Building

The project uses a Makefile to automate the build process. Here are the available make targets:

```bash
make        # Build both the API server and certs plugin
make build  # Same as above
make test   # Run all tests
make clean  # Remove build artifacts
```

Individual components can be built using:
```bash
make build-api     # Build only the API server
make build-plugin  # Build only the certs plugin
```

## Running

1. Build the project:
```bash
make
```

2. Start the API server:
```bash
./bin/api
```

The API server will automatically start the certs plugin as a separate process.

## API Documentation

The API provides RESTful endpoints for managing domains and their certificates. All endpoints are prefixed with `/api/v1`.

### Endpoints

#### List Domains
```
GET /api/v1/domains
```
Returns a list of all domains.

Response:
```json
{
  "success": true,
  "data": [
    {
      "domain": "example.com",
      "alternative_names": ["www.example.com"],
      "alias": "",
      "enabled": true,
      "comment": "Main website",
      "metadata": {
        "cert_info": {
          "is_valid": true,
          "issuer": "Let's Encrypt",
          "subject": "example.com",
          "not_before": "2024-03-20T00:00:00Z",
          "not_after": "2024-06-18T00:00:00Z"
        }
      }
    }
  ]
}
```

#### Get Domain
```
GET /api/v1/domains/:domain
```
Returns details for a specific domain.

Response:
```json
{
  "success": true,
  "data": {
    "domain": "example.com",
    "alternative_names": ["www.example.com"],
    "alias": "",
    "enabled": true,
    "comment": "Main website",
    "metadata": {
      "cert_info": {
        "is_valid": true,
        "issuer": "Let's Encrypt",
        "subject": "example.com",
        "not_before": "2024-03-20T00:00:00Z",
        "not_after": "2024-06-18T00:00:00Z"
      }
    }
  }
}
```

#### Create Domain
```
POST /api/v1/domains
```
Creates a new domain entry.

Request:
```json
{
  "domain": "example.com",
  "alternative_names": ["www.example.com"],
  "alias": "",
  "enabled": true,
  "comment": "Main website"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "domain": "example.com",
    "alternative_names": ["www.example.com"],
    "alias": "",
    "enabled": true,
    "comment": "Main website"
  }
}
```

#### Update Domain
```
PUT /api/v1/domains/:domain
```
Updates an existing domain entry.

Request:
```json
{
  "alternative_names": ["www.example.com"],
  "alias": "",
  "enabled": true,
  "comment": "Updated comment"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "domain": "example.com",
    "alternative_names": ["www.example.com"],
    "alias": "",
    "enabled": true,
    "comment": "Updated comment"
  }
}
```

#### Delete Domain
```
DELETE /api/v1/domains/:domain
```
Deletes a domain entry.

Response: 204 No Content

### Error Responses

All endpoints return error responses in the following format:
```json
{
  "success": false,
  "error": "Error message"
}
```

Common HTTP status codes:
- 200: Success
- 201: Created
- 204: No Content (successful deletion)
- 400: Bad Request
- 404: Not Found
- 500: Internal Server Error

## Configuration

The application can be configured using environment variables:

- `DEHYDRATED_BASE_DIR` - Base directory for dehydrated (optional)
  - Default: Current working directory
  - This directory should contain the `domains.txt` file and certificate files

## Features

### Domain Management
- CRUD operations for domain entries
- Support for alternative names (SANs)
- Domain aliases
- Enable/disable domains
- Comments for documentation

### Certificate Management
- Automatic certificate status monitoring
- Certificate metadata enrichment via plugin system
- Certificate validity period tracking
- Certificate issuer information

### Plugin System
- Modular plugin architecture using gRPC
- Independent plugin processes for isolation
- Built-in certs plugin for certificate management
- Extensible for additional functionality

### File Management
- Automatic watching of domains.txt for changes
- Safe file operations with proper locking
- Preserves file format and comments

## Development

The project uses gRPC for plugin communication. If you modify the proto files, you'll need to regenerate the Go code:

```bash
protoc --go_out=. --go-grpc_out=. internal/dehydrated/plugin/rpc/plugin.proto
```

### Adding New Plugins

1. Define the plugin interface in a proto file
2. Generate the gRPC code
3. Implement the plugin server
4. Create a new plugin binary
5. Register the plugin in the main application

## Project Structure

- `cmd/api` - Main API server
- `cmd/certs-plugin` - Certificate management plugin
- `internal/dehydrated` - Core functionality
  - `config` - Configuration handling
  - `handler` - HTTP handlers
  - `model` - Data models
  - `plugin` - Plugin system
  - `service` - Business logic

## Installation

```bash
go get github.com/schumann-it/dehydrated-api-go
```