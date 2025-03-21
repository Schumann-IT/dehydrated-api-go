package plugininterface

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

// PluginError represents an error that occurred during plugin operations
type PluginError struct {
	Name    string
	Message string
	Cause   error
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin %s: %s: %v", e.Name, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin %s: %s", e.Name, e.Message)
}

func (e *PluginError) Unwrap() error {
	return e.Cause
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	Name     string
	Enabled  bool
	Settings map[string]interface{}
}

// Plugin defines the interface that all plugins must implement.
// This interface provides the core functionality required for plugins
// to integrate with the dehydrated API system.
type Plugin interface {
	// Name returns the unique name of the plugin.
	// This name is used to identify the plugin in the system
	// and must be unique across all loaded plugins.
	Name() string

	// Initialize sets up the plugin with the given configuration.
	// This method is called when the plugin is first loaded and
	// should perform any necessary setup or validation.
	Initialize(cfg *service.Config) error

	// EnrichDomainEntry allows the plugin to add information to a domain entry.
	// This method is called for each domain entry and should add any
	// plugin-specific information to the entry's metadata.
	EnrichDomainEntry(entry *model.DomainEntry) error

	// Close cleans up any resources used by the plugin.
	// This method is called when the plugin is being unloaded and
	// should perform any necessary cleanup.
	Close() error
}
