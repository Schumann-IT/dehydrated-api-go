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
	cache.Prepare(baseDir)

	r := &Registry{
		clients: make(map[string]*client.Client),
		logger:  logger,
	}

	for n, c := range cfg {
		if !c.Enabled {
			continue
		}

		cache.Add(n, c.Registry)

		pluginConfig, err := c.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert plugin config to proto",
				zap.String("plugin", n),
				zap.Error(err))
			panic("Failed to convert plugin config to proto: " + err.Error())
		}
		r.register(n, pluginConfig)
	}

	return r
}

func (r *Registry) register(name string, cfg map[string]*structpb.Value) {
	// Get plugin path using the new registry system or fallback to old system
	pluginPath, err := cache.Get(name)
	if err != nil {
		r.logger.Error("Failed to get plugin path",
			zap.String("plugin", name),
			zap.Error(err))
		panic("Failed to get plugin path: " + err.Error())
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
