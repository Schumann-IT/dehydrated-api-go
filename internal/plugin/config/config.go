package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// RegistryType represents the type of plugin registry
type PluginSourceType string

const (
	PluginSourceTypeLocal  PluginSourceType = "local"
	PluginSourceTypeGitHub PluginSourceType = "github"
)

// PluginConfig holds configuration for a plugin.
// It defines the basic settings needed to load and configure a plugin.
type PluginConfig struct {
	// Enabled determines whether the plugin should be loaded and used.
	Enabled bool `yaml:"enabled"`

	// Registry configuration for plugin source
	Registry *RegistryConfig `yaml:"registry"`

	// Config contains plugin-specific configuration settings.
	// The structure of this map depends on the specific plugin implementation.
	Config map[string]any `yaml:"config"`
}

// RegistryConfig represents the configuration for a plugin registry
type RegistryConfig struct {
	Type   PluginSourceType `yaml:"type"`
	Config map[string]any   `yaml:"config"`
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
