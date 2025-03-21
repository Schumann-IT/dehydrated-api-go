# Plugin System Documentation

This document explains how to build and use plugins for the dehydrated-api-go system.

## Overview

The dehydrated-api-go system uses a gRPC-based plugin architecture that allows you to extend its functionality by implementing custom plugins. Plugins can be written in any language that supports gRPC, but this documentation focuses on Go plugins.

## Plugin Interface

The plugin interface is defined in `proto/plugin/plugin.proto` and consists of three main RPC methods:

1. `Initialize`: Called when the plugin is loaded, providing configuration
2. `EnrichDomainEntry`: Called to enrich domain entries with additional information
3. `Close`: Called when the plugin is being shut down

## Building a Plugin

### 1. Project Structure

Create a new directory for your plugin:

```bash
mkdir -p internal/dehydrated/plugin/yourplugin/grpc
cd internal/dehydrated/plugin/yourplugin/grpc
```

### 2. Server Implementation

Create a new file `server/server.go`:

```go
package server

import (
    "context"
    pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

type Server struct {
    pb.UnimplementedPluginServer
    // Add your plugin-specific fields here
}

func NewServer() *Server {
    return &Server{}
}

func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
    // Initialize your plugin with the provided configuration
    return &pb.InitializeResponse{
        Success: true,
    }, nil
}

func (s *Server) EnrichDomainEntry(ctx context.Context, req *pb.EnrichDomainEntryRequest) (*pb.EnrichDomainEntryResponse, error) {
    // Enrich the domain entry with your plugin's information
    // Add your information to req.Entry.Metadata
    return &pb.EnrichDomainEntryResponse{
        Entry:   req.Entry,
        Success: true,
    }, nil
}

func (s *Server) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
    // Clean up any resources
    return &pb.CloseResponse{
        Success: true,
    }, nil
}

func (s *Server) Serve() error {
    // Start the gRPC server
    // Implementation similar to the certs plugin
}
```

### 3. Main Function

Create a new file `main.go`:

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/yourplugin/grpc/server"
)

func main() {
    var port int
    flag.IntVar(&port, "port", 0, "Port to listen on (0 for random)")
    flag.Parse()

    srv := server.NewServer()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        srv.Close(nil, nil)
        os.Exit(0)
    }()

    if err := srv.Serve(); err != nil {
        fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
        os.Exit(1)
    }
}
```

### 4. Building the Plugin

Build your plugin as a standalone binary:

```bash
go build -o yourplugin internal/dehydrated/plugin/yourplugin/grpc/main.go
```

## Using a Plugin

### 1. Configuration

Add your plugin configuration to the dehydrated-api-go configuration:

```yaml
plugins:
  yourplugin:
    enabled: true
    path: /path/to/yourplugin
    config:
      # Your plugin-specific configuration
```

### 2. Running the Plugin

The plugin system will automatically start your plugin when the dehydrated-api-go service starts. The plugin will:

1. Receive initialization parameters
2. Start listening on a random port
3. Register itself with the main service
4. Begin processing domain entries

### 3. Plugin Lifecycle

- **Initialization**: The plugin receives configuration and sets up its resources
- **Operation**: The plugin processes domain entries and enriches them with metadata
- **Shutdown**: The plugin receives a close request and cleans up its resources

## Example: Certs Plugin

The `certs` plugin is a good example of a plugin implementation. It:

1. Reads certificate information from the dehydrated certificates directory
2. Validates certificates
3. Adds certificate metadata to domain entries

You can find its implementation in `internal/dehydrated/plugin/certs/grpc/`.

## Best Practices

1. **Error Handling**: Always handle errors gracefully and provide meaningful error messages
2. **Resource Management**: Clean up resources in the Close method
3. **Configuration**: Use the configuration map for plugin-specific settings
4. **Metadata**: Use clear, consistent keys for metadata
5. **Logging**: Implement appropriate logging for debugging and monitoring

## Testing

Create tests for your plugin in a `plugin_test.go` file:

```go
func TestYourPlugin(t *testing.T) {
    tests := []struct {
        name string
        // Add test cases
    }{
        // Define test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Implement test logic
        })
    }
}
```

## Troubleshooting

1. **Plugin Not Starting**: Check the plugin path and permissions
2. **Connection Issues**: Verify the plugin is listening on the correct port
3. **Configuration Problems**: Ensure all required configuration is provided
4. **Resource Leaks**: Monitor resource usage and cleanup in the Close method 