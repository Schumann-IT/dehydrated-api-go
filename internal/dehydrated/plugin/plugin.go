package plugin

import (
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
)

// Plugin is the interface that must be implemented by all plugins
type Plugin interface {
	// Initialize initializes the plugin with configuration
	Initialize(cfg *config.Config) error

	// EnrichDomainEntry enriches a domain entry with additional metadata
	EnrichDomainEntry(entry *model.DomainEntry) error

	// Close cleans up any resources used by the plugin
	Close() error
}

// Registry manages the registered plugins
type Registry struct {
	plugins []Plugin
	config  *config.Config
}

// NewRegistry creates a new plugin registry
func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{
		plugins: make([]Plugin, 0),
		config:  cfg,
	}
}

// Register adds a plugin to the registry and initializes it
func (r *Registry) Register(p Plugin) error {
	if err := p.Initialize(r.config); err != nil {
		return err
	}
	r.plugins = append(r.plugins, p)
	return nil
}

// EnrichDomainEntry runs all plugins to enrich the domain entry
func (r *Registry) EnrichDomainEntry(entry *model.DomainEntry) error {
	for _, p := range r.plugins {
		if err := p.EnrichDomainEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

// Close cleans up all plugins
func (r *Registry) Close() error {
	for _, p := range r.plugins {
		if err := p.Close(); err != nil {
			return err
		}
	}
	return nil
}
