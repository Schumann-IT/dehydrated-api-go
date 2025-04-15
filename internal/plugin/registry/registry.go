package registry

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/builtin"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/grpc"
	plugininterface "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"
)

// Registry manages plugin instances
type Registry struct {
	mu        sync.RWMutex
	plugins   map[string]plugininterface.Plugin
	Config    *dehydrated.Config
	closeOnce sync.Once
	closed    bool
}

// NewRegistry creates a new plugin registry
func NewRegistry(pluginConfig map[string]plugin.PluginConfig, cfg *dehydrated.Config) (*Registry, error) {
	r := &Registry{
		plugins: make(map[string]plugininterface.Plugin),
		Config:  cfg,
	}

	for name, pc := range pluginConfig {
		if !pc.Enabled {
			log.Info("skipping disabled plugin: ", name)
			continue
		}
		if err := r.LoadPlugin(name, pc); err != nil {
			return nil, fmt.Errorf("failed to load plugin %s: %w", name, err)
		}
	}

	return r, nil
}

// LoadPlugin loads a plugin from the given path with the provided configuration
func (r *Registry) LoadPlugin(name string, cfg plugin.PluginConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("registry is closed")
	}

	// Check if plugin is already loaded
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already loaded", name)
	}

	// Convert Config to map[string]string
	configMap := make(map[string]any)
	for k, v := range cfg.Config {
		if str, ok := v.(string); ok {
			configMap[k] = str
		}
	}

	// If no path is provided, try to load as built-in plugin
	if cfg.Path == "" {
		p, err := builtin.LoadPlugin(name)
		if err != nil {
			return fmt.Errorf("failed to load built-in plugin %s: %w", name, err)
		}
		err = p.Initialize(context.Background(), configMap, r.Config)
		if err != nil {
			return fmt.Errorf("failed to initialize built-in plugin %s: %w", name, err)
		}
		r.plugins[name] = p
		return nil
	}

	// Create new gRPC client
	client, err := grpc.NewClient(cfg.Path, configMap, r.Config)
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

	// After closing, all plugins are considered "not found"
	p, exists := r.plugins[name]
	if !exists || r.closed {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return p, nil
}

// GetPlugins returns all loaded plugins as a map of name to plugin
func (r *Registry) GetPlugins() map[string]plugininterface.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy of the plugins map
	plugins := make(map[string]plugininterface.Plugin, len(r.plugins))
	for name, p := range r.plugins {
		plugins[name] = p
	}

	return plugins
}

// Close closes all plugins
func (r *Registry) Close(ctx context.Context) error {
	var closeErr error
	r.closeOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.closed {
			return
		}

		for name, p := range r.plugins {
			if err := p.Close(ctx); err != nil {
				closeErr = fmt.Errorf("failed to close plugin %s: %w", name, err)
				return
			}
		}

		r.plugins = make(map[string]plugininterface.Plugin)
		r.closed = true
	})

	return closeErr
}
