package plugininterface

import (
	"context"
	"errors"
)

// ErrPluginError is returned when a plugin encounters an error
var ErrPluginError = errors.New("plugin error")

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Initialize is called when the plugin is loaded
	Initialize(ctx context.Context, config map[string]any) error

	// GetMetadata returns metadata for a domain entry
	GetMetadata(ctx context.Context, domain string) (map[string]any, error)

	// Close is called when the plugin is being unloaded
	Close(ctx context.Context) error
}
