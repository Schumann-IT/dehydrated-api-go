package grpc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mockServer struct {
	plugin.UnimplementedPluginServer
	initializeErr  error
	getMetadataErr error
	closeErr       error
}

func (s *mockServer) Initialize(ctx context.Context, req *plugin.InitializeRequest) (*plugin.InitializeResponse, error) {
	if s.initializeErr != nil {
		return nil, s.initializeErr
	}
	return &plugin.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *plugin.GetMetadataRequest) (*plugin.GetMetadataResponse, error) {
	if s.getMetadataErr != nil {
		return nil, s.getMetadataErr
	}
	return &plugin.GetMetadataResponse{
		Metadata: map[string]string{
			"test": "value",
		},
	}, nil
}

func (s *mockServer) Close(ctx context.Context, req *plugin.CloseRequest) (*plugin.CloseResponse, error) {
	if s.closeErr != nil {
		return nil, s.closeErr
	}
	return &plugin.CloseResponse{}, nil
}

func setupTestServer(t *testing.T) (*grpc.Server, string, func()) {
	// Create a temporary directory for the socket
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	if err != nil {
		t.Fatal(err)
	}

	sockFile := filepath.Join(tmpDir, "plugin.sock")

	// Create a Unix domain socket listener
	lis, err := net.Listen("unix", sockFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Create a gRPC server
	s := grpc.NewServer()
	plugin.RegisterPluginServer(s, &mockServer{})

	// Start the server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for the socket file to be created
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(sockFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	cleanup := func() {
		s.Stop()
		lis.Close()
		os.RemoveAll(tmpDir)
	}

	return s, sockFile, cleanup
}

func TestNewClient(t *testing.T) {
	mockPluginPath := "testdata/mock-plugin/mock-plugin"
	if _, err := os.Stat(mockPluginPath); os.IsNotExist(err) {
		t.Skip("mock plugin not built, run 'go build -o mock-plugin' in testdata/mock-plugin directory")
	}

	tests := []struct {
		name        string
		pluginPath  string
		config      map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful client creation",
			pluginPath: mockPluginPath,
			config:     map[string]string{"test": "config"},
			wantErr:    false,
		},
		{
			name:        "non-existent plugin",
			pluginPath:  "/non/existent/plugin",
			config:      map[string]string{},
			wantErr:     true,
			errContains: "failed to start plugin",
		},
		{
			name:        "invalid plugin",
			pluginPath:  "client_test.go", // Use this file as an invalid plugin
			config:      map[string]string{},
			wantErr:     true,
			errContains: "failed to start plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.pluginPath, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
			if client != nil {
				client.Close(context.Background())
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	mockPluginPath := "testdata/mock-plugin/mock-plugin"
	if _, err := os.Stat(mockPluginPath); os.IsNotExist(err) {
		t.Skip("mock plugin not built, run 'go build -o mock-plugin' in testdata/mock-plugin directory")
	}

	// Create a client
	client, err := NewClient(mockPluginPath, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close(context.Background())

	t.Run("Initialize", func(t *testing.T) {
		err := client.Initialize(map[string]string{"test": "config"})
		assert.NoError(t, err)
	})

	t.Run("GetMetadata", func(t *testing.T) {
		metadata, err := client.GetMetadata("example.com", map[string]string{"test": "config"})
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"test": "value"}, metadata)
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close(context.Background())
		assert.NoError(t, err)
	})
}

func TestClientErrors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Client)
		operation   func(*Client) error
		wantErr     bool
		errContains string
	}{
		{
			name: "close with nil connection",
			setup: func(c *Client) {
				c.conn = nil
				c.client = plugin.NewPluginClient(nil)
			},
			operation: func(c *Client) error {
				return c.Close(context.Background())
			},
			wantErr:     true,
			errContains: "connection is nil",
		},
		{
			name: "close with nil client",
			setup: func(c *Client) {
				conn, _ := grpc.Dial("unix:///non-existent", grpc.WithTransportCredentials(insecure.NewCredentials()))
				c.conn = conn
				c.client = nil
			},
			operation: func(c *Client) error {
				return c.Close(context.Background())
			},
			wantErr:     true,
			errContains: "client is nil",
		},
		{
			name: "initialize with nil client",
			setup: func(c *Client) {
				c.client = nil
			},
			operation: func(c *Client) error {
				return c.Initialize(map[string]string{})
			},
			wantErr:     true,
			errContains: "client is nil",
		},
		{
			name: "get metadata with nil client",
			setup: func(c *Client) {
				c.client = nil
			},
			operation: func(c *Client) error {
				_, err := c.GetMetadata("example.com", map[string]string{})
				return err
			},
			wantErr:     true,
			errContains: "client is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			if tt.setup != nil {
				tt.setup(client)
			}
			err := tt.operation(client)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestClientConcurrency(t *testing.T) {
	mockPluginPath := "testdata/mock-plugin/mock-plugin"
	if _, err := os.Stat(mockPluginPath); os.IsNotExist(err) {
		t.Skip("mock plugin not built, run 'go build -o mock-plugin' in testdata/mock-plugin directory")
	}

	// Create a client
	client, err := NewClient(mockPluginPath, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close(context.Background())

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			metadata, err := client.GetMetadata("example.com", map[string]string{"test": "config"})
			assert.NoError(t, err)
			assert.Equal(t, map[string]string{"test": "value"}, metadata)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
