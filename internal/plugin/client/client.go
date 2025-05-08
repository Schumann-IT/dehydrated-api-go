package client

import (
	"context"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client represents a plugin client
type Client struct {
	client     *plugin.Client
	rpcClient  plugin.ClientProtocol
	plugin     pb.PluginClient
	logger     hclog.Logger
	socketPath string
}

// GRPCPlugin is the plugin implementation for go-plugin
type GRPCPlugin struct {
	plugin.GRPCPlugin
}

// GRPCServer is required by the go-plugin interface
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterPluginServer(s, &pb.UnimplementedPluginServer{})
	return nil
}

// GRPCClient is required by the go-plugin interface
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return pb.NewPluginClient(c), nil
}

// Server is required by the go-plugin interface
func (p *GRPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf("net/rpc not supported")
}

// Client is required by the go-plugin interface
func (p *GRPCPlugin) Client(broker *plugin.MuxBroker, rpcClient *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf("net/rpc not supported")
}

// NewClient creates a new plugin client
func NewClient(ctx context.Context, pluginName string, pluginPath string, config map[string]*structpb.Value) (*Client, error) {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   fmt.Sprintf("plugin-client-%s", pluginName),
		Level:  hclog.Trace,
		Output: os.Stdout,
	})

	// Create the plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "dehydrated-api-go",
		},
		Plugins: map[string]plugin.Plugin{
			pluginName: &GRPCPlugin{},
		},
		Cmd:    exec.Command(pluginPath),
		Logger: logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
	})

	// Connect to the plugin
	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to plugin %s: %w", pluginName, err)
	}

	// Get the plugin instance
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin %s: %w", pluginName, err)
	}

	// Type assert to our plugin interface
	p, ok := raw.(pb.PluginClient)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement Plugin interface")
	}

	if _, err := p.Initialize(ctx, &pb.InitializeRequest{
		Config: config,
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	return &Client{
		client:    client,
		rpcClient: rpcClient,
		plugin:    p,
		logger:    logger,
	}, nil
}

func (c *Client) Plugin() pb.PluginClient {
	return c.plugin
}

// Close closes the plugin client and cleans up resources
func (c *Client) Close() error {
	var errs []error

	// Create a test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := c.plugin.Close(ctx, &pb.CloseRequest{}); err != nil {
		return fmt.Errorf("failed to close plugin: %w", err)
	}

	// Kill the plugin process
	if c.client != nil {
		c.client.Kill()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}
	return nil
}
