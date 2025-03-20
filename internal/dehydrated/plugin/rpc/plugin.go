package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin"
	"google.golang.org/grpc"
)

// Client implements the Plugin interface and communicates with an RPC plugin
type Client struct {
	conn   *grpc.ClientConn
	client PluginClient
	cmd    *exec.Cmd
}

// NewClient creates a new RPC plugin client
func NewClient(pluginPath string) (plugin.Plugin, error) {
	// Start plugin process
	cmd := exec.Command(pluginPath)
	cmd.Stderr = os.Stderr

	// Create pipe for plugin to write its address
	addrPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Read plugin address
	buf := make([]byte, 1024)
	n, err := addrPipe.Read(buf)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to read plugin address: %w", err)
	}
	addr := string(buf[:n])

	// Connect to plugin
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	return &Client{
		conn:   conn,
		client: NewPluginClient(conn),
		cmd:    cmd,
	}, nil
}

// Initialize initializes the plugin with configuration
func (c *Client) Initialize(cfg *config.Config) error {
	resp, err := c.client.Initialize(context.Background(), &InitializeRequest{
		Config: &Config{
			BaseDir: cfg.BaseDir,
			CertDir: cfg.CertDir,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("plugin error: %s", resp.Error)
	}
	return nil
}

// EnrichDomainEntry enriches a domain entry with additional metadata
func (c *Client) EnrichDomainEntry(entry *model.DomainEntry) error {
	// Convert model.DomainEntry to rpc.DomainEntry
	rpcEntry := &DomainEntry{
		Domain:           entry.Domain,
		AlternativeNames: entry.AlternativeNames,
		Enabled:          entry.Enabled,
		Metadata:         make(map[string]*MetadataValue),
	}

	// Convert metadata
	for k, v := range entry.Metadata {
		if certInfo, ok := v.(*model.CertInfo); ok {
			rpcEntry.Metadata[k] = &MetadataValue{
				Value: &MetadataValue_CertInfo{
					CertInfo: &CertInfo{
						IsValid:   certInfo.IsValid,
						Issuer:    certInfo.Issuer,
						Subject:   certInfo.Subject,
						NotBefore: certInfo.NotBefore.Format(time.RFC3339),
						NotAfter:  certInfo.NotAfter.Format(time.RFC3339),
						Error:     certInfo.Error,
					},
				},
			}
		}
	}

	resp, err := c.client.EnrichDomainEntry(context.Background(), &EnrichDomainEntryRequest{
		Entry: rpcEntry,
	})
	if err != nil {
		return fmt.Errorf("failed to enrich domain entry: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("plugin error: %s", resp.Error)
	}

	// Update entry metadata from response
	for k, v := range resp.Entry.Metadata {
		switch val := v.Value.(type) {
		case *MetadataValue_CertInfo:
			notBefore, _ := time.Parse(time.RFC3339, val.CertInfo.NotBefore)
			notAfter, _ := time.Parse(time.RFC3339, val.CertInfo.NotAfter)
			entry.Metadata[k] = &model.CertInfo{
				IsValid:   val.CertInfo.IsValid,
				Issuer:    val.CertInfo.Issuer,
				Subject:   val.CertInfo.Subject,
				NotBefore: notBefore,
				NotAfter:  notAfter,
				Error:     val.CertInfo.Error,
			}
		}
	}

	return nil
}

// Close cleans up any resources used by the plugin
func (c *Client) Close() error {
	resp, err := c.client.Close(context.Background(), &CloseRequest{})
	if err != nil {
		return fmt.Errorf("failed to close plugin: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("plugin error: %s", resp.Error)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	if err := c.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill plugin process: %w", err)
	}

	return nil
}

// Serve starts a gRPC server for a plugin
func Serve(server interface{ PluginServer }) error {
	// Create listener on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	RegisterPluginServer(grpcServer, server)

	// Write address to stdout
	fmt.Println(listener.Addr().String())

	// Start server
	return grpcServer.Serve(listener)
}
