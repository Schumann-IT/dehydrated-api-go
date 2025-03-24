package plugininterface

import (
	"context"
	"errors"
	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated/model"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"
)

// ErrPluginError is returned when a plugin encounters an error
var ErrPluginError = errors.New("plugin error")

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Initialize is called when the plugin is loaded
	// config contains plugin-specific configuration
	// dehydratedConfig contains the dehydrated configuration
	Initialize(ctx context.Context, config map[string]any, dehydratedConfig *dehydrated.Config) error

	// GetMetadata returns metadata for a domain entry
	GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]any, error)

	// Close is called when the plugin is being unloaded
	Close(ctx context.Context) error
}
