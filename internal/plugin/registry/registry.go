package registry

import (
	"context"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/client"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"path/filepath"
)

type Registry struct {
	clients map[string]*client.Client
}

func NewRegistry(cfg map[string]config.PluginConfig) *Registry {
	r := &Registry{
		clients: make(map[string]*client.Client),
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
	// Build the example plugin
	pluginPath, err := filepath.Abs(pluginConfig.Path)
	if err != nil {
		panic("Plugin path must be absolute")
	}

	cfg, err := pluginConfig.ToProto()
	if err != nil {
		panic("Invalud plugin config: " + err.Error() + "")
	}

	// Create a new client
	c, err := client.NewClient(context.Background(), name, pluginPath, cfg)
	if err != nil {
		panic("Failed to create client: " + err.Error() + "")
	}

	r.clients[name] = c
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
	for _, c := range r.clients {
		c.Close()
	}
}
