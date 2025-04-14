package plugin

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	Enabled bool           `yaml:"enabled"`
	Path    string         `yaml:"path"`
	Config  map[string]any `yaml:"config"`
}
