package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	ctx := context.Background()

	pc := plugin.PluginConfig{
		Enabled: true,
		Path:    mustGetPluginPath(t),
		Config:  map[string]any{"key": "value"},
	}
	registry, err := NewRegistry(map[string]plugin.PluginConfig{
		"test": pc,
	})
	if err != nil {
		t.Errorf("LoadPlugin failed: %v", err)
	}

	// Test getting a plugin
	p, err := registry.GetPlugin("test")
	if err != nil {
		t.Errorf("GetPlugin failed: %v", err)
	}
	if p == nil {
		t.Error("GetPlugin returned nil")
	}

	// Test getting all plugins
	plugins := registry.GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	// Test loading duplicate plugin
	err = registry.LoadPlugin("test", pc)
	if err == nil {
		t.Error("Expected error loading duplicate plugin")
	}

	// Test getting non-existent plugin
	_, err = registry.GetPlugin("nonexistent")
	if err == nil {
		t.Error("Expected error getting non-existent plugin")
	}

	// Test closing plugins
	err = registry.Close(ctx)
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify plugins are cleared
	plugins = registry.GetPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}

func TestRegistryConcurrency(t *testing.T) {
	ctx := context.Background()

	pc := plugin.PluginConfig{
		Enabled: true,
		Path:    mustGetPluginPath(t),
		Config:  map[string]any{"key": "value"},
	}
	registry, err := NewRegistry(map[string]plugin.PluginConfig{})
	require.NoError(t, err)

	// Test concurrent plugin loading
	t.Run("ConcurrentLoad", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				err := registry.LoadPlugin(
					fmt.Sprintf("plugin%d", i),
					pc,
				)
				if err != nil {
					t.Errorf("Failed to load plugin %d: %v", i, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		plugins := registry.GetPlugins()
		if len(plugins) != 10 {
			t.Errorf("Expected 10 plugins, got %d", len(plugins))
		}
	})

	// Test concurrent plugin access
	t.Run("ConcurrentAccess", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				p, err := registry.GetPlugin(fmt.Sprintf("plugin%d", i))
				if err != nil {
					t.Errorf("Failed to get plugin %d: %v", i, err)
				}
				if p == nil {
					t.Errorf("Plugin %d is nil", i)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Test concurrent close
	t.Run("ConcurrentClose", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				err := registry.Close(ctx)
				if err != nil {
					t.Errorf("Failed to close registry: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		plugins := registry.GetPlugins()
		if len(plugins) != 0 {
			t.Errorf("Expected 0 plugins, got %d", len(plugins))
		}
	})
}

func TestLoadBuiltinPlugin(t *testing.T) {
	// Create test Config
	cfg := &dehydrated.Config{
		DehydratedConfig: pb.DehydratedConfig{
			BaseDir:       "/test/base",
			CertDir:       "/test/certs",
			DomainsDir:    "/test/domains",
			ChallengeType: "dns-01",
			Ca:            "https://acme-v02.api.letsencrypt.org/directory",
		},
	}

	tests := []struct {
		name         string
		pluginName   string
		pluginConfig plugin.PluginConfig
		wantErr      bool
		errContains  string
	}{
		{
			name:       "load openssl plugin successfully",
			pluginName: "openssl",
			pluginConfig: plugin.PluginConfig{
				Enabled: true,
				Config: map[string]any{
					"cert": true,
				},
			},
			wantErr: false,
		},
		{
			name:       "load non-existent built-in plugin",
			pluginName: "non-existent",
			pluginConfig: plugin.PluginConfig{
				Enabled: true,
				Config:  map[string]any{},
			},
			wantErr:     true,
			errContains: "built-in plugin non-existent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new registry
			reg, err := NewRegistry(map[string]plugin.PluginConfig{
				tt.pluginName: tt.pluginConfig,
			})
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)

			// Verify p is loaded
			p, err := reg.GetPlugin(tt.pluginName)
			require.NoError(t, err)
			assert.NotNil(t, p)

			// Test p functionality
			ctx := context.Background()
			err = p.Initialize(ctx, tt.pluginConfig.Config)
			require.NoError(t, err)

			// Test GetMetadata
			metadata, err := p.GetMetadata(ctx, &model.DomainEntry{DomainEntry: pb.DomainEntry{Domain: "example.com"}}, cfg)
			require.NoError(t, err)
			assert.NotNil(t, metadata)

			// Test Close
			err = p.Close(ctx)
			require.NoError(t, err)
		})
	}
}

func TestLoadPluginTwice(t *testing.T) {
	reg, err := NewRegistry(map[string]plugin.PluginConfig{
		"openssl": {
			Enabled: true,
			Config:  map[string]any{},
		},
	})
	require.NoError(t, err)

	// Try to load same plugin again
	err = reg.LoadPlugin("openssl", plugin.PluginConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin openssl is already loaded")
}

func TestGetNonExistentPlugin(t *testing.T) {
	reg, err := NewRegistry(map[string]plugin.PluginConfig{})
	require.NoError(t, err)

	// Try to get non-existent p
	p, err := reg.GetPlugin("non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin non-existent not found")
	assert.Nil(t, p)
}

func TestCloseRegistry(t *testing.T) {
	reg, err := NewRegistry(map[string]plugin.PluginConfig{
		"openssl": {
			Enabled: true,
			Config:  map[string]any{},
		},
	})
	require.NoError(t, err)

	// Close registry
	ctx := context.Background()
	err = reg.Close(ctx)
	require.NoError(t, err)

	// Verify p is removed
	p, err := reg.GetPlugin("openssl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin openssl not found")
	assert.Nil(t, p)
}

func mustGetPluginPath(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	p := filepath.Join(dir, "..", "grpc", "testdata", "test-plugin", "test-plugin")

	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("Failed to get abs path for %s: %v", p, err)
	}

	return abs
}
