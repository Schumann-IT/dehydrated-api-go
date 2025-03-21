package registry

import (
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"sync"

	plugininterface "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"
)

// Registry manages the registered plugins and provides thread-safe access to them.
type Registry struct {
	plugins []plugininterface.Plugin
	config  *service.Config
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry with the given configuration.
func NewRegistry(cfg *service.Config) *Registry {
	return &Registry{
		plugins: make([]plugininterface.Plugin, 0),
		config:  cfg,
	}
}

// Register adds a plugin to the registry and initializes it.
// This method is thread-safe and will return an error if the plugin
// fails to initialize or if a plugin with the same name is already registered.
func (r *Registry) Register(p plugininterface.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if plugin with same name is already registered
	for _, existing := range r.plugins {
		if existing.Name() == p.Name() {
			return &plugininterface.PluginError{
				Name:    p.Name(),
				Message: "plugin already registered",
			}
		}
	}

	// Initialize the plugin
	if err := p.Initialize(r.config); err != nil {
		return &plugininterface.PluginError{
			Name:    p.Name(),
			Message: "failed to initialize plugin",
			Cause:   err,
		}
	}

	r.plugins = append(r.plugins, p)
	return nil
}

// EnrichDomainEntry runs all plugins to enrich the domain entry.
// This method is thread-safe and will return an error if any plugin
// fails to enrich the entry.
func (r *Registry) EnrichDomainEntry(entry *model.DomainEntry) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.plugins {
		if err := p.EnrichDomainEntry(entry); err != nil {
			return &plugininterface.PluginError{
				Name:    p.Name(),
				Message: "failed to enrich domain entry",
				Cause:   err,
			}
		}
	}
	return nil
}

// Close cleans up all plugins.
// This method is thread-safe and will return an error if any plugin
// fails to close properly.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for _, p := range r.plugins {
		if err := p.Close(); err != nil {
			lastErr = &plugininterface.PluginError{
				Name:    p.Name(),
				Message: "failed to close plugin",
				Cause:   err,
			}
		}
	}
	// Clear the plugins slice after closing all plugins
	r.plugins = make([]plugininterface.Plugin, 0)
	return lastErr
}

// GetPlugin returns a plugin by name.
// This method is thread-safe and returns nil if no plugin with the
// given name is found.
func (r *Registry) GetPlugin(name string) plugininterface.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.plugins {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// ListPlugins returns a list of all registered plugin names.
// This method is thread-safe.
func (r *Registry) ListPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, len(r.plugins))
	for i, p := range r.plugins {
		names[i] = p.Name()
	}
	return names
}
