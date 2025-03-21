package grpc

import (
	"context"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/config"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/proto/plugin"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Manager handles gRPC plugin communication
type Manager struct {
	plugins map[string]*PluginClient
	mu      sync.RWMutex
}

// PluginClient represents a connected gRPC plugin
type PluginClient struct {
	client plugin.PluginClient
	conn   *grpc.ClientConn
	cmd    *exec.Cmd
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]*PluginClient),
	}
}

// LoadPlugin loads and initializes a plugin
func (m *Manager) LoadPlugin(name, path string, cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already loaded", name)
	}

	// Start the plugin process
	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// Create gRPC connection
	conn, err := grpc.Dial("pipe",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return &pipeConn{
				stdin:  stdin,
				stdout: stdout,
			}, nil
		}),
	)
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client := &PluginClient{
		client: plugin.NewPluginClient(conn),
		conn:   conn,
		cmd:    cmd,
	}

	// Initialize the plugin
	ctx := context.Background()
	initReq := &plugin.InitializeRequest{
		CertDir: cfg.CertDir,
		BaseDir: cfg.BaseDir,
		Config:  make(map[string]string),
	}

	initResp, err := client.client.Initialize(ctx, initReq)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	if !initResp.Success {
		client.Close()
		return fmt.Errorf("plugin initialization failed: %s", initResp.Error)
	}

	m.plugins[name] = client
	return nil
}

// EnrichDomainEntry enriches a domain entry using all loaded plugins
func (m *Manager) EnrichDomainEntry(entry *model.DomainEntry) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, client := range m.plugins {
		ctx := context.Background()
		req := &plugin.EnrichDomainEntryRequest{
			Entry: &plugin.DomainEntry{
				Domain:           entry.Domain,
				AlternativeNames: entry.AlternativeNames,
				Enabled:          entry.Enabled,
				Metadata:         make(map[string]string),
			},
		}

		resp, err := client.client.EnrichDomainEntry(ctx, req)
		if err != nil {
			return fmt.Errorf("plugin %s failed: %w", name, err)
		}

		if !resp.Success {
			return fmt.Errorf("plugin %s failed: %s", name, resp.Error)
		}

		// Update entry with enriched data
		entry.AlternativeNames = resp.Entry.AlternativeNames
		entry.Enabled = resp.Entry.Enabled
		for k, v := range resp.Entry.Metadata {
			if entry.Metadata == nil {
				entry.Metadata = make(map[string]interface{})
			}
			entry.Metadata[k] = v
		}
	}

	return nil
}

// Close closes all plugin connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.plugins {
		if err := client.Close(); err != nil {
			return fmt.Errorf("failed to close plugin %s: %w", name, err)
		}
	}

	m.plugins = make(map[string]*PluginClient)
	return nil
}

// Close closes the plugin client connection
func (c *PluginClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// pipeConn implements net.Conn for plugin communication
type pipeConn struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (c *pipeConn) Read(b []byte) (n int, err error)  { return c.stdout.Read(b) }
func (c *pipeConn) Write(b []byte) (n int, err error) { return c.stdin.Write(b) }
func (c *pipeConn) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	return nil
}
func (c *pipeConn) LocalAddr() net.Addr                { return nil }
func (c *pipeConn) RemoteAddr() net.Addr               { return nil }
func (c *pipeConn) SetDeadline(t time.Time) error      { return nil }
func (c *pipeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *pipeConn) SetWriteDeadline(t time.Time) error { return nil }
