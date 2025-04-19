// Package grpc provides gRPC-based plugin client implementation.
// It handles communication with plugins using gRPC protocol, including process management,
// socket file handling, and request/response conversion.
package grpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	plugininterface "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

// Client represents a gRPC plugin client.
// It manages the lifecycle of a plugin process and handles communication via gRPC.
type Client struct {
	client     pb.PluginClient
	conn       *grpc.ClientConn
	tmpDir     string
	socketFile string
	lastError  error
	mu         sync.RWMutex
	cmd        *exec.Cmd
}

// NewClient creates a new gRPC plugin client.
// It sets up a temporary directory for the socket file and starts the plugin process.
// The pluginPath parameter specifies the path to the plugin executable.
// Returns a new Client instance and an error if initialization fails.
func NewClient(pluginPath string, config map[string]any) (*Client, error) {
	// Convert config to structpb.Value to validate it
	_, err := convertToStructValue(config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Create a temporary directory for the socket file
	tmpDir, err := os.MkdirTemp("", "plugin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	socketFile := filepath.Join(tmpDir, "plugin.sock")

	// Start the plugin process
	cmd := exec.Command(pluginPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PLUGIN_SOCKET=%s", socketFile))

	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Wait for the socket file to be created
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			cmd.Process.Kill()
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("timeout waiting for plugin socket")
		default:
			if _, err := os.Stat(socketFile); err == nil {
				goto connected
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

connected:
	// Connect to the plugin
	conn, err := grpc.Dial(
		"unix://"+socketFile,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		cmd.Process.Kill()
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	client := &Client{
		client:     pb.NewPluginClient(conn),
		conn:       conn,
		tmpDir:     tmpDir,
		socketFile: socketFile,
	}

	// Initialize the plugin
	if err := client.Initialize(context.Background(), config); err != nil {
		client.Close(context.Background())
		return nil, err
	}

	return client, nil
}

// Initialize initializes the plugin with the provided configuration.
// It converts the configuration to protobuf format and sends it to the plugin.
// The context can be used for cancellation and timeout control.
// Returns an error if initialization fails.
func (c *Client) Initialize(ctx context.Context, config map[string]any) error {
	c.mu.RLock()
	if c.client == nil {
		c.mu.RUnlock()
		return fmt.Errorf("client is nil")
	}
	c.mu.RUnlock()

	// Convert config to structpb.Value map
	configValues, err := convertToStructValue(config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	req := &pb.InitializeRequest{
		Config: configValues,
	}

	_, err = c.client.Initialize(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", plugininterface.ErrPluginError, err)
	}
	return nil
}

// GetMetadata retrieves metadata for a domain entry from the plugin.
// It converts the domain entry to protobuf format and sends it to the plugin.
// The context can be used for cancellation and timeout control.
// Returns a map of metadata key-value pairs and an error if the operation fails.
func (c *Client) GetMetadata(ctx context.Context, entry *model.DomainEntry, dehydratedConfig *dehydrated.Config) (map[string]any, error) {
	c.mu.RLock()
	if c.client == nil {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client is nil")
	}
	c.mu.RUnlock()

	req := &pb.GetMetadataRequest{
		DomainEntry:      &entry.DomainEntry,
		DehydratedConfig: &dehydratedConfig.DehydratedConfig,
	}

	resp, err := c.client.GetMetadata(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", plugininterface.ErrPluginError, err)
	}

	return model.FromProto(resp), nil
}

// Close terminates the plugin process and cleans up resources.
// It sends a shutdown request to the plugin and waits for it to terminate.
// The context can be used for cancellation and timeout control.
// Returns an error if cleanup fails.
func (c *Client) Close(ctx context.Context) error {
	var errs []error

	if c.client != nil {
		if _, err := c.client.Close(ctx, &pb.CloseRequest{}); err != nil {
			errs = append(errs, fmt.Errorf("failed to close plugin: %w", err))
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill plugin process: %w", err))
		}
	}

	if c.tmpDir != "" {
		if err := os.RemoveAll(c.tmpDir); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove temporary directory: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}

	// Return the last error if one exists
	return c.lastError
}
