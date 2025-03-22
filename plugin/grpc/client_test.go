package grpc

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"
	plugininterface "github.com/schumann-it/dehydrated-api-go/plugin/interface"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockServer struct {
	pb.UnimplementedPluginServer
	initializeErr  error
	getMetadataErr error
	closeErr       error
}

func (s *mockServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	if s.initializeErr != nil {
		return nil, s.initializeErr
	}
	return &pb.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	if s.getMetadataErr != nil {
		return nil, s.getMetadataErr
	}
	metadata := make(map[string]*structpb.Value)
	strValue, _ := structpb.NewValue("value")
	metadata["test"] = strValue
	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (s *mockServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	if s.closeErr != nil {
		return nil, s.closeErr
	}
	return &pb.CloseResponse{}, nil
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
	pb.RegisterPluginServer(s, &mockServer{})

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

	// Create a test dehydrated config
	dehydratedConfig := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		AccountsDir:   "/test/accounts",
		ChallengesDir: "/test/challenges",
		DomainsFile:   "/test/domains.txt",
	}

	tests := []struct {
		name             string
		pluginPath       string
		config           map[string]any
		dehydratedConfig *dehydrated.Config
		wantErr          bool
		errContains      string
	}{
		{
			name:             "successful client creation",
			pluginPath:       mockPluginPath,
			config:           map[string]any{"test": "config"},
			dehydratedConfig: dehydratedConfig,
			wantErr:          false,
		},
		{
			name:             "non-existent plugin",
			pluginPath:       "/non/existent/plugin",
			config:           map[string]any{},
			dehydratedConfig: dehydratedConfig,
			wantErr:          true,
			errContains:      "failed to start plugin",
		},
		{
			name:             "invalid plugin",
			pluginPath:       "client_test.go", // Use this file as an invalid plugin
			config:           map[string]any{},
			dehydratedConfig: dehydratedConfig,
			wantErr:          true,
			errContains:      "failed to start plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.pluginPath, tt.config, tt.dehydratedConfig)
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

	ctx := context.Background()

	// Create a test dehydrated config
	dehydratedConfig := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		AccountsDir:   "/test/accounts",
		ChallengesDir: "/test/challenges",
		DomainsFile:   "/test/domains.txt",
	}

	// Create a client
	client, err := NewClient(mockPluginPath, map[string]any{}, dehydratedConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close(ctx)

	t.Run("Initialize", func(t *testing.T) {
		err := client.Initialize(ctx, map[string]any{"test": "config"}, dehydratedConfig)
		assert.NoError(t, err)
	})

	t.Run("GetMetadata", func(t *testing.T) {
		metadata, err := client.GetMetadata(ctx, "example.com")
		if err != nil {
			t.Errorf("GetMetadata failed: %v", err)
		}
		if metadata["test"] != "value" {
			t.Errorf("Expected metadata test=value, got %v", metadata)
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := client.Close(ctx)
		assert.NoError(t, err)
	})

	t.Run("GetMetadata Error", func(t *testing.T) {
		mockPluginPath := "testdata/mock-plugin/mock-plugin"
		if _, err := os.Stat(mockPluginPath); os.IsNotExist(err) {
			t.Skip("mock plugin not built, run 'go build -o mock-plugin' in testdata/mock-plugin directory")
		}

		client, err := NewClient(mockPluginPath, map[string]any{}, dehydratedConfig)
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close(ctx)

		_, err = client.GetMetadata(ctx, "error.com")
		assert.Error(t, err)
		assert.Equal(t, plugininterface.ErrPluginError, err)
	})
}

func TestClientErrors(t *testing.T) {
	ctx := context.Background()

	// Create a test dehydrated config
	dehydratedConfig := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		AccountsDir:   "/test/accounts",
		ChallengesDir: "/test/challenges",
		DomainsFile:   "/test/domains.txt",
	}

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
				c.client = pb.NewPluginClient(nil)
			},
			operation: func(c *Client) error {
				return c.Close(ctx)
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
				return c.Close(ctx)
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
				return c.Initialize(ctx, map[string]any{}, dehydratedConfig)
			},
			wantErr:     true,
			errContains: "client is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			tt.setup(client)

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

	ctx := context.Background()

	// Create a test dehydrated config
	dehydratedConfig := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		AccountsDir:   "/test/accounts",
		ChallengesDir: "/test/challenges",
		DomainsFile:   "/test/domains.txt",
	}

	// Create a client
	client, err := NewClient(mockPluginPath, map[string]any{}, dehydratedConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close(ctx)

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			metadata, err := client.GetMetadata(ctx, "example.com")
			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"test": "value"}, metadata)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClientEdgeCases(t *testing.T) {
	// Save original TMPDIR
	origTmpDir := os.Getenv("TMPDIR")
	defer os.Setenv("TMPDIR", origTmpDir)

	ctx := context.Background()

	// Create a test dehydrated config
	dehydratedConfig := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		AccountsDir:   "/test/accounts",
		ChallengesDir: "/test/challenges",
		DomainsFile:   "/test/domains.txt",
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		config      map[string]any
		wantErr     bool
		errContains string
	}{
		{
			name: "socket file creation failure",
			setup: func(t *testing.T) string {
				// Create a read-only directory
				tmpDir := t.TempDir()
				if err := os.Chmod(tmpDir, 0o444); err != nil {
					t.Fatal(err)
				}
				return filepath.Join(tmpDir, "plugin.sock")
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to start plugin: fork/exec",
		},
		{
			name: "plugin process startup failure",
			setup: func(t *testing.T) string {
				// Create a temporary directory
				tmpDir, err := os.MkdirTemp("", "plugin-test-*")
				if err != nil {
					t.Fatal(err)
				}
				// Create a non-executable file
				sockFile := filepath.Join(tmpDir, "plugin.sock")
				if err := os.WriteFile(sockFile, []byte("not executable"), 0644); err != nil {
					t.Fatal(err)
				}
				return sockFile
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to start plugin",
		},
		{
			name: "plugin initialization failure",
			setup: func(t *testing.T) string {
				// Create a temporary directory
				tmpDir, err := os.MkdirTemp("", "plugin-test-*")
				if err != nil {
					t.Fatal(err)
				}
				// Create a mock plugin that fails to initialize
				sockFile := filepath.Join(tmpDir, "plugin.sock")
				pluginContent := `package main

import (
	"context"
	"fmt"
	"net"
	"os"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

type mockServer struct {
	pb.UnimplementedPluginServer
}

func (s *mockServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return nil, fmt.Errorf("initialization failed")
}

func (s *mockServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	return &pb.GetMetadataResponse{}, nil
}

func (s *mockServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}

func main() {
	sockFile := os.Getenv("PLUGIN_SOCKET")
	if sockFile == "" {
		fmt.Fprintln(os.Stderr, "PLUGIN_SOCKET environment variable not set")
		os.Exit(1)
	}

	lis, err := net.Listen("unix", sockFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterPluginServer(s, &mockServer{})

	if err := s.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}`
				pluginFile := filepath.Join(tmpDir, "plugin.go")
				if err := os.WriteFile(pluginFile, []byte(pluginContent), 0644); err != nil {
					t.Fatal(err)
				}
				// Build the plugin
				cmd := exec.Command("go", "build", "-o", sockFile, pluginFile)
				if err := cmd.Run(); err != nil {
					t.Fatal(err)
				}
				return sockFile
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to initialize plugin",
		},
		{
			name: "socket file timeout",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a mock plugin that doesn't create the socket file
				pluginPath := filepath.Join(tmpDir, "timeout-plugin")
				script := `#!/bin/sh
sleep 1
exit 0
`
				if err := os.WriteFile(pluginPath, []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
				return pluginPath
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to initialize plugin: rpc error: code = Unavailable desc = connection error",
		},
		{
			name: "temp dir creation failure",
			setup: func(t *testing.T) string {
				// Set TMPDIR to a non-existent directory
				nonExistentDir := "/non/existent/dir"
				os.Setenv("TMPDIR", nonExistentDir)
				return "testdata/mock-plugin/mock-plugin"
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to create temp dir",
		},
		{
			name: "grpc connection failure",
			setup: func(t *testing.T) string {
				// Reset TMPDIR to original value
				os.Setenv("TMPDIR", origTmpDir)
				// Create a plugin that exits immediately
				tmpDir := t.TempDir()
				pluginPath := filepath.Join(tmpDir, "exit-plugin")
				script := `#!/bin/sh
exit 0
`
				if err := os.WriteFile(pluginPath, []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
				return pluginPath
			},
			config:      map[string]any{},
			wantErr:     true,
			errContains: "failed to initialize plugin: rpc error: code = Unavailable desc = connection error",
		},
		{
			name: "get metadata error",
			setup: func(t *testing.T) string {
				// Reset TMPDIR to original value
				os.Setenv("TMPDIR", origTmpDir)
				// Create a plugin that returns an error for GetMetadata
				tmpDir := t.TempDir()
				pluginPath := filepath.Join(tmpDir, "error-plugin")
				pluginContent := `package main

import (
	"context"
	"fmt"
	"net"
	"os"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

type mockServer struct {
	pb.UnimplementedPluginServer
}

func (s *mockServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return &pb.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	return nil, fmt.Errorf("metadata error")
}

func (s *mockServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}

func main() {
	sockFile := os.Getenv("PLUGIN_SOCKET")
	if sockFile == "" {
		fmt.Fprintln(os.Stderr, "PLUGIN_SOCKET environment variable not set")
		os.Exit(1)
	}

	lis, err := net.Listen("unix", sockFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterPluginServer(s, &mockServer{})

	if err := s.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}`
				pluginFile := filepath.Join(tmpDir, "plugin.go")
				if err := os.WriteFile(pluginFile, []byte(pluginContent), 0644); err != nil {
					t.Fatal(err)
				}
				// Build the plugin
				cmd := exec.Command("go", "build", "-o", pluginPath, pluginFile)
				if err := cmd.Run(); err != nil {
					t.Fatal(err)
				}
				return pluginPath
			},
			config:      map[string]any{},
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginPath := tt.setup(t)
			client, err := NewClient(pluginPath, tt.config, dehydratedConfig)
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
				if tt.name == "get metadata error" {
					metadata, err := client.GetMetadata(ctx, "example.com")
					if err != plugininterface.ErrPluginError {
						t.Errorf("Expected GetMetadata error, got %v", err)
					} else if metadata != nil {
						t.Error("Expected nil metadata on error")
					}
				}
				client.Close(ctx)
			}
		})
	}
}
