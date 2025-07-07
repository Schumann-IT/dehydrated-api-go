package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/schumann-it/dehydrated-api-go/plugin/server"
)

// ExamplePlugin is a simple plugin implementation
type ExamplePlugin struct {
	proto.UnimplementedPluginServer
	logger hclog.Logger
	config *proto.PluginConfig
}

// Initialize implements the plugin.Plugin interface
func (p *ExamplePlugin) Initialize(_ context.Context, req *proto.InitializeRequest) (*proto.InitializeResponse, error) {
	p.logger.Debug("Initialize called")
	p.config.FromProto(req.Config)
	return &proto.InitializeResponse{}, nil
}

// GetMetadata implements the plugin.Plugin interface
func (p *ExamplePlugin) GetMetadata(_ context.Context, req *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	p.logger.Debug("GetMetadata called", "domain", req.GetDomainEntry().GetDomain())

	// Create a new Metadata for the response
	metadata := proto.NewMetadata()

	if req.GetDomainEntry().GetEnabled() {
		// Get the name from config
		name, err := p.config.GetString("name")
		if err != nil {
			metadata.SetError(fmt.Sprintf("failed to get name from config: %v", err))
			return metadata.ToGetMetadataResponse()
		}

		// Set example metadata
		metadata.Set("domain", req.GetDomainEntry().GetDomain())
		metadata.Set("name", name)
		metadata.Set("example_key", "example_value")
		metadata.Set("example_number", 42)
		metadata.Set("example_bool", true)
		metadata.SetMap("config", req.GetDehydratedConfig())
	}

	return metadata.ToGetMetadataResponse()
}

// Close implements the plugin.Plugin interface
func (p *ExamplePlugin) Close(_ context.Context, _ *proto.CloseRequest) (*proto.CloseResponse, error) {
	p.logger.Debug("Close called")
	return &proto.CloseResponse{}, nil
}

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "example-plugin",
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	// Create plugin instance
	plugin := &ExamplePlugin{
		logger: logger,
		config: proto.NewPluginConfig(),
	}

	// Create and start a plugin server
	server.NewPluginServer(plugin).Serve()
}
