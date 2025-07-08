package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegistry(t *testing.T) {
	// Build the example plugin
	pluginPath := filepath.Join("..", "..", "..", "examples", "plugins", "simple", "simple")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Example plugin not built, skipping test")
	}

	// Create a test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test configuration
	cfg := map[string]config.PluginConfig{
		"simple": {
			Enabled: true,
			Registry: &config.RegistryConfig{
				Type: config.PluginSourceTypeLocal,
				Config: map[string]any{
					"path": pluginPath,
				},
			},
			Config: map[string]any{
				"name": "example",
			},
		},
	}

	// Create logger for testing
	logger := zap.NewNop()

	r := New("", cfg, logger)
	defer r.Close()

	// Test that plugins are available
	plugins := r.Plugins()
	require.NotNil(t, plugins)
	require.Contains(t, plugins, "simple")

	// Test plugin functionality
	plugin := plugins["simple"]
	require.NotNil(t, plugin)

	// Test GetMetadata call
	domainEntry := &pb.DomainEntry{
		Domain:           "example.com",
		AlternativeNames: []string{"www.example.com"},
		Alias:            "example",
		Enabled:          true,
		Comment:          "Test domain",
	}

	resp, err := plugin.GetMetadata(ctx, &pb.GetMetadataRequest{
		DomainEntry: domainEntry,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Metadata)

	cache.Clean()
}
