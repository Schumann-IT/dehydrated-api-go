package registry

import (
	"context"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/client"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"go.uber.org/zap"
)

type Registry struct {
	clients map[string]*client.Client
	logger  *zap.Logger
}

func New(baseDir string, cfg map[string]config.PluginConfig, logger *zap.Logger) *Registry {
	r := &Registry{
		clients: make(map[string]*client.Client),
		logger:  logger,
	}

	err := cache.Prepare(baseDir)
	if err != nil {
		logger.Error("Failed to prepare plugin cache",
			zap.String("baseDir", baseDir),
			zap.Error(err))
		return r
	}

	for n, c := range cfg {
		if !c.Enabled {
			continue
		}

		_, err := cache.Add(n, c.Registry)
		if err != nil {
			r.logger.Error("Failed to add plugin to cache; ignoring plugin",
				zap.String("plugin", n),
				zap.Error(err))
			continue
		}

		// add log level configuration form the main logger, if not set specifically
		if _, ok := c.Config["logLevel"]; !ok {
			c.Config["logLevel"] = logger.Level().String()
		}

		pluginConfig, err := c.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert plugin config to proto; ignoring plugin",
				zap.String("plugin", n),
				zap.Error(err))
			continue
		}
		r.register(n, pluginConfig)
	}

	return r
}

func (r *Registry) register(name string, cfg map[string]*structpb.Value) {
	// Get plugin path using the new registry system or fallback to old system
	pluginPath, err := cache.Get(name)
	if err != nil {
		r.logger.Error("Failed to get plugin path; ignoring plugin",
			zap.String("plugin", name),
			zap.Error(err))
		return
	}

	// Create a new client
	c, err := client.NewClient(context.Background(), name, pluginPath, cfg)
	if err != nil {
		r.logger.Error("Failed to create plugin client; ignoring plugin",
			zap.String("plugin", name),
			zap.String("path", pluginPath),
			zap.Error(err))
		return
	}

	r.clients[name] = c
	r.logger.Info("Plugin registered successfully",
		zap.String("plugin", name),
		zap.String("path", pluginPath))
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
		err := c.Close()
		if err != nil {
			r.logger.Error("Failed to close plugin client", zap.String("plugin", name))
		}
	}
}
