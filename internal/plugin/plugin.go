package plugin

import (
	"github.com/schumann-it/dehydrated-api-go/internal/config"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string

	// Initialize sets up the plugin with the given configuration
	Initialize(cfg *config.Config) error

	// EnrichDomainEntry allows the plugin to add information to a domain entry
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
