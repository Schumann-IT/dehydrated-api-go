# Configuring External Plugins

This document explains how to configure external plugins for the dehydrated-api-go application using the
`EXTERNAL_PLUGINS` environment variable.

## Format

The `EXTERNAL_PLUGINS` environment variable uses a JSON format to specify plugin configurations:

```json
{
  "plugin_name": {
    "enabled": true|false,  // Optional, defaults to true
    "path": "/path/to/plugin",  // Required
    "config": {  // Optional, defaults to empty map {}
      "key1": "value1",
      "key2": "value2"
    }
  }
}
```

### Required and Optional Fields

- **path** (required): Absolute path to the plugin script within the container
- **enabled** (optional): Boolean flag to enable/disable the plugin. Defaults to `true`
- **config** (optional): Map of key-value pairs for plugin configuration. Defaults to `{}`

## Examples

### 1. Minimal Plugin Configuration

Only specifying the required path:

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/dehydrated:/data/dehydrated \
  -v /path/to/azure-hook.sh:/app/plugins/azure-hook.sh \
  -e EXTERNAL_PLUGINS='{"azure":{"path":"/app/plugins/azure-hook.sh"}}' \
  dehydrated-api-go
```

### 2. Full Plugin Configuration

Including all available options:

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/dehydrated:/data/dehydrated \
  -v /path/to/azure-hook.sh:/app/plugins/azure-hook.sh \
  -e EXTERNAL_PLUGINS='{
    "azure": {
      "enabled": true,
      "path": "/app/plugins/azure-hook.sh",
      "config": {
        "subscription_id": "your-subscription-id",
        "resource_group": "your-resource-group",
        "zone_name": "your-zone-name"
      }
    }
  }' \
  dehydrated-api-go
```

### 3. Multiple Plugins with Configuration

Configuring multiple plugins with different settings:

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/dehydrated:/data/dehydrated \
  -v /path/to/azure-hook.sh:/app/plugins/azure-hook.sh \
  -v /path/to/dns-hook.sh:/app/plugins/dns-hook.sh \
  -e EXTERNAL_PLUGINS='{
    "azure": {
      "enabled": true,
      "path": "/app/plugins/azure-hook.sh",
      "config": {
        "subscription_id": "your-subscription-id",
        "resource_group": "your-resource-group"
      }
    },
    "dns": {
      "path": "/app/plugins/dns-hook.sh",
      "config": {
        "api_token": "your-api-token",
        "zone_id": "your-zone-id"
      }
    }
  }' \
  dehydrated-api-go
```

### 4. Disabling a Plugin with Configuration

Including configuration but keeping the plugin disabled:

```bash
docker run -d \
  -p 3000:3000 \
  -v /path/to/dehydrated:/data/dehydrated \
  -v /path/to/azure-hook.sh:/app/plugins/azure-hook.sh \
  -e EXTERNAL_PLUGINS='{
    "azure": {
      "enabled": false,
      "path": "/app/plugins/azure-hook.sh",
      "config": {
        "subscription_id": "your-subscription-id"
      }
    }
  }' \
  dehydrated-api-go
```

## Important Notes

1. **JSON Format**: The value of `EXTERNAL_PLUGINS` must be valid JSON. Make sure to:
    - Use double quotes for strings
    - Use `true` or `false` (lowercase) for boolean values
    - Escape special characters in paths if necessary

2. **Plugin Paths**:
    - The `path` field is required for each plugin
    - Paths should be absolute within the container
    - Mount plugin scripts using volumes to make them available inside the container
    - Ensure plugin scripts have executable permissions

3. **Plugin Names**:
    - Use lowercase names without spaces
    - Names should be unique
    - Avoid using special characters

4. **Default Values**:
    - `enabled`: Defaults to `true` if not specified
    - `config`: Defaults to empty map `{}` if not specified
    - `path`: Must be specified (no default)

5. **Default Plugins**:
    - The OpenSSL plugin is built-in and configured separately using `ENABLE_OPENSSL_PLUGIN`
    - External plugins configuration doesn't affect built-in plugins

## Generated Configuration

The environment variable will generate entries in the `config.yaml` file like this:

```yaml
plugins:
  openssl:
    enabled: true
  azure:
    enabled: true
    path: "/app/plugins/azure-hook.sh"
    config:
      subscription_id: "your-subscription-id"
      resource_group: "your-resource-group"
  dns:
    enabled: true
    path: "/app/plugins/dns-hook.sh"
    config:
      api_token: "your-api-token"
      zone_id: "your-zone-id"
```

## Troubleshooting

1. **Invalid JSON**: If the JSON format is invalid, the script will log an error and skip plugin configuration.
2. **Missing Path**: If a plugin configuration doesn't include a path, it will be skipped with an error message.
3. **Missing Plugin Files**: Ensure that plugin scripts are mounted correctly and have the right permissions.
4. **Plugin Path**: Make sure the path in the configuration matches the actual mounted location in the container.
5. **Config Values**: All config values are passed as strings to the plugin. Make sure your plugin handles type
   conversion if needed. 