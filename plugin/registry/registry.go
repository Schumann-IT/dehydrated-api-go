package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated/model"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/plugin/builtin/timestamp"
	"github.com/schumann-it/dehydrated-api-go/plugin/grpc"
	plugininterface "github.com/schumann-it/dehydrated-api-go/plugin/interface"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// Registry manages plugin instances
type Registry struct {
	plugins map[string]plugininterface.Plugin
	config  *dehydrated.Config
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry(cfg *dehydrated.Config) *Registry {
	return &Registry{
		plugins: make(map[string]plugininterface.Plugin),
		config:  cfg,
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

	// If no path is provided, try to load as built-in plugin
	if path == "" {
		plugin, err := r.loadBuiltinPlugin(name)
		if err != nil {
			return fmt.Errorf("failed to load built-in plugin %s: %w", name, err)
		}
		r.plugins[name] = plugin
		return nil
	}

	// Convert config to map[string]string
	configMap := make(map[string]any)
	for k, v := range config {
		if str, ok := v.(string); ok {
			configMap[k] = str
		}
	}

	// Create new gRPC client
	client, err := grpc.NewClient(path, configMap, r.config)
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
	case "timestamp":
		server = timestamp.New()
	default:
		return nil, fmt.Errorf("built-in plugin %s not found", name)
	}

	// Create a wrapper for the built-in plugin
	wrapper := &builtinWrapper{
		server: server,
		config: r.config,
	}

	return wrapper, nil
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

// builtinWrapper wraps a built-in plugin to implement the Plugin interface
type builtinWrapper struct {
	server pb.PluginServer
	config *dehydrated.Config
}

func (w *builtinWrapper) Initialize(ctx context.Context, config map[string]any, dehydratedConfig *dehydrated.Config) error {
	// Convert config to map[string]*structpb.Value
	configMap := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return fmt.Errorf("failed to convert config value for key %s: %w", k, err)
		}
		configMap[k] = value
	}

	// Convert dehydrated config
	dehydratedConfigProto := &pb.DehydratedConfig{
		BaseDir:       dehydratedConfig.BaseDir,
		CertDir:       dehydratedConfig.CertDir,
		DomainsDir:    dehydratedConfig.DomainsDir,
		ChallengeType: dehydratedConfig.ChallengeType,
		Ca:            dehydratedConfig.CA,
	}

	req := &pb.InitializeRequest{
		Config:           configMap,
		DehydratedConfig: dehydratedConfigProto,
	}
	_, err := w.server.Initialize(ctx, req)
	return err
}

func (w *builtinWrapper) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]any, error) {
	// Convert metadata to structpb.Value map
	metadataValues, err := plugininterface.ConvertToStructValue(entry.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to convert metadata: %w", err)
	}

	req := &pb.GetMetadataRequest{
		Domain:           entry.Domain,
		AlternativeNames: entry.AlternativeNames,
		Alias:            entry.Alias,
		Enabled:          entry.Enabled,
		Comment:          entry.Comment,
		Metadata:         metadataValues,
	}
	resp, err := w.server.GetMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert metadata to map[string]any
	metadata := make(map[string]any)
	for k, v := range resp.Metadata {
		switch v.GetKind().(type) {
		case *structpb.Value_StringValue:
			metadata[k] = v.GetStringValue()
		case *structpb.Value_NumberValue:
			metadata[k] = v.GetNumberValue()
		case *structpb.Value_BoolValue:
			metadata[k] = v.GetBoolValue()
		case *structpb.Value_StructValue:
			metadata[k] = v.GetStructValue()
		}
	}
	return metadata, nil
}

func (w *builtinWrapper) Close(ctx context.Context) error {
	req := &pb.CloseRequest{}
	_, err := w.server.Close(ctx, req)
	return err
}
