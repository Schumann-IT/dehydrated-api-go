package grpc

import (
	"context"
	"os"
	"os/exec"
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
}

func (s *mockServer) Initialize(ctx context.Context, req *plugin.InitializeRequest) (*plugin.InitializeResponse, error) {
	return &plugin.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *plugin.GetMetadataRequest) (*plugin.GetMetadataResponse, error) {
	return &plugin.GetMetadataResponse{
		Metadata: map[string]string{
			"test": "value",
		},
	}, nil
}

func (s *mockServer) Close(ctx context.Context, req *plugin.CloseRequest) (*plugin.CloseResponse, error) {
	return &plugin.CloseResponse{}, nil
}

func TestClient(t *testing.T) {
	// Create a temporary directory for the socket
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sockFile := filepath.Join(tmpDir, "plugin.sock")

	// Start the test plugin
	cmd := exec.Command("go", "run", "testdata/test-plugin/main.go")
	cmd.Env = append(os.Environ(), "PLUGIN_SOCKET="+sockFile)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Wait for the socket file to be created
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(sockFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Create a gRPC connection
	conn, err := grpc.Dial(
		"unix://"+sockFile,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := plugin.NewPluginClient(conn)

	// Test Initialize
	_, err = client.Initialize(context.Background(), &plugin.InitializeRequest{
		Config: map[string]string{
			"test": "config",
		},
	})
	assert.NoError(t, err)

	// Test GetMetadata
	resp, err := client.GetMetadata(context.Background(), &plugin.GetMetadataRequest{
		Domain: "example.com",
		Config: map[string]string{
			"test": "config",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"test": "value",
	}, resp.Metadata)

	// Test Close
	_, err = client.Close(context.Background(), &plugin.CloseRequest{})
	assert.NoError(t, err)
}

func TestClientErrors(t *testing.T) {
	// Test with non-existent plugin
	_, err := NewClient("/non/existent/plugin", map[string]string{})
	assert.Error(t, err)

	// Test with invalid socket
	client := &Client{
		conn:     nil,
		client:   plugin.NewPluginClient(nil),
		sockFile: "/non/existent/socket",
	}
	err = client.Close(context.Background())
	assert.Error(t, err)
}
