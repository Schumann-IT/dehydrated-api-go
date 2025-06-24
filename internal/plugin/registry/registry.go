package registry

import (
	"context"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/client"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/manager"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"go.uber.org/zap"
)

type Registry struct {
	clients map[string]*client.Client
	manager *manager.Manager
	logger  *zap.Logger
}

func NewRegistry(cfg map[string]config.PluginConfig, logger *zap.Logger) *Registry {
	r := &Registry{
		clients: make(map[string]*client.Client),
		manager: manager.NewManager(logger, ""),
		logger:  logger,
	}

	for n, c := range cfg {
		if !c.Enabled {
			continue
		}
		r.register(n, c)
	}

	return r
}

func (r *Registry) register(name string, pluginConfig config.PluginConfig) {
	// Validate plugin configuration
	if err := pluginConfig.Validate(); err != nil {
		r.logger.Error("Invalid plugin configuration",
			zap.String("plugin", name),
			zap.Error(err))
		panic("Invalid plugin config: " + err.Error())
	}

	// Get plugin path using the new registry system or fallback to old system
	pluginPath, err := r.getPluginPath(pluginConfig)
	if err != nil {
		r.logger.Error("Failed to get plugin path",
			zap.String("plugin", name),
			zap.Error(err))
		panic("Failed to get plugin path: " + err.Error())
	}

	// Convert config to proto format
	cfg, err := pluginConfig.ToProto()
	if err != nil {
		r.logger.Error("Invalid plugin config",
			zap.String("plugin", name),
			zap.Error(err))
		panic("Invalid plugin config: " + err.Error())
	}

	// Create a new client
	c, err := client.NewClient(context.Background(), name, pluginPath, cfg)
	if err != nil {
		r.logger.Error("Failed to create plugin client",
			zap.String("plugin", name),
			zap.String("path", pluginPath),
			zap.Error(err))
		panic("Failed to create client: " + err.Error())
	}

	r.clients[name] = c
	r.logger.Info("Plugin registered successfully",
		zap.String("plugin", name),
		zap.String("path", pluginPath))
}

// getPluginPath gets the plugin path using the new registry system or falls back to the old system
func (r *Registry) getPluginPath(pluginConfig config.PluginConfig) (string, error) {
	// If new registry configuration is provided, use it
	if pluginConfig.Registry != nil {
		pluginRegistry, err := NewPluginRegistry(*pluginConfig.Registry, r.logger, r.manager)
		if err != nil {
			return "", err
		}
		return pluginRegistry.GetPluginPath()
	}

	// Fallback to old system
	return r.manager.GetPluginPath(pluginConfig)
}

func (r *Registry) Plugins() map[string]pb.PluginClient {
	p := make(map[string]pb.PluginClient)

	if r != nil {
		for n, c := range r.clients {
			p[n] = c.Plugin()
		}
	}

	return p
}

func (r *Registry) Close() {
	for name, c := range r.clients {
		r.logger.Debug("Closing plugin client", zap.String("plugin", name))
		c.Close()
	}
}
