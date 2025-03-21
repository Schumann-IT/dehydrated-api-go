package grpc

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/dehydrated/plugin/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client implements the plugin interface using gRPC
type Client struct {
	client pb.PluginClient
	conn   *grpc.ClientConn
	cmd    *exec.Cmd
	name   string
}

// NewClient creates a new gRPC plugin client
func NewClient(name, path string, args []string) (*Client, error) {
	// Start the plugin process
	cmd := exec.Command(path, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Create gRPC connection
	conn, err := grpc.Dial("pipe",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return &pipeConn{stdin: stdin, stdout: stdout}, nil
		}),
	)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client := &Client{
		client: pb.NewPluginClient(conn),
		conn:   conn,
		cmd:    cmd,
		name:   name,
	}

	// Initialize the plugin
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.client.Initialize(ctx, &pb.InitializeRequest{}); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	return client, nil
}

// EnrichDomainEntry sends a request to enrich a domain entry
func (c *Client) EnrichDomainEntry(ctx context.Context, domain string) (map[string]string, error) {
	req := &pb.EnrichDomainEntryRequest{
		Entry: &pb.DomainEntry{
			Domain: domain,
		},
	}

	resp, err := c.client.EnrichDomainEntry(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich domain entry: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return resp.Entry.Metadata, nil
}

// Close closes the gRPC connection and kills the plugin process
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	return nil
}
