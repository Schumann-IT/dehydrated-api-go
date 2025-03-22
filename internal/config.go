package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Port int `yaml:"port"`

	// Plugin configuration
	Plugins map[string]PluginConfig `yaml:"plugins"`

	// Dehydrated configuration
	DehydratedBaseDir string `yaml:"dehydratedBaseDir"`
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	Enabled bool              `yaml:"enabled"`
	Path    string            `yaml:"path"`
	Config  map[string]string `yaml:"config"`
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:              8080,
		DehydratedBaseDir: "/etc/dehydrated",
		Plugins:           make(map[string]PluginConfig),
	}
}

// WithBaseDir sets the dehydrated base directory
func (c *Config) WithBaseDir(dir string) *Config {
	c.DehydratedBaseDir = dir
	return c
}

// Load loads configuration from a YAML file
func (c *Config) Load(path string) *Config {
	absConfigPath, _ := filepath.Abs(path)

	// Load configuration from file if it exists
	if _, err := os.Stat(absConfigPath); err == nil {
		data, err := os.ReadFile(absConfigPath)
		if err != nil {
			return c
		}

		err = yaml.Unmarshal(data, c)
		if err != nil {
			return c
		}
	}

	return c
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}

	// Validate dehydrated base dir
	if _, err := os.Stat(c.DehydratedBaseDir); os.IsNotExist(err) {
		return fmt.Errorf("dehydrated base dir does not exist: %s", c.DehydratedBaseDir)
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

// DomainsFile returns the path to the domains.txt file
func (c *Config) DomainsFile() string {
	return filepath.Join(c.DehydratedBaseDir, "domains.txt")
}
