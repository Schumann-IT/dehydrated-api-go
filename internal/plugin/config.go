// Package plugin provides functionality for managing plugins in the dehydrated-api-go application.
// It includes interfaces for plugin implementation, configuration structures, and plugin registry management.
package plugin

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
