package plugininterface

import "context"

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Initialize is called when the plugin is loaded
	Initialize(config map[string]string) error

	// GetMetadata returns metadata for a domain entry
	GetMetadata(domain string, config map[string]string) (map[string]string, error)

	// Close is called when the plugin is being unloaded
	Close(ctx context.Context) error
}
