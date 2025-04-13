package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/builtin/openssl"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/grpc"
	plugininterface "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
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
func NewRegistry(pluginConfig map[string]internal.PluginConfig, cfg *dehydrated.Config) (*Registry, error) {
	r := &Registry{
		plugins: make(map[string]plugininterface.Plugin),
		Config:  cfg,
	}

	for name, pc := range pluginConfig {
		if err := r.LoadPlugin(name, pc); err != nil {
			return nil, fmt.Errorf("failed to load plugin %s: %w", name, err)
		}
	}

	return r, nil
}

// LoadPlugin loads a plugin from the given path with the provided configuration
func (r *Registry) LoadPlugin(name string, cfg internal.PluginConfig) error {
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
		plugin, err := r.loadBuiltinPlugin(name)
		if err != nil {
			return fmt.Errorf("failed to load built-in plugin %s: %w", name, err)
		}
		err = plugin.Initialize(context.Background(), configMap, r.Config)
		if err != nil {
			return fmt.Errorf("failed to initialize built-in plugin %s: %w", name, err)
		}
		r.plugins[name] = plugin
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

// loadBuiltinPlugin loads a built-in plugin by name
func (r *Registry) loadBuiltinPlugin(name string) (plugininterface.Plugin, error) {
	var server pb.PluginServer

	switch name {
	case "openssl":
		server = openssl.New()
	default:
		return nil, fmt.Errorf("built-in plugin %s not found", name)
	}

	// Create a wrapper for the built-in plugin
	wrapper := &builtinWrapper{
		server: server,
		config: r.Config,
	}

	return wrapper, nil
}

// GetPlugin returns a plugin by name
func (r *Registry) GetPlugin(name string) (plugininterface.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// After closing, all plugins are considered "not found"
	plugin, exists := r.plugins[name]
	if !exists || r.closed {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// GetPlugins returns all loaded plugins as a map of name to plugin
func (r *Registry) GetPlugins() map[string]plugininterface.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy of the plugins map
	plugins := make(map[string]plugininterface.Plugin, len(r.plugins))
	for name, plugin := range r.plugins {
		plugins[name] = plugin
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

		for name, plugin := range r.plugins {
			if err := plugin.Close(ctx); err != nil {
				closeErr = fmt.Errorf("failed to close plugin %s: %w", name, err)
				return
			}
		}

		r.plugins = make(map[string]plugininterface.Plugin)
		r.closed = true
	})

	return closeErr
}

// builtinWrapper wraps a built-in plugin to implement the Plugin interface
type builtinWrapper struct {
	server pb.PluginServer
	config *dehydrated.Config
}

func (w *builtinWrapper) Initialize(ctx context.Context, config map[string]any, dehydratedConfig *dehydrated.Config) error {
	// Convert Config to map[string]*structpb.Value
	configMap := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return fmt.Errorf("failed to convert Config value for key %s: %w", k, err)
		}
		configMap[k] = value
	}

	// Convert dehydrated Config
	dehydratedConfigProto := &pb.DehydratedConfig{
		BaseDir:       dehydratedConfig.BaseDir,
		CertDir:       dehydratedConfig.CertDir,
		DomainsDir:    dehydratedConfig.DomainsDir,
		ChallengeType: dehydratedConfig.ChallengeType,
		Ca:            dehydratedConfig.Ca,
	}

	req := &pb.InitializeRequest{
		Config:           configMap,
		DehydratedConfig: dehydratedConfigProto,
	}
	_, err := w.server.Initialize(ctx, req)
	return err
}

func (w *builtinWrapper) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]any, error) {
	req := entry.ToProto()
	resp, err := w.server.GetMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return model.FromProto(resp).Metadata, nil
}

func (w *builtinWrapper) Close(ctx context.Context) error {
	req := &pb.CloseRequest{}
	_, err := w.server.Close(ctx, req)
	return err
}
