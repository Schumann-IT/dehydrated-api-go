package config

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// PluginConfig holds configuration for a plugin.
// It defines the basic settings needed to load and configure a plugin.
type PluginConfig struct {
	// Enabled determines whether the plugin should be loaded and used.
	Enabled bool `yaml:"enabled"`

	// Path specifies the location of the plugin executable or library.
	// This is used for local plugins or to override GitHub repository settings.
	// DEPRECATED: Use Registry instead
	Path string `yaml:"path"`

	// GitHub repository information for remote plugins
	// DEPRECATED: Use Registry instead
	GitHub *GitHubConfig `yaml:"github"`

	// Registry configuration for plugin source
	Registry *RegistryConfig `yaml:"registry"`

	// Config contains plugin-specific configuration settings.
	// The structure of this map depends on the specific plugin implementation.
	Config map[string]any `yaml:"config"`
}

// RegistryConfig represents the configuration for a plugin registry
type RegistryConfig struct {
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

// GitHubConfig holds configuration for GitHub-based plugins
type GitHubConfig struct {
	// Repository in format "owner/repo" (e.g., "Schumann-IT/dehydrated-api-metadata-plugin-netscaler")
	Repository string `yaml:"repository"`

	// Version tag to use (e.g., "v1.0.0", "latest")
	// If not specified, defaults to "latest"
	Version string `yaml:"version"`

	// Platform to download (e.g., "linux-amd64", "darwin-amd64")
	// If not specified, will be auto-detected
	Platform string `yaml:"platform"`
}

// ToProto converts the config to a proto InitializeRequest
func (c *PluginConfig) ToProto() (map[string]*structpb.Value, error) {
	values := make(map[string]*structpb.Value)
	for k, v := range c.Config {
		val, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", k, err)
		}
		values[k] = val
	}

	return values, nil
}

// IsGitHubPlugin returns true if this plugin should be loaded from GitHub
func (c *PluginConfig) IsGitHubPlugin() bool {
	// Check new registry configuration first
	if c.Registry != nil && c.Registry.Type == "github" {
		return true
	}
	// Fallback to old GitHub configuration
	return c.GitHub != nil && c.GitHub.Repository != ""
}

// GetPluginPath returns the effective plugin path
// If Path is set, it takes precedence over GitHub configuration
func (c *PluginConfig) GetPluginPath() string {
	if c.Path != "" {
		return c.Path
	}
	return ""
}

// GetGitHubInfo returns GitHub configuration if available
func (c *PluginConfig) GetGitHubInfo() *GitHubConfig {
	if c.IsGitHubPlugin() {
		return c.GitHub
	}
	return nil
}

// Validate checks if the plugin configuration is valid
func (c *PluginConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// If Registry is configured, validate it
	if c.Registry != nil {
		if c.Registry.Type == "" {
			return fmt.Errorf("registry type is required")
		}

		// Basic validation based on registry type
		switch c.Registry.Type {
		case "local":
			if c.Registry.Config == nil {
				return fmt.Errorf("registry config is required for local registry")
			}
			if path, ok := c.Registry.Config["path"].(string); !ok || path == "" {
				return fmt.Errorf("path is required for local registry")
			}
		case "github":
			if c.Registry.Config == nil {
				return fmt.Errorf("registry config is required for GitHub registry")
			}
			if repository, ok := c.Registry.Config["repository"].(string); !ok || repository == "" {
				return fmt.Errorf("repository is required for GitHub registry")
			}
			// Validate repository format (should be owner/repo)
			if repository, ok := c.Registry.Config["repository"].(string); ok {
				parts := strings.Split(repository, "/")
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return fmt.Errorf("invalid GitHub repository format: %s (expected owner/repo)", repository)
				}
			}
		default:
			return fmt.Errorf("unsupported registry type: %s", c.Registry.Type)
		}

		return nil
	}

	// If Path is set, it takes precedence
	if c.Path != "" {
		return nil
	}

	// If GitHub is configured, validate it
	if c.GitHub != nil {
		if c.GitHub.Repository == "" {
			return fmt.Errorf("GitHub repository is required when using GitHub configuration")
		}

		// Validate repository format (should be owner/repo)
		parts := strings.Split(c.GitHub.Repository, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid GitHub repository format: %s (expected owner/repo)", c.GitHub.Repository)
		}
	}

	return nil
}
