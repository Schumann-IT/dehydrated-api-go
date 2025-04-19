// Package plugininterface defines the core interfaces for plugin implementations.
// It provides the contract that all plugins must follow to integrate with the dehydrated-api-go application.
package plugininterface

import (
	"context"
	"errors"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// ErrPluginError is returned when a plugin encounters an error during operation.
// This error is used to distinguish plugin-specific errors from other types of errors.
var ErrPluginError = errors.New("plugin error")

// Plugin defines the interface that all plugins must implement.
// This interface provides the contract for plugin initialization, metadata retrieval,
// and cleanup operations.
type Plugin interface {
	// Initialize is called when the plugin is loaded.
	// It sets up the plugin with the provided configuration.
	// The context can be used for cancellation and timeout control.
	// Returns an error if initialization fails.
	Initialize(ctx context.Context, config map[string]any) error

	// GetMetadata returns metadata for a domain entry.
	// This method is called to retrieve plugin-specific information about a domain.
	// The dehydratedConfig parameter provides access to the dehydrated configuration
	// for the specific domain being processed.
	// The context can be used for cancellation and timeout control.
	// Returns a map of metadata key-value pairs and an error if the operation fails.
	GetMetadata(ctx context.Context, entry model.DomainEntry, dehydratedConfig *dehydrated.Config) (map[string]any, error)

	// Close is called when the plugin is being unloaded.
	// It performs any necessary cleanup operations.
	// The context can be used for cancellation and timeout control.
	// Returns an error if cleanup fails.
	Close(ctx context.Context) error
}
