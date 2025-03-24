package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/logger"
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

	// Weather to enable file watcher
	EnableWatcher bool `yaml:"enableWatcher"`

	// Logging configuration
	Logging *logger.Config `yaml:"logging"`
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	Enabled bool           `yaml:"enabled"`
	Path    string         `yaml:"path"`
	Config  map[string]any `yaml:"config"`
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Port:              3000,
		DehydratedBaseDir: ".",
		Plugins:           make(map[string]PluginConfig),
		EnableWatcher:     false,
		Logging:           logger.DefaultConfig(),
	}
}

// WithBaseDir sets the dehydrated base directory
func (c *Config) WithBaseDir(dir string) *Config {
	c.DehydratedBaseDir = dir
	return c
}

// Load loads configuration from a YAML file and merges it with defaults
func (c *Config) Load(path string) *Config {
	absConfigPath, _ := filepath.Abs(path)

	// Create a temporary config to hold file contents
	fileConfig := &Config{
		Plugins: make(map[string]PluginConfig),
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
		for name, plugin := range fileConfig.Plugins {
			// If plugin doesn't exist in defaults, create it
			if _, exists := c.Plugins[name]; !exists {
				c.Plugins[name] = PluginConfig{
					Config: make(map[string]any),
				}
			}

			// Get reference to existing plugin config
			existingPlugin := c.Plugins[name]

			// Merge plugin settings
			if plugin.Enabled {
				existingPlugin.Enabled = true
			}
			if plugin.Path != "" {
				existingPlugin.Path = plugin.Path
			}
			if plugin.Config != nil {
				// Merge plugin config maps
				if existingPlugin.Config == nil {
					existingPlugin.Config = make(map[string]any)
				}
				for k, v := range plugin.Config {
					existingPlugin.Config[k] = v
				}
			}

			// Update the plugin in the main config
			c.Plugins[name] = existingPlugin
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
