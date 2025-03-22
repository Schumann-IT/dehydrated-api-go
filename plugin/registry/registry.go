package registry

import (
	"context"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/plugin/grpc"
	"github.com/schumann-it/dehydrated-api-go/plugin/interface"
	"sync"
)

// Registry manages plugin instances
type Registry struct {
	plugins map[string]plugininterface.Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]plugininterface.Plugin),
	}
}

// LoadPlugin loads a plugin from the given path with the provided configuration
func (r *Registry) LoadPlugin(name string, path string, config map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if plugin is already loaded
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already loaded", name)
	}

	// Convert config to map[string]string
	configMap := make(map[string]any)
	for k, v := range config {
		if str, ok := v.(string); ok {
			configMap[k] = str
		}
	}

	// Create new gRPC client
	client, err := grpc.NewClient(path, configMap)
	if err != nil {
		return fmt.Errorf("failed to create plugin client: %w", err)
	}

	// Store plugin
	r.plugins[name] = client
	return nil
}

// GetPlugin returns a plugin by name
func (r *Registry) GetPlugin(name string) (plugininterface.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// GetPlugins returns all loaded plugins
func (r *Registry) GetPlugins() []plugininterface.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]plugininterface.Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// Close closes all plugins
func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, plugin := range r.plugins {
		if err := plugin.Close(ctx); err != nil {
			return fmt.Errorf("failed to close plugin %s: %w", name, err)
		}
	}

	r.plugins = make(map[string]plugininterface.Plugin)
	return nil
}
