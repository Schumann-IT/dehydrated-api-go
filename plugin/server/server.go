package server

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"google.golang.org/grpc"
)

// PluginServer is the server implementation that plugins will use
type PluginServer struct {
	pb.UnimplementedPluginServer
	impl   pb.PluginServer
	config map[string]interface{}
}

// NewPluginServer creates a new plugin server
func NewPluginServer(impl pb.PluginServer) *PluginServer {
	return &PluginServer{
		impl: impl,
	}
}

// Serve starts the plugin server
func (p *PluginServer) Serve() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "dehydrated-api-go",
		},
		Plugins: map[string]plugin.Plugin{
			"dehydrated-plugin": &GRPCPlugin{Impl: p},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// GRPCPlugin is the plugin implementation for go-plugin
type GRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl pb.PluginServer
}

// GRPCServer is required by the go-plugin interface
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServer(s, p.Impl)
	return nil
}

// GRPCClient is required by the go-plugin interface
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pb.NewPluginClient(c), nil
}

// Client is required by the go-plugin interface
func (p *GRPCPlugin) Client(broker *plugin.MuxBroker, rpcClient *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf("net/rpc not supported")
}

// Server is required by the go-plugin interface
func (p *GRPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf("net/rpc not supported")
}

// Initialize implements the plugin.Plugin interface
func (p *PluginServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return p.impl.Initialize(ctx, req)
}

// GetMetadata implements the plugin.Plugin interface
func (p *PluginServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	return p.impl.GetMetadata(ctx, req)
}

// Close implements the plugin.Plugin interface
func (p *PluginServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return p.impl.Close(ctx, req)
}
