package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/manager"
	"go.uber.org/zap"
)

// RegistryType represents the type of plugin registry
type RegistryType string

const (
	RegistryTypeLocal  RegistryType = "local"
	RegistryTypeGitHub RegistryType = "github"
)

// RegistryConfig represents the configuration for a plugin registry
type RegistryConfig struct {
	Type   RegistryType   `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

// PluginRegistry interface defines methods for plugin registry implementations
type PluginRegistry interface {
	// GetPluginPath returns the path to the plugin executable
	GetPluginPath() (string, error)

	// Validate validates the registry configuration
	Validate() error
}

// LocalRegistry represents a local plugin registry
type LocalRegistry struct {
	config map[string]any
	logger *zap.Logger
}

// NewLocalRegistry creates a new local registry
func NewLocalRegistry(config map[string]any, logger *zap.Logger) *LocalRegistry {
	return &LocalRegistry{
		config: config,
		logger: logger,
	}
}

// GetPluginPath returns the absolute path to the local plugin
func (r *LocalRegistry) GetPluginPath() (string, error) {
	path, ok := r.config["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required for local registry")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("plugin file does not exist: %s", absPath)
	}

	return absPath, nil
}

// Validate validates the local registry configuration
func (r *LocalRegistry) Validate() error {
	path, ok := r.config["path"].(string)
	if !ok {
		return fmt.Errorf("path is required for local registry")
	}

	if path == "" {
		return fmt.Errorf("path cannot be empty for local registry")
	}

	return nil
}

// GitHubRegistry represents a GitHub plugin registry
type GitHubRegistry struct {
	config  map[string]any
	logger  *zap.Logger
	manager *manager.Manager
}

// NewGitHubRegistry creates a new GitHub registry
func NewGitHubRegistry(config map[string]any, logger *zap.Logger, manager *manager.Manager) *GitHubRegistry {
	return &GitHubRegistry{
		config:  config,
		logger:  logger,
		manager: manager,
	}
}

// GetPluginPath downloads and returns the path to the GitHub plugin
func (r *GitHubRegistry) GetPluginPath() (string, error) {
	return r.manager.DownloadGitHubPlugin(r.config)
}

// Validate validates the GitHub registry configuration
func (r *GitHubRegistry) Validate() error {
	repository, ok := r.config["repository"].(string)
	if !ok {
		return fmt.Errorf("repository is required for GitHub registry")
	}

	if repository == "" {
		return fmt.Errorf("repository cannot be empty for GitHub registry")
	}

	// Validate repository format (should be owner/repo)
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid GitHub repository format: %s (expected owner/repo)", repository)
	}

	return nil
}

// NewPluginRegistry creates a new registry based on the configuration
func NewPluginRegistry(registryConfig config.RegistryConfig, logger *zap.Logger, manager *manager.Manager) (PluginRegistry, error) {
	switch registryConfig.Type {
	case "local":
		return NewLocalRegistry(registryConfig.Config, logger), nil
	case "github":
		return NewGitHubRegistry(registryConfig.Config, logger, manager), nil
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", registryConfig.Type)
	}
}
