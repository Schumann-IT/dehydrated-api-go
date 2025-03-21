package dehydrated

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the API configuration
type Config struct {
	// Server configuration
	Port int `yaml:"port"`

	// Plugin configuration
	Plugins map[string]PluginConfig `yaml:"plugins"`

	// Dehydrated configuration
	DehydratedBaseDir string `yaml:"dehydratedBaseDir"`
}

// PluginConfig holds the configuration for a specific plugin
type PluginConfig struct {
	Enabled bool                   `yaml:"enabled"`
	Path    string                 `yaml:"path"`
	Config  map[string]interface{} `yaml:"config"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Port:              8080,
		Plugins:           make(map[string]PluginConfig),
		DehydratedBaseDir: "",
	}
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Validate port
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}

	// Validate dehydrated config
	if c.DehydratedBaseDir == "" {
		return fmt.Errorf("dehydrated base directory is required")
	}

	// Validate plugin configurations
	for name, plugin := range c.Plugins {
		if !plugin.Enabled {
			continue
		}

		if plugin.Path == "" {
			return fmt.Errorf("plugin path is required for enabled plugin: %s", name)
		}

		// Check if plugin path exists and is executable
		if _, err := os.Stat(plugin.Path); err != nil {
			return fmt.Errorf("plugin path does not exist: %s", plugin.Path)
		}

		// Check if plugin path is absolute
		if !filepath.IsAbs(plugin.Path) {
			return fmt.Errorf("plugin path must be absolute: %s", plugin.Path)
		}
	}

	return nil
}
