package config

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// PluginConfig holds configuration for a plugin.
// It defines the basic settings needed to load and configure a plugin.
type PluginConfig struct {
	// Enabled determines whether the plugin should be loaded and used.
	Enabled bool `yaml:"enabled"`

	// Path specifies the location of the plugin executable or library.
	Path string `yaml:"path"`

	// Config contains plugin-specific configuration settings.
	// The structure of this map depends on the specific plugin implementation.
	Config map[string]any `yaml:"config"`
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
