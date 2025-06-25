# Docker Setup for dehydrated-api-go

This document explains how to build and run the dehydrated-api-go application using Docker.

## Overview

The Dockerfile is designed to download binary artifacts from GitHub releases instead of building from source code. This approach provides several benefits:

- **Faster builds**: No compilation time required
- **Consistent artifacts**: Uses the same binaries as official releases
- **Smaller images**: No build tools or source code included
- **Security**: Uses pre-built, tested binaries

## Build Scenarios

### 1. Using Latest Release (Default)

Build the image using the latest GitHub release:

```bash
docker build -t dehydrated-api-go:latest .
```

### 2. Using Specific Version

Build the image using a specific version tag:

```bash
docker build --build-arg VERSION=v1.0.0 -t dehydrated-api-go:v1.0.0 .
```

### 3. Using Snapshot/Local Binary

For development or when no releases are available, the Dockerfile will fall back to using a local binary:

```bash
# First build the local binary
make build

# Then build the Docker image
docker build --build-arg VERSION=snapshot -t dehydrated-api-go:snapshot .
```

## Running the Container

### Basic Usage

```bash
docker run -d \
  --name dehydrated-api \
  -p 3000:3000 \
  -v /path/to/config:/app/config \
  -v /path/to/data:/app/data \
  dehydrated-api-go:latest
```

### With Custom Configuration

```bash
docker run -d \
  --name dehydrated-api \
  -p 3000:3000 \
  -v /path/to/config.yaml:/app/config/config.yaml \
  -v /path/to/data:/app/data \
  dehydrated-api-go:latest \
  --config /app/config/config.yaml
```

### Environment Variables

The container supports the following environment variables:

- `VERSION`: Specify the version to download (default: latest)
- `GITHUB_REPO`: Override the GitHub repository (default: schumann-it/dehydrated-api-go)

## Docker Compose Example

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  dehydrated-api:
    build:
      context: .
      args:
        VERSION: latest
    container_name: dehydrated-api
    ports:
      - "3000:3000"
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

Run with:

```bash
docker-compose up -d
```

## Multi-Architecture Support

The Dockerfile automatically detects the target architecture and downloads the appropriate binary:

- **x86_64**: Downloads `dehydrated-api-go_Linux_x86_64.tar.gz`
- **arm64**: Downloads `dehydrated-api-go_Linux_arm64.tar.gz`

## Security Features

- **Non-root user**: The container runs as user `appuser` (UID 1000)
- **Minimal base image**: Uses Alpine Linux for a smaller attack surface
- **Health checks**: Built-in health monitoring
- **Read-only filesystem**: Only necessary directories are writable

## Troubleshooting

### Build Fails with "Failed to download from GitHub releases"

This happens when:
1. No releases are available on GitHub
2. The specified version doesn't exist
3. Network connectivity issues

**Solution**: Build a local binary first:
```bash
make build
docker build --build-arg VERSION=snapshot -t dehydrated-api-go:snapshot .
```

### Permission Issues

If you encounter permission issues with mounted volumes:

```bash
# Create directories with correct permissions
mkdir -p config data
chown -R 1000:1000 config data

# Run container
docker run -v $(pwd)/config:/app/config -v $(pwd)/data:/app/data dehydrated-api-go:latest
```

### Health Check Fails

The health check expects the application to respond on `/health` endpoint. If your configuration uses a different health endpoint, you can override it:

```bash
docker run --health-cmd="curl -f http://localhost:3000/your-health-endpoint" dehydrated-api-go:latest
```

## Development Workflow

For development, you can use the snapshot build:

1. Make your changes
2. Build the local binary: `make build`
3. Build the Docker image: `docker build --build-arg VERSION=snapshot -t dehydrated-api-go:dev .`
4. Test your changes: `docker run dehydrated-api-go:dev`

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push Docker Image

on:
  push:
    tags:
      - 'v*'

jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: |
          docker build --build-arg VERSION=${{ github.ref_name }} -t dehydrated-api-go:${{ github.ref_name }} .
      
      - name: Push to registry
        run: |
          docker tag dehydrated-api-go:${{ github.ref_name }} your-registry/dehydrated-api-go:${{ github.ref_name }}
          docker push your-registry/dehydrated-api-go:${{ github.ref_name }}
```

## Best Practices

1. **Always specify a version**: Use specific version tags in production
2. **Use multi-stage builds**: The Dockerfile uses multi-stage builds for efficiency
3. **Mount configuration**: Mount configuration files as volumes for easy updates
4. **Health checks**: Use the built-in health checks for monitoring
5. **Resource limits**: Set appropriate resource limits in production
6. **Logging**: Configure proper logging for production deployments

## Performance Considerations

- **Image size**: ~15MB base image + binary size
- **Startup time**: ~1-2 seconds (no compilation)
- **Memory usage**: Minimal, depends on application configuration
- **CPU usage**: No build overhead during container startup 