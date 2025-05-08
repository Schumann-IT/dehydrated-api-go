// Package server provides configuration management for the dehydrated-api-go server.
// It handles loading and validating server configuration from YAML files,
// including server settings, plugin configurations, and logging options.
package server

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/auth"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for the dehydrated-api-go server.
// It includes settings for the HTTP server, plugin management, dehydrated client,
// and logging configuration.
type Config struct {
	// Server configuration
	Port int `yaml:"port"` // Port number for the HTTP server (1-65535)

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

	// Authentication configuration
	Auth *auth.Config `yaml:"auth"` // Azure AD authentication configuration

	Plugins map[string]config.PluginConfig `yaml:"plugins"`

	err          error
	parsedConfig *Config
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
		EnableWatcher:        false,
	}
}

// WithBaseDir sets the dehydrated base directory in the configuration.
// This method returns the config instance for method chaining.
func (c *Config) WithBaseDir(dir string) *Config {
	c.DehydratedBaseDir = dir
	return c
}

func (c *Config) parse(path string) *Config {
	fc := &Config{}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			fc.err = err
			return fc
		}

		err = yaml.Unmarshal(data, fc)
		if err != nil {
			fc.err = err
			return fc
		}
	}

	c.parsedConfig = fc

	return fc
}

// Load loads configuration from a YAML file and merges it with defaults.
// If the file doesn't exist or has invalid content, the default configuration is returned.
// The method merges non-zero values from the file with the existing configuration.
func (c *Config) Load(path string) *Config {
	absConfigPath, _ := filepath.Abs(path)
	fc := c.parse(absConfigPath)
	if fc.err != nil {
		c.err = fc.err
		return c
	}

	// Merge non-zero values from file config
	if fc.Port != 0 {
		c.Port = fc.Port
	}
	if fc.DehydratedBaseDir != "" {
		c.DehydratedBaseDir = fc.DehydratedBaseDir
	}
	if fc.DehydratedConfigFile != "" {
		c.DehydratedConfigFile = fc.DehydratedConfigFile
	}
	if fc.EnableWatcher {
		c.EnableWatcher = true
	}

	// Merge logging configuration
	if fc.Logging != nil {
		if c.Logging == nil {
			c.Logging = &logger.Config{}
		}
		if fc.Logging.Level != "" {
			c.Logging.Level = fc.Logging.Level
		}
		if fc.Logging.Encoding != "" {
			c.Logging.Encoding = fc.Logging.Encoding
		}
		if fc.Logging.OutputPath != "" {
			c.Logging.OutputPath = fc.Logging.OutputPath
		}
	}

	// Merge auth configuration
	if fc.Auth != nil {
		c.Auth = fc.Auth
	}

	// Merge plugin config
	if fc.Plugins != nil {
		c.Plugins = fc.Plugins
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
