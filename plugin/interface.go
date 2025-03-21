package plugin

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

	// EnrichDomainEntry is called for each domain
	EnrichDomainEntry(ctx context.Context, domain *Domain) (*Metadata, error)

	// Close is called when shutting down
	Close(ctx context.Context) error
}

// Domain represents a dehydrated domain
type Domain struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Metadata represents plugin-provided metadata
type Metadata struct {
	Values map[string]string `json:"values"`
}

// NewDomain creates a new Domain instance
func NewDomain(name string) *Domain {
	return &Domain{
		Name:     name,
		Metadata: make(map[string]string),
	}
}

// NewMetadata creates a new Metadata instance
func NewMetadata() *Metadata {
	return &Metadata{
		Values: make(map[string]string),
	}
}
