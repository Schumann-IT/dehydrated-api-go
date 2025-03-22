package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"

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
			err := client.Initialize(context.Background(), tt.config, tt.dehydratedConfig)
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
		domain       string
		mockResponse *pb.GetMetadataResponse
		mockError    error
		wantErr      bool
		wantMetadata map[string]any
	}{
		{
			name:   "successful metadata retrieval",
			domain: "example.com",
			mockResponse: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{
					"test": structpb.NewStringValue("value"),
				},
			},
			mockError:    nil,
			wantErr:      false,
			wantMetadata: map[string]any{"test": "value"},
		},
		{
			name:         "client nil error",
			domain:       "example.com",
			mockResponse: nil,
			mockError:    nil,
			wantErr:      true,
			wantMetadata: nil,
		},
		{
			name:         "plugin error",
			domain:       "error.com",
			mockResponse: nil,
			mockError:    fmt.Errorf("plugin error"),
			wantErr:      true,
			wantMetadata: nil,
		},
		{
			name:   "complex metadata types",
			domain: "example.com",
			mockResponse: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{
					"string": structpb.NewStringValue("value"),
					"number": structpb.NewNumberValue(42),
					"bool":   structpb.NewBoolValue(true),
					"struct": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"name": structpb.NewStringValue("test"),
							"age":  structpb.NewNumberValue(42),
						},
					}),
				},
			},
			mockError: nil,
			wantErr:   false,
			wantMetadata: map[string]any{
				"string": "value",
				"number": 42.0,
				"bool":   true,
				"struct": map[string]interface{}{
					"name": "test",
					"age":  42.0,
				},
			},
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
			metadata, err := client.GetMetadata(context.Background(), tt.domain)
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
			err = client.Initialize(context.Background(), tt.config, tt.dehydratedConfig)
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
