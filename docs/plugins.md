# Plugin System Documentation

This document explains how to build and use plugins for the dehydrated-api-go system.

## Overview

The dehydrated-api-go system uses a gRPC-based plugin architecture that allows you to extend its functionality by implementing custom plugins. Plugins can be written in any language that supports gRPC, but this documentation focuses on Go plugins.

## Plugin Interface

The plugin interface is defined in `proto/plugin/plugin.proto` and consists of three main RPC methods:

1. `Initialize`: Called when the plugin is loaded, providing configuration
2. `EnrichDomainEntry`: Called to enrich domain entries with additional information
3. `Close`: Called when the plugin is being shut down

## Plugin Configuration

The system uses static configuration to manage plugins. Each plugin must be explicitly configured in the dehydrated-api-go configuration file:

```yaml
plugins:
  your-plugin:
    enabled: true
    path: /path/to/your-plugin
    config:
      # Plugin-specific configuration
```

### Plugin Configuration Structure

Each plugin configuration contains the following fields:
```go
type PluginConfig struct {
    Enabled bool                   // Whether the plugin is enabled
    Path    string                 // Path to the plugin binary
    Config  map[string]interface{} // Plugin-specific configuration
}
```

**Pros:**
- Simple to implement and understand
- Explicit control over which plugins are loaded
- Easy to manage plugin versions and dependencies
- No runtime plugin discovery overhead
- Better security control

**Cons:**
- Manual configuration required
- No automatic plugin updates
- Limited flexibility
- Requires service restart for plugin changes

### Best Practices for Plugin Configuration

1. **Security**
   - Use absolute paths for plugin binaries
   - Set appropriate file permissions
   - Validate plugin signatures
   - Implement access controls
   - Sandbox plugin execution

2. **Reliability**
   - Use stable plugin versions
   - Implement proper error handling
   - Monitor plugin health
   - Log plugin operations

3. **Performance**
   - Optimize plugin resource usage
   - Implement caching where appropriate
   - Monitor plugin performance
   - Handle plugin failures gracefully

4. **Monitoring**
   - Track plugin metrics
   - Log plugin events
   - Monitor plugin health
   - Alert on failures

## Building a Plugin

### 1. Project Structure

Create a new directory for your plugin:

```bash
mkdir -p yourplugin
cd yourplugin
```

### 2. Server Implementation

Create a new file `plugin.go`:

```go
package main

import (
    "context"
    pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

// Server implements the gRPC plugin service
type Server struct {
    pb.UnimplementedPluginServer
    // Add your plugin-specific fields here
}

// NewServer creates a new gRPC plugin server
func NewServer() *Server {
    return &Server{}
}

// Initialize implements the Initialize RPC
func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
    // Initialize your plugin with the provided configuration
    return &pb.InitializeResponse{
        Success: true,
    }, nil
}

// EnrichDomainEntry implements the EnrichDomainEntry RPC
func (s *Server) EnrichDomainEntry(ctx context.Context, req *pb.EnrichDomainEntryRequest) (*pb.EnrichDomainEntryResponse, error) {
    // Enrich the domain entry with your plugin's information
    // Add your information to req.Entry.Metadata
    return &pb.EnrichDomainEntryResponse{
        Entry:   req.Entry,
        Success: true,
    }, nil
}

// Close implements the Close RPC
func (s *Server) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
    // Clean up any resources
    return &pb.CloseResponse{
        Success: true,
    }, nil
}
```

### 3. Main Function

Create a new file `main.go`:

```go
package main

import (
    "flag"
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"

    "google.golang.org/grpc"
    pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

func main() {
    var port int
    flag.IntVar(&port, "port", 0, "Port to listen on (0 for random)")
    flag.Parse()

    srv := NewServer()

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        srv.Close(nil, nil)
        os.Exit(0)
    }()

    // Start the gRPC server
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
        os.Exit(1)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterPluginServer(grpcServer, srv)

    if err := grpcServer.Serve(lis); err != nil {
        fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
        os.Exit(1)
    }
}
```

### 4. Building the Plugin

Build your plugin as a standalone binary:

```bash
go build -o yourplugin
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