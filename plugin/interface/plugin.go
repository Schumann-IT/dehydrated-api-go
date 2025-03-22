package plugininterface

import "context"

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Initialize is called when the plugin is loaded
	Initialize(ctx context.Context, config map[string]string) error

	// GetMetadata returns metadata for a domain entry
	GetMetadata(ctx context.Context, domain string) (map[string]any, error)

	// Close is called when the plugin is being unloaded
	Close(ctx context.Context) error
}
