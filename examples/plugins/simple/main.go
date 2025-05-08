package main

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/schumann-it/dehydrated-api-go/plugin/server"
	"google.golang.org/protobuf/types/known/structpb"
	"os"
)

// ExamplePlugin is a simple plugin implementation
type ExamplePlugin struct {
	proto.UnimplementedPluginServer
	logger hclog.Logger
	config map[string]interface{}
}

// Initialize implements the plugin.Plugin interface
func (p *ExamplePlugin) Initialize(ctx context.Context, req *proto.InitializeRequest) (*proto.InitializeResponse, error) {
	p.logger.Debug("Initialize called")
	// Convert the proto config to a map
	config := make(map[string]interface{})
	for k, v := range req.Config {
		config[k] = v.AsInterface()
	}
	p.config = config

	return &proto.InitializeResponse{}, nil
}

// GetMetadata implements the plugin.Plugin interface
func (p *ExamplePlugin) GetMetadata(ctx context.Context, req *proto.GetMetadataRequest) (*proto.GetMetadataResponse, error) {
	p.logger.Debug("GetMetadata called")

	res := &proto.GetMetadataResponse{
		Metadata: map[string]*structpb.Value{},
	}

	if req.GetDomainEntry().GetEnabled() {
		// Create example metadata
		res.Metadata = map[string]*structpb.Value{
			"domain":         structpb.NewStringValue(req.GetDomainEntry().GetDomain()),
			"name":           structpb.NewStringValue(p.config["name"].(string)),
			"example_key":    structpb.NewStringValue("example_value"),
			"example_number": structpb.NewNumberValue(42),
			"example_bool":   structpb.NewBoolValue(true),
		}
	}

	return res, nil
}

// Close implements the plugin.Plugin interface
func (p *ExamplePlugin) Close(ctx context.Context, req *proto.CloseRequest) (*proto.CloseResponse, error) {
	p.logger.Debug("Close called")
	return &proto.CloseResponse{}, nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	plugin := &ExamplePlugin{
		logger: logger,
	}

	server.NewPluginServer(plugin).Serve()
}
