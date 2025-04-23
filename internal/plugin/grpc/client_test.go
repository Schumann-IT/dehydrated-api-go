package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockServer implements the gRPC server interface for testing
type mockServer struct {
	pb.UnimplementedPluginServer
	initializeErr   error
	getMetadataErr  error
	closeErr        error
	initializeResp  *pb.InitializeResponse
	getMetadataResp *pb.GetMetadataResponse
}

func (s *mockServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	if s.initializeErr != nil {
		return nil, s.initializeErr
	}
	return s.initializeResp, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	if s.getMetadataErr != nil {
		return nil, s.getMetadataErr
	}
	return s.getMetadataResp, nil
}

func (s *mockServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	if s.closeErr != nil {
		return nil, s.closeErr
	}
	return &pb.CloseResponse{}, nil
}

// setupMockServer creates a mock gRPC server for testing
func setupMockServer(t *testing.T) (*mockServer, string, func()) {
	// Use /tmp for Unix sockets as it's more reliable
	tmpDir := filepath.Join(os.TempDir(), "dehydrated-api-go-test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	sockPath := filepath.Join(tmpDir, fmt.Sprintf("plugin-%d.sock", time.Now().UnixNano()))

	// Remove any existing socket file
	os.Remove(sockPath)

	lis, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	mock := &mockServer{}
	pb.RegisterPluginServer(s, mock)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		s.Stop()
		lis.Close()
		os.Remove(sockPath)
		os.RemoveAll(tmpDir)
	}

	return mock, sockPath, cleanup
}

func TestConvertToStructValue(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantErr bool
	}{
		{
			name:    "nil map",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]any{},
			wantErr: false,
		},
		{
			name:    "string value",
			input:   map[string]any{"key": "value"},
			wantErr: false,
		},
		{
			name:    "number value",
			input:   map[string]any{"key": 42},
			wantErr: false,
		},
		{
			name:    "boolean value",
			input:   map[string]any{"key": true},
			wantErr: false,
		},
		{
			name:    "simple map value",
			input:   map[string]any{"key": map[string]any{"nested": "value"}},
			wantErr: false,
		},
		{
			name:    "unsupported type",
			input:   map[string]any{"key": make(chan int)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToStructValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.input == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.input), len(result))
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		name             string
		config           map[string]any
		dehydratedConfig *dehydrated.Config
		mockResponse     *pb.InitializeResponse
		mockError        error
		wantErr          bool
	}{
		{
			name:   "successful initialization",
			config: map[string]any{"test": "config"},
			dehydratedConfig: &dehydrated.Config{
				BaseDir: "/test/base",
			},
			mockResponse: &pb.InitializeResponse{},
			mockError:    nil,
			wantErr:      false,
		},
		{
			name:             "client nil error",
			config:           map[string]any{},
			dehydratedConfig: &dehydrated.Config{},
			mockResponse:     nil,
			mockError:        nil,
			wantErr:          true,
		},
		{
			name:             "plugin error",
			config:           map[string]any{},
			dehydratedConfig: &dehydrated.Config{},
			mockResponse:     nil,
			mockError:        fmt.Errorf("plugin error"),
			wantErr:          true,
		},
		{
			name:             "invalid config conversion",
			config:           map[string]any{"invalid": make(chan int)},
			dehydratedConfig: &dehydrated.Config{},
			mockResponse:     nil,
			mockError:        nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, sockPath, cleanup := setupMockServer(t)
			defer cleanup()

			mock.initializeErr = tt.mockError
			mock.initializeResp = tt.mockResponse

			var client *Client
			if tt.name == "client nil error" {
				client = &Client{client: nil}
			} else {
				// Create a client that connects to the mock server
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}
				defer conn.Close()

				client = &Client{
					client: pb.NewPluginClient(conn),
				}
			}

			// Test initialization
			err := client.Initialize(context.Background(), tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetMetadata(t *testing.T) {
	tests := []struct {
		name         string
		domain       model.DomainEntry
		mockResponse *pb.GetMetadataResponse
		mockError    error
		wantErr      bool
		wantMetadata map[string]any
	}{
		{
			name: "successful metadata retrieval",
			domain: model.DomainEntry{
				Domain:           "example.com",
				AlternativeNames: []string{"www.example.com"},
				Alias:            "example",
				Enabled:          true,
				Comment:          "test domain",
				Metadata:         map[string]any{},
			},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{},
			},
			mockError:    nil,
			wantErr:      false,
			wantMetadata: map[string]any{},
		},
		{
			name: "client nil error",
			domain: model.DomainEntry{
				Domain: "example.com",
			},
			mockResponse: nil,
			mockError:    nil,
			wantErr:      true,
			wantMetadata: nil,
		},
		{
			name: "plugin error",
			domain: model.DomainEntry{
				Domain: "example.com",
			},
			mockResponse: nil,
			mockError:    fmt.Errorf("plugin error"),
			wantErr:      true,
			wantMetadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, sockPath, cleanup := setupMockServer(t)
			defer cleanup()

			mock.getMetadataErr = tt.mockError
			mock.getMetadataResp = tt.mockResponse

			var client *Client
			if tt.name == "client nil error" {
				client = &Client{client: nil}
			} else {
				// Create a client that connects to the mock server
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}
				defer conn.Close()

				client = &Client{
					client: pb.NewPluginClient(conn),
				}
			}

			// Test metadata retrieval
			metadata, err := client.GetMetadata(context.Background(), tt.domain, &dehydrated.Config{})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantMetadata, metadata)
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		config           map[string]any
		dehydratedConfig *dehydrated.Config
		mockError        error
		wantErr          bool
	}{
		{
			name:   "successful client creation",
			config: map[string]any{"test": "config"},
			dehydratedConfig: &dehydrated.Config{
				BaseDir: "/test/base",
			},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:             "plugin error",
			config:           map[string]any{},
			dehydratedConfig: &dehydrated.Config{},
			mockError:        fmt.Errorf("plugin error"),
			wantErr:          true,
		},
		{
			name:             "invalid config conversion",
			config:           map[string]any{"invalid": make(chan int)},
			dehydratedConfig: &dehydrated.Config{},
			mockError:        nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, sockPath, cleanup := setupMockServer(t)
			defer cleanup()

			mock.initializeErr = tt.mockError

			// Create a client that connects to the mock server
			conn, err := grpc.Dial(
				"unix://"+sockPath,
				grpc.WithInsecure(),
				grpc.WithBlock(),
			)
			if err != nil {
				t.Fatalf("failed to connect to mock server: %v", err)
			}
			defer conn.Close()

			client := &Client{
				client: pb.NewPluginClient(conn),
			}

			// Test initialization
			err = client.Initialize(context.Background(), tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)

			// Clean up
			if client != nil {
				client.Close(context.Background())
			}
		})
	}
}

func TestClientLifecycle(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) (*Client, func())
		actions func(*testing.T, *Client)
		wantErr bool
	}{
		{
			name: "normal lifecycle",
			setup: func(t *testing.T) (*Client, func()) {
				_, sockPath, cleanup := setupMockServer(t)
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}

				client := &Client{
					client: pb.NewPluginClient(conn),
					conn:   conn,
				}

				return client, func() {
					conn.Close()
					cleanup()
				}
			},
			actions: func(t *testing.T, c *Client) {
				ctx := context.Background()
				err := c.Initialize(ctx, map[string]any{"test": "value"})
				assert.NoError(t, err)

				_, err = c.GetMetadata(ctx, model.DomainEntry{Domain: "test.com"}, &dehydrated.Config{})
				assert.NoError(t, err)

				err = c.Close(ctx)
				assert.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "concurrent operations",
			setup: func(t *testing.T) (*Client, func()) {
				_, sockPath, cleanup := setupMockServer(t)
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}

				client := &Client{
					client: pb.NewPluginClient(conn),
					conn:   conn,
				}

				return client, func() {
					conn.Close()
					cleanup()
				}
			},
			actions: func(t *testing.T, c *Client) {
				var wg sync.WaitGroup
				ctx := context.Background()
				numGoroutines := 10

				// Initialize once before concurrent operations
				err := c.Initialize(ctx, map[string]any{"test": "value"})
				assert.NoError(t, err)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						_, err := c.GetMetadata(ctx, model.DomainEntry{Domain: fmt.Sprintf("test%d.com", id)}, &dehydrated.Config{})
						assert.NoError(t, err)
					}(i)
				}

				wg.Wait()
			},
			wantErr: false,
		},
		{
			name: "cleanup on error",
			setup: func(t *testing.T) (*Client, func()) {
				mock, sockPath, cleanup := setupMockServer(t)
				mock.initializeErr = fmt.Errorf("initialization error")

				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}

				client := &Client{
					client: pb.NewPluginClient(conn),
					conn:   conn,
				}

				return client, func() {
					conn.Close()
					cleanup()
				}
			},
			actions: func(t *testing.T, c *Client) {
				ctx := context.Background()
				err := c.Initialize(ctx, map[string]any{})
				assert.Error(t, err)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := tt.setup(t)
			defer cleanup()

			tt.actions(t, client)
		})
	}
}

func TestMetadataEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		domain       model.DomainEntry
		mockResponse *pb.GetMetadataResponse
		mockError    error
		wantErr      bool
	}{
		{
			name:   "very large metadata",
			domain: model.DomainEntry{Domain: "test.com"},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: generateLargeMetadata(t),
			},
			wantErr: false,
		},
		{
			name:   "nested metadata",
			domain: model.DomainEntry{Domain: "test.com"},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: generateNestedMetadata(t),
			},
			wantErr: false,
		},
		{
			name:   "empty metadata",
			domain: model.DomainEntry{Domain: "test.com"},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{},
			},
			wantErr: false,
		},
		{
			name:   "nil metadata",
			domain: model.DomainEntry{Domain: "test.com"},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: nil,
			},
			wantErr: false,
		},
		{
			name: "domain with alternative names",
			domain: model.DomainEntry{
				Domain:           "test.com",
				AlternativeNames: []string{"www.test.com", "api.test.com"},
			},
			mockResponse: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{
					"domain": structpb.NewStringValue("test.com"),
					"alternatives": func() *structpb.Value {
						v, _ := structpb.NewValue([]interface{}{"www.test.com", "api.test.com"})
						return v
					}(),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, sockPath, cleanup := setupMockServer(t)
			defer cleanup()

			mock.getMetadataErr = tt.mockError
			mock.getMetadataResp = tt.mockResponse

			conn, err := grpc.Dial(
				"unix://"+sockPath,
				grpc.WithInsecure(),
				grpc.WithBlock(),
			)
			if err != nil {
				t.Fatalf("failed to connect to mock server: %v", err)
			}
			defer conn.Close()

			client := &Client{
				client: pb.NewPluginClient(conn),
			}

			metadata, err := client.GetMetadata(context.Background(), tt.domain, &dehydrated.Config{})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, metadata)
		})
	}
}

func TestContextHandling(t *testing.T) {
	tests := []struct {
		name        string
		ctxTimeout  time.Duration
		operation   func(context.Context, *Client) error
		wantTimeout bool
	}{
		{
			name:       "short timeout",
			ctxTimeout: 1 * time.Millisecond,
			operation: func(ctx context.Context, c *Client) error {
				time.Sleep(10 * time.Millisecond) // Force timeout
				return c.Initialize(ctx, map[string]any{})
			},
			wantTimeout: true,
		},
		{
			name:       "cancel during operation",
			ctxTimeout: 100 * time.Millisecond,
			operation: func(ctx context.Context, c *Client) error {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				// Cancel context after a short delay
				go func() {
					time.Sleep(1 * time.Millisecond)
					cancel()
				}()

				// Add a delay to ensure the operation is cancelled
				time.Sleep(10 * time.Millisecond)
				return c.Initialize(ctx, map[string]any{})
			},
			wantTimeout: true,
		},
		{
			name:       "normal completion",
			ctxTimeout: 1 * time.Second,
			operation: func(ctx context.Context, c *Client) error {
				return c.Initialize(ctx, map[string]any{})
			},
			wantTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sockPath, cleanup := setupMockServer(t)
			defer cleanup()

			conn, err := grpc.Dial(
				"unix://"+sockPath,
				grpc.WithInsecure(),
				grpc.WithBlock(),
			)
			if err != nil {
				t.Fatalf("failed to connect to mock server: %v", err)
			}
			defer conn.Close()

			client := &Client{
				client: pb.NewPluginClient(conn),
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			ctx = context.WithValue(ctx, "cancel", cancel)
			defer cancel()

			err = tt.operation(ctx, client)
			if tt.wantTimeout {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for generating test data
func generateLargeMetadata(t *testing.T) map[string]*structpb.Value {
	metadata := make(map[string]*structpb.Value)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value, err := structpb.NewValue(fmt.Sprintf("value%d", i))
		assert.NoError(t, err)
		metadata[key] = value
	}
	return metadata
}

func generateNestedMetadata(t *testing.T) map[string]*structpb.Value {
	nestedMap := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"string": "value",
					"number": 42,
					"bool":   true,
					"array":  []interface{}{"item1", "item2"},
				},
			},
		},
	}

	metadata := make(map[string]*structpb.Value)
	value, err := structpb.NewValue(nestedMap)
	assert.NoError(t, err)
	metadata["nested"] = value
	return metadata
}

func TestNewClientErrors(t *testing.T) {
	tests := []struct {
		name             string
		pluginPath       string
		config           map[string]any
		dehydratedConfig *dehydrated.Config
		wantErr          bool
		errorContains    string
	}{
		{
			name:             "invalid plugin path",
			pluginPath:       "/nonexistent/plugin",
			config:           map[string]any{},
			dehydratedConfig: &dehydrated.Config{},
			wantErr:          true,
			errorContains:    "failed to start plugin",
		},
		{
			name:             "invalid config",
			pluginPath:       "/nonexistent/plugin",
			config:           map[string]any{"invalid": make(chan int)},
			dehydratedConfig: &dehydrated.Config{},
			wantErr:          true,
			errorContains:    "failed to convert config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.pluginPath, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, client)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestClientErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T) (*Client, *mockServer, func())
		operation func(*Client) error
		wantErr   bool
	}{
		{
			name: "connection lost during operation",
			setup: func(t *testing.T) (*Client, *mockServer, func()) {
				mock, sockPath, cleanup := setupMockServer(t)
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}

				client := &Client{
					client: pb.NewPluginClient(conn),
					conn:   conn,
				}

				return client, mock, func() {
					conn.Close()
					cleanup()
				}
			},
			operation: func(c *Client) error {
				// Close the connection to simulate connection loss
				c.conn.Close()
				return c.Initialize(context.Background(), map[string]any{})
			},
			wantErr: true,
		},
		{
			name: "plugin returns error",
			setup: func(t *testing.T) (*Client, *mockServer, func()) {
				mock, sockPath, cleanup := setupMockServer(t)
				mock.initializeErr = fmt.Errorf("plugin error")

				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				if err != nil {
					t.Fatalf("failed to connect to mock server: %v", err)
				}

				client := &Client{
					client: pb.NewPluginClient(conn),
					conn:   conn,
				}

				return client, mock, func() {
					conn.Close()
					cleanup()
				}
			},
			operation: func(c *Client) error {
				return c.Initialize(context.Background(), map[string]any{})
			},
			wantErr: true,
		},
		{
			name: "nil client",
			setup: func(t *testing.T) (*Client, *mockServer, func()) {
				mock, _, cleanup := setupMockServer(t)
				return &Client{}, mock, cleanup
			},
			operation: func(c *Client) error {
				return c.Initialize(context.Background(), map[string]any{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _, cleanup := tt.setup(t)
			defer cleanup()

			err := tt.operation(client)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestResourceCleanup(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) (*Client, string, func())
		verify  func(*testing.T, string)
		wantErr bool
	}{
		{
			name: "cleanup temporary directory",
			setup: func(t *testing.T) (*Client, string, func()) {
				tmpDir, err := os.MkdirTemp("", "plugin-test-*")
				assert.NoError(t, err)

				_, sockPath, cleanup := setupMockServer(t)
				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				assert.NoError(t, err)

				client := &Client{
					client:     pb.NewPluginClient(conn),
					conn:       conn,
					tmpDir:     tmpDir,
					socketFile: filepath.Join(tmpDir, "plugin.sock"),
				}

				return client, tmpDir, func() {
					conn.Close()
					cleanup()
				}
			},
			verify: func(t *testing.T, tmpDir string) {
				// Verify directory is removed after cleanup
				_, err := os.Stat(tmpDir)
				assert.True(t, os.IsNotExist(err))
			},
			wantErr: false,
		},
		{
			name: "cleanup on initialization error",
			setup: func(t *testing.T) (*Client, string, func()) {
				tmpDir, err := os.MkdirTemp("", "plugin-test-*")
				assert.NoError(t, err)

				mock, sockPath, cleanup := setupMockServer(t)
				mock.initializeErr = fmt.Errorf("initialization error")

				conn, err := grpc.Dial(
					"unix://"+sockPath,
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				assert.NoError(t, err)

				client := &Client{
					client:     pb.NewPluginClient(conn),
					conn:       conn,
					tmpDir:     tmpDir,
					socketFile: filepath.Join(tmpDir, "plugin.sock"),
					lastError:  fmt.Errorf("initialization error"),
				}

				return client, tmpDir, func() {
					conn.Close()
					cleanup()
				}
			},
			verify: func(t *testing.T, tmpDir string) {
				// Verify directory is removed after cleanup
				_, err := os.Stat(tmpDir)
				assert.True(t, os.IsNotExist(err))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, tmpDir, cleanup := tt.setup(t)
			defer cleanup()

			ctx := context.Background()
			err := client.Close(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.verify(t, tmpDir)
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	mock, sockPath, cleanup := setupMockServer(t)
	defer cleanup()

	// Configure mock server responses
	mock.initializeResp = &pb.InitializeResponse{}
	mock.getMetadataResp = &pb.GetMetadataResponse{
		Metadata: map[string]*structpb.Value{
			"test": structpb.NewStringValue("value"),
		},
	}

	conn, err := grpc.Dial(
		"unix://"+sockPath,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("failed to connect to mock server at %s: %v", sockPath, err)
	}
	defer conn.Close()

	client := &Client{
		client: pb.NewPluginClient(conn),
		conn:   conn,
	}

	// Test concurrent initialization
	var wg sync.WaitGroup
	numGoroutines := 10
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			config := map[string]any{
				"id": id,
			}
			if err := client.Initialize(ctx, config); err != nil {
				errors <- err
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}

	// Test concurrent metadata retrieval
	wg = sync.WaitGroup{}
	metadataErrors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			domain := model.DomainEntry{
				Domain: fmt.Sprintf("test%d.com", id),
			}
			if _, err := client.GetMetadata(ctx, domain, &dehydrated.Config{}); err != nil {
				metadataErrors <- err
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(metadataErrors)

	// Check for any errors
	for err := range metadataErrors {
		t.Errorf("concurrent metadata error: %v", err)
	}
}
