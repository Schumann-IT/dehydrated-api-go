# Plugin Development Guide

## Table of Contents

1. [Overview](#overview)
2. [Plugin Types](#plugin-types)
    - [Built-in Plugins](#built-in-plugins)
    - [External Plugins](#external-plugins)
3. [Plugin Interface](#plugin-interface)
4. [Plugin Lifecycle](#plugin-lifecycle)
5. [Creating a New Plugin](#creating-a-new-plugin)
    - [Basic Structure](#1-basic-structure)
    - [Configuration](#2-configuration)
    - [Error Handling](#3-error-handling)
    - [Context Usage](#4-context-usage)
6. [Best Practices](#best-practices)
7. [Example Plugin Implementation](#example-plugin-implementation)
8. [Testing Your Plugin](#testing-your-plugin)
9. [Deployment](#deployment)
10. [Troubleshooting](#troubleshooting)

This guide provides detailed information about developing plugins for the Dehydrated API Go service.

## Overview

Plugins in Dehydrated API Go are used to extend the functionality of the service by providing additional metadata for
domains. Each plugin can implement its own logic to gather and return domain-specific information.

## Plugin Types

The service supports two types of plugins: built-in plugins and external plugins. Each type has its own advantages and
use cases.

### Built-in Plugins

Built-in plugins are implemented directly in the service's codebase and are compiled as part of the main application.
They are:

- Located in `internal/plugin/builtin/`
- Currently includes:
    - `timestamp`: Adds timestamp information to domain metadata
    - `openssl`: Provides OpenSSL-related metadata
- No separate compilation or deployment required
- Loaded automatically when specified in config without a path
- Faster execution as they run in-process
- Easier to maintain and debug
- Limited to Go language implementation

Example configuration for a built-in plugin:

```yaml
plugins:
  timestamp:
    enabled: true
    config:
      time_format: "RFC3339"
```

### External Plugins

External plugins are separate executables that communicate with the service via gRPC protocol. They:

- Can be implemented in any language that supports gRPC
- Must be compiled separately and deployed
- Require a path in the configuration
- Run as separate processes
- More flexible for different use cases
- Can be updated independently of the main service
- Better isolation and fault tolerance

Example configuration for an external plugin:

```yaml
plugins:
  external-plugin:
    enabled: true
    path: "/path/to/external-plugin"
    config:
      api_key: "your-api-key"
      timeout: 30
```

### Choosing Between Plugin Types

Choose a built-in plugin when:

- The functionality is closely tied to the service
- Performance is critical
- You want to maintain the plugin in the main codebase
- The implementation is relatively simple

Choose an external plugin when:

- You need to use a different programming language
- The plugin needs to be updated independently
- The plugin has complex dependencies
- You want better isolation and fault tolerance

## Plugin Interface

All plugins must implement the following interface:

```go
type Plugin interface {
    // Initialize is called when the plugin is loaded.
    // It sets up the plugin with the provided configuration.
    // The context can be used for cancellation and timeout control.
    // Returns an error if initialization fails.
    Initialize(ctx context.Context, config map[string]any) error

    // GetMetadata returns metadata for a domain entry.
    // This method is called to retrieve plugin-specific information about a domain.
    // The dehydratedConfig parameter provides access to the dehydrated configuration
    // for the specific domain being processed.
    // The context can be used for cancellation and timeout control.
    // Returns a map of metadata key-value pairs and an error if the operation fails.
    GetMetadata(ctx context.Context, entry model.DomainEntry, dehydratedConfig *dehydrated.Config) (map[string]any, error)

    // Close is called when the plugin is being unloaded.
    // It performs any necessary cleanup operations.
    // The context can be used for cancellation and timeout control.
    // Returns an error if cleanup fails.
    Close(ctx context.Context) error
}
```

## Plugin Lifecycle

1. **Initialization**: When the plugin is loaded, the `Initialize` method is called with the plugin's configuration. This is where you should set up any resources needed by the plugin.

2. **Metadata Retrieval**: The `GetMetadata` method is called whenever metadata is needed for a domain. This method receives:
   - The domain entry being processed
   - The dehydrated configuration specific to the domain
   - A context for cancellation and timeout control

3. **Cleanup**: When the plugin is being unloaded, the `Close` method is called to perform any necessary cleanup.

## Creating a New Plugin

### 1. Basic Structure

Here's a basic plugin structure:

```go
package main

import (
    "context"
    "time"
)

type MyPlugin struct {
    config map[string]interface{}
}

func NewMyPlugin(config map[string]interface{}) (*MyPlugin, error) {
    // Validate and process configuration
    if err := validateConfig(config); err != nil {
        return nil, err
    }
    
    return &MyPlugin{
        config: config,
    }, nil
}

func (p *MyPlugin) GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error) {
    // Implement your metadata gathering logic here
    return map[string]interface{}{
        "status": "active",
        "last_check": time.Now().Unix(),
    }, nil
}

func (p *MyPlugin) Close(ctx context.Context) error {
    // Clean up any resources
    return nil
}

func validateConfig(config map[string]interface{}) error {
    // Implement configuration validation
    return nil
}
```

### 2. Configuration

Plugins receive their configuration through the `config.yaml` file:

```yaml
plugins:
  my-plugin:
    enabled: true
    path: "/path/to/my-plugin"
    config:
      api_key: "your-api-key"
      timeout: 30
      options:
        check_ssl: true
        check_dns: true
```

### 3. Error Handling

Proper error handling is crucial for plugin development:

```go
func (p *MyPlugin) GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error) {
    // Check context cancellation
    if err := ctx.Err(); err != nil {
        return nil, fmt.Errorf("context error: %w", err)
    }

    // Implement your logic with proper error handling
    metadata, err := gatherMetadata(domain)
    if err != nil {
        return nil, fmt.Errorf("failed to gather metadata: %w", err)
    }

    return metadata, nil
}
```

### 4. Context Usage

Always respect the context for cancellation and timeouts:

```go
func (p *MyPlugin) GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error) {
    // Create a timeout context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Use the context in your operations
    result := make(chan map[string]interface{})
    errChan := make(chan error)

    go func() {
        metadata, err := gatherMetadata(domain)
        if err != nil {
            errChan <- err
            return
        }
        result <- metadata
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case err := <-errChan:
        return nil, err
    case metadata := <-result:
        return metadata, nil
    }
}
```

## Best Practices

1. **Configuration Validation**
    - Always validate plugin configuration at initialization
    - Provide clear error messages for invalid configurations
    - Use type assertions safely when accessing configuration values

2. **Resource Management**
    - Properly clean up resources in the `Close` method
    - Use connection pools for external services
    - Implement timeouts for external calls
    - Handle context cancellation

3. **Error Handling**
    - Use wrapped errors with `fmt.Errorf` and `%w`
    - Provide meaningful error messages
    - Handle context cancellation
    - Log errors appropriately

4. **Performance**
    - Cache results when appropriate
    - Use goroutines for concurrent operations
    - Implement rate limiting for external API calls
    - Use timeouts to prevent hanging operations

5. **Testing**
    - Write unit tests for your plugin
    - Mock external dependencies
    - Test error scenarios
    - Test configuration validation
    - Test resource cleanup

## Example Plugin Implementation

Here's a complete example of a DNS check plugin:

```go
package main

import (
    "context"
    "fmt"
    "net"
    "time"
)

type DNSPlugin struct {
    config struct {
        Timeout time.Duration
        Nameservers []string
    }
}

func NewDNSPlugin(config map[string]interface{}) (*DNSPlugin, error) {
    p := &DNSPlugin{}
    
    // Parse configuration
    if timeout, ok := config["timeout"].(int); ok {
        p.config.Timeout = time.Duration(timeout) * time.Second
    } else {
        p.config.Timeout = 5 * time.Second
    }

    if nameservers, ok := config["nameservers"].([]interface{}); ok {
        p.config.Nameservers = make([]string, len(nameservers))
        for i, ns := range nameservers {
            if str, ok := ns.(string); ok {
                p.config.Nameservers[i] = str
            }
        }
    }

    return p, nil
}

func (p *DNSPlugin) GetMetadata(ctx context.Context, domain string) (map[string]interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
    defer cancel()

    result := make(chan map[string]interface{})
    errChan := make(chan error)

    go func() {
        metadata := make(map[string]interface{})

        // Check A record
        ips, err := net.LookupIP(domain)
        if err != nil {
            errChan <- fmt.Errorf("failed to lookup A record: %w", err)
            return
        }
        metadata["a_records"] = ips

        // Check MX record
        mx, err := net.LookupMX(domain)
        if err != nil {
            errChan <- fmt.Errorf("failed to lookup MX record: %w", err)
            return
        }
        metadata["mx_records"] = mx

        // Check TXT record
        txt, err := net.LookupTXT(domain)
        if err != nil {
            errChan <- fmt.Errorf("failed to lookup TXT record: %w", err)
            return
        }
        metadata["txt_records"] = txt

        result <- metadata
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case err := <-errChan:
        return nil, err
    case metadata := <-result:
        return metadata, nil
    }
}

func (p *DNSPlugin) Close(ctx context.Context) error {
    // No cleanup needed
    return nil
}
```

## Testing Your Plugin

Create a test file for your plugin:

```go
package main

import (
    "context"
    "testing"
    "time"
)

func TestDNSPlugin(t *testing.T) {
    config := map[string]interface{}{
        "timeout": 5,
        "nameservers": []string{"8.8.8.8", "8.8.4.4"},
    }

    plugin, err := NewDNSPlugin(config)
    if err != nil {
        t.Fatalf("Failed to create plugin: %v", err)
    }
    defer plugin.Close(context.Background())

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    metadata, err := plugin.GetMetadata(ctx, "example.com")
    if err != nil {
        t.Fatalf("Failed to get metadata: %v", err)
    }

    // Assert expected metadata
    if _, ok := metadata["a_records"]; !ok {
        t.Error("Expected A records in metadata")
    }
    if _, ok := metadata["mx_records"]; !ok {
        t.Error("Expected MX records in metadata")
    }
    if _, ok := metadata["txt_records"]; !ok {
        t.Error("Expected TXT records in metadata")
    }
}
```

## Deployment

1. Build your plugin:

```bash
go build -o my-plugin ./plugin/main.go
```

2. Update the configuration:

```yaml
plugins:
  my-plugin:
    enabled: true
    path: "/path/to/my-plugin"
    config:
      timeout: 5
      nameservers: ["8.8.8.8", "8.8.4.4"]
```

3. Restart the Dehydrated API service to load the new plugin.

## Troubleshooting

1. **Plugin Not Loading**
    - Check if the plugin path is absolute
    - Verify the plugin is executable
    - Check the service logs for errors

2. **Configuration Issues**
    - Validate the plugin configuration format
    - Check for required configuration values
    - Verify configuration value types

3. **Performance Issues**
    - Check for timeouts in external calls
    - Monitor resource usage
    - Implement caching if needed

4. **Error Handling**
    - Check error messages in logs
    - Verify error propagation
    - Test error scenarios 