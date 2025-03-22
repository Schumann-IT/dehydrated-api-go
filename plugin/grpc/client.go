package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/schumann-it/dehydrated-api-go/plugin"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements the plugin.Plugin interface using gRPC
type Client struct {
	conn     *grpc.ClientConn
	client   pb.PluginClient
	cmd      *exec.Cmd
	sockFile string
	mu       sync.RWMutex
}

// NewClient creates a new gRPC plugin client
func NewClient(pluginPath string, config map[string]string) (*Client, error) {
	// Create a temporary directory for the socket
	tmpDir, err := os.MkdirTemp("", "plugin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	sockFile := filepath.Join(tmpDir, "plugin.sock")

	// Start the plugin process
	cmd := exec.Command(pluginPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PLUGIN_SOCKET=%s", sockFile))
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Wait for the socket file to be created
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(sockFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Connect to the plugin
	conn, err := grpc.Dial(
		fmt.Sprintf("unix://%s", sockFile),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", sockFile, timeout)
		}),
	)
	if err != nil {
		cmd.Process.Kill()
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	client := pb.NewPluginClient(conn)

	// Initialize the plugin
	_, err = client.Initialize(context.Background(), &pb.InitializeRequest{
		Config: config,
	})
	if err != nil {
		conn.Close()
		cmd.Process.Kill()
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	return &Client{
		conn:     conn,
		client:   client,
		cmd:      cmd,
		sockFile: sockFile,
	}, nil
}

// Initialize implements plugin.Plugin
func (c *Client) Initialize(ctx context.Context, config map[string]string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return fmt.Errorf("client is nil")
	}

	_, err := c.client.Initialize(ctx, &pb.InitializeRequest{
		Config: config,
	})
	if err != nil {
		return plugin.ErrPluginError
	}
	return nil
}

// GetMetadata implements plugin.Plugin
func (c *Client) GetMetadata(ctx context.Context, domain string) (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return nil, fmt.Errorf("client is nil")
	}

	resp, err := c.client.GetMetadata(ctx, &pb.GetMetadataRequest{
		Domain: domain,
	})
	if err != nil {
		return nil, plugin.ErrPluginError
	}

	result := make(map[string]any)
	for k, v := range resp.Metadata {
		switch v.Kind.(type) {
		case *structpb.Value_StringValue:
			result[k] = v.GetStringValue()
		case *structpb.Value_NumberValue:
			result[k] = v.GetNumberValue()
		case *structpb.Value_BoolValue:
			result[k] = v.GetBoolValue()
		case *structpb.Value_ListValue:
			list := v.GetListValue()
			values := make([]any, len(list.Values))
			for i, item := range list.Values {
				switch item.Kind.(type) {
				case *structpb.Value_StringValue:
					values[i] = item.GetStringValue()
				case *structpb.Value_NumberValue:
					values[i] = item.GetNumberValue()
				case *structpb.Value_BoolValue:
					values[i] = item.GetBoolValue()
				case *structpb.Value_ListValue:
					values[i] = item.GetListValue()
				case *structpb.Value_StructValue:
					values[i] = item.GetStructValue()
				}
			}
			result[k] = values
		case *structpb.Value_StructValue:
			result[k] = v.GetStructValue()
		}
	}

	return result, nil
}

// Close implements plugin.Plugin
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if c.client == nil {
		return fmt.Errorf("client is nil")
	}

	_, err := c.client.Close(ctx, &pb.CloseRequest{})
	if err != nil {
		return fmt.Errorf("failed to close plugin: %w", err)
	}
	c.conn.Close()

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}

	if c.sockFile != "" {
		os.RemoveAll(filepath.Dir(c.sockFile))
	}

	return nil
}
