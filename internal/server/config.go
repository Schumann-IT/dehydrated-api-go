// Package server provides configuration management for the dehydrated-api-go server.
// It handles loading and validating server configuration from YAML files,
// including server settings, plugin configurations, and logging options.
package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for the dehydrated-api-go server.
// It includes settings for the HTTP server, plugin management, dehydrated client,
// and logging configuration.
type Config struct {
	// Server configuration
	Port int `yaml:"port"` // Port number for the HTTP server (1-65535)

	// Plugin configuration
	Plugins map[string]plugin.PluginConfig `yaml:"plugins"` // Map of plugin names to their configurations

	// Dehydrated configuration
	DehydratedBaseDir string `yaml:"dehydratedBaseDir"` // Base directory for dehydrated client files

	// DehydratedConfigFile specifies the path to the dehydrated configuration file.
	// This file is typically located under the base directory and contains
	// dehydrated client-specific settings.
	DehydratedConfigFile string `yaml:"dehydratedConfigFile"`

	// EnableWatcher determines whether the file watcher is active.
	// When enabled, the server monitors for changes in the dehydrated configuration.
	EnableWatcher bool `yaml:"enableWatcher"`

	// Logging configuration
	Logging *logger.Config `yaml:"logging"` // Configuration for the application logger
}

// NewConfig creates a new Config instance with default values.
// The default configuration includes:
// - Port: 3000
// - DehydratedBaseDir: "."
// - DehydratedConfigFile: "config"
// - EnableWatcher: false
// - Logging: default logger configuration
func NewConfig() *Config {
	return &Config{
		Port:                 3000,
		DehydratedBaseDir:    ".",
		DehydratedConfigFile: "config",
		Plugins:              make(map[string]plugin.PluginConfig),
		EnableWatcher:        false,
		Logging:              logger.DefaultConfig(),
	}
}

// WithBaseDir sets the dehydrated base directory in the configuration.
// This method returns the config instance for method chaining.
func (c *Config) WithBaseDir(dir string) *Config {
	c.DehydratedBaseDir = dir
	return c
}

// Load loads configuration from a YAML file and merges it with defaults.
// If the file doesn't exist or has invalid content, the default configuration is returned.
// The method merges non-zero values from the file with the existing configuration.
func (c *Config) Load(path string) *Config {
	absConfigPath, _ := filepath.Abs(path)

	// Create a temporary config to hold file contents
	fileConfig := &Config{
		Plugins: make(map[string]plugin.PluginConfig),
	}

	// Load configuration from file if it exists
	if _, err := os.Stat(absConfigPath); err == nil {
		data, err := os.ReadFile(absConfigPath)
		if err != nil {
			return c
		}

		err = yaml.Unmarshal(data, fileConfig)
		if err != nil {
			return c
		}

		// Merge non-zero values from file config
		if fileConfig.Port != 0 {
			c.Port = fileConfig.Port
		}
		if fileConfig.DehydratedBaseDir != "" {
			c.DehydratedBaseDir = fileConfig.DehydratedBaseDir
		}
		if fileConfig.DehydratedConfigFile != "" {
			c.DehydratedConfigFile = fileConfig.DehydratedConfigFile
		}
		if fileConfig.EnableWatcher {
			c.EnableWatcher = true
		}

		// Merge logging configuration
		if fileConfig.Logging != nil {
			if c.Logging == nil {
				c.Logging = logger.DefaultConfig()
			}
			if fileConfig.Logging.Level != "" {
				c.Logging.Level = fileConfig.Logging.Level
			}
			if fileConfig.Logging.Encoding != "" {
				c.Logging.Encoding = fileConfig.Logging.Encoding
			}
			if fileConfig.Logging.OutputPath != "" {
				c.Logging.OutputPath = fileConfig.Logging.OutputPath
			}
		}

		// Merge plugin configurations
		for name, p := range fileConfig.Plugins {
			// If plugin doesn't exist in defaults, create it
			if _, exists := c.Plugins[name]; !exists {
				c.Plugins[name] = plugin.PluginConfig{
					Config: make(map[string]any),
				}
			}

			// Get reference to existing plugin config
			existingPlugin := c.Plugins[name]

			// Merge plugin settings
			if p.Enabled {
				existingPlugin.Enabled = true
			}
			if p.Path != "" {
				existingPlugin.Path = p.Path
			}
			if p.Config != nil {
				// Merge plugin config maps
				if existingPlugin.Config == nil {
					existingPlugin.Config = make(map[string]any)
				}
				for k, v := range p.Config {
					existingPlugin.Config[k] = v
				}
			}

			// Update the plugin in the main config
			c.Plugins[name] = existingPlugin
		}
	}

	if !filepath.IsAbs(c.DehydratedBaseDir) {
		c.DehydratedBaseDir = filepath.Join(filepath.Dir(absConfigPath), c.DehydratedBaseDir)
	}

	if !filepath.IsAbs(c.DehydratedConfigFile) {
		c.DehydratedConfigFile = filepath.Join(c.DehydratedBaseDir, c.DehydratedConfigFile)
	}

	return c
}

// Validate checks if the configuration is valid and returns an error if any issues are found.
// It validates:
// - Port number (must be between 1 and 65535)
// - Dehydrated base directory (must exist)
// - Plugin configurations (paths must exist and be absolute)
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
	for name, p := range c.Plugins {
		if !p.Enabled {
			continue
		}

		if p.Path == "" {
			return fmt.Errorf("plugin path is required for enabled plugin: %s", name)
		}

		// Check if plugin path exists and is executable
		if _, err := os.Stat(p.Path); err != nil {
			return fmt.Errorf("plugin path does not exist: %s", p.Path)
		}

		// Check if plugin path is absolute
		if !filepath.IsAbs(p.Path) {
			return fmt.Errorf("plugin path must be absolute: %s", p.Path)
		}
	}

	return nil
}

// DomainsFile returns the absolute path to the domains.txt file.
// This file contains the list of domains managed by the dehydrated client.
func (c *Config) DomainsFile() string {
	return filepath.Join(c.DehydratedBaseDir, "domains.txt")
}

func (c *Config) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(b)
}
