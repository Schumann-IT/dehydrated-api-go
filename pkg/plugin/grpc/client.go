package grpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
func (c *Client) Initialize(config map[string]string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, err := c.client.Initialize(context.Background(), &pb.InitializeRequest{
		Config: config,
	})
	return err
}

// GetMetadata implements plugin.Plugin
func (c *Client) GetMetadata(domain string, config map[string]string) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, err := c.client.GetMetadata(context.Background(), &pb.GetMetadataRequest{
		Domain: domain,
		Config: config,
	})
	if err != nil {
		return nil, err
	}

	return resp.Metadata, nil
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
