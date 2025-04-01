# Configuration Guide

This guide explains how to configure the Dehydrated API Go service.

## Configuration File

The service uses a YAML configuration file (default: `config.yaml`) to manage its settings. Here's a complete example with all available options:

```yaml
# Server configuration
port: 3000

# Dehydrated configuration
dehydrated_base_dir: "/path/to/dehydrated"
enable_watcher: true

# Logging configuration
logging:
  level: "info"      # debug, info, warn, error
  encoding: "console" # json or console
  output_path: ""    # path to log file, empty for stdout

# Plugin configuration
plugins:
  dns-plugin:
    enabled: true
    path: "/path/to/dns-plugin"
    config:
      timeout: 5
      nameservers: ["8.8.8.8", "8.8.4.4"]
  
  ssl-plugin:
    enabled: true
    path: "/path/to/ssl-plugin"
    config:
      check_interval: 3600
      timeout: 10
      verify_chain: true
```

## Configuration Options

### Server Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `port` | integer | 3000 | The port number the server will listen on |

### Dehydrated Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dehydrated_base_dir` | string | "." | The base directory where Dehydrated is installed |
| `enable_watcher` | boolean | false | Whether to enable file watching for domain changes |

### Logging Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `logging.level` | string | "error" | Logging level (debug, info, warn, error) |
| `logging.encoding` | string | "console" | Log format (json, console) |
| `logging.output_path` | string | "" | Path to log file, empty for stdout |

### Plugin Configuration

The `plugins` section allows you to configure multiple plugins. Each plugin has the following options:

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `enabled` | boolean | Yes | Whether the plugin is enabled |
| `path` | string | Yes | Absolute path to the plugin binary |
| `config` | object | No | Plugin-specific configuration |

## Environment Variables

You can configure the logging level using the `LOG_LEVEL` environment variable. This takes precedence over the configuration file.

| Environment Variable | Description |
|---------------------|-------------|
| `LOG_LEVEL` | Logging level (debug, info, warn, error) |

## Example Configurations

### Basic Configuration

```yaml
port: 3000
dehydrated_base_dir: "/opt/dehydrated"
```

### Production Configuration

```yaml
port: 3000
dehydrated_base_dir: "/opt/dehydrated"
enable_watcher: true

logging:
  level: "info"
  encoding: "json"
  output_path: "/var/log/dehydrated-api.log"

plugins:
  dns-plugin:
    enabled: true
    path: "/opt/plugins/dns-plugin"
    config:
      timeout: 5
      nameservers: ["8.8.8.8", "8.8.4.4"]
```

### Development Configuration

```yaml
port: 3000
dehydrated_base_dir: "./dehydrated"
enable_watcher: true

logging:
  level: "debug"
  encoding: "console"

plugins:
  test-plugin:
    enabled: true
    path: "./plugins/test-plugin"
    config:
      debug: true
```

## Configuration File Location

The service looks for the configuration file in the following locations (in order):

1. File specified by the `-config` command-line flag
2. `config.yaml` in the current directory
3. `config.yaml` in the user's home directory
4. `/etc/dehydrated-api/config.yaml`

## Command-Line Flags

The following command-line flags are available:

| Flag | Description | Default |
|------|-------------|---------|
| `-config` | Path to configuration file | "config.yaml" |

## Configuration Validation

The service validates the configuration on startup and will fail if:

1. Port number is invalid (must be between 1 and 65535)
2. Dehydrated base directory does not exist
3. Plugin path is not absolute
4. Plugin path does not exist
5. Plugin path is not executable

## Graceful Shutdown

The service supports graceful shutdown through SIGINT and SIGTERM signals. When a shutdown signal is received:

1. The server stops accepting new connections
2. Existing connections are allowed to complete
3. Resources are cleaned up
4. The process exits

To gracefully shut down the service:

```bash
kill -SIGTERM $(pgrep dehydrated-api)
``` 