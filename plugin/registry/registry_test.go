package registry

import (
	"context"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated/model"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestPlugin(t *testing.T) string {
	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Build the test plugin
	pluginPath := filepath.Join(dir, "testdata", "test-plugin")
	cmd := exec.Command("go", "build", "-o", "test-plugin", "main.go")
	cmd.Dir = pluginPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test plugin: %v", err)
	}

	return filepath.Join(pluginPath, "test-plugin")
}

func TestRegistry(t *testing.T) {
	ctx := context.Background()
	registry := NewRegistry(&dehydrated.Config{})

	// Build the test plugin
	pluginPath := buildTestPlugin(t)
	defer os.Remove(pluginPath)

	// Test loading a plugin
	err := registry.LoadPlugin("test", pluginPath, map[string]any{"key": "value"})
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
	err = registry.LoadPlugin("test", pluginPath, map[string]any{"key": "value"})
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
	registry := NewRegistry(&dehydrated.Config{})

	// Build the test plugin
	pluginPath := buildTestPlugin(t)
	defer os.Remove(pluginPath)

	// Test concurrent plugin loading
	t.Run("ConcurrentLoad", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				err := registry.LoadPlugin(
					fmt.Sprintf("plugin%d", i),
					pluginPath,
					map[string]any{"key": "value"},
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
	// Create test config
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		CA:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	tests := []struct {
		name        string
		pluginName  string
		config      map[string]any
		wantErr     bool
		errContains string
	}{
		{
			name:       "load timestamp plugin successfully",
			pluginName: "timestamp",
			config: map[string]any{
				"time_format": "2006-01-02 15:04:05",
			},
			wantErr: false,
		},
		{
			name:        "load non-existent built-in plugin",
			pluginName:  "non-existent",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "built-in plugin non-existent not found",
		},
		{
			name:       "load timestamp plugin with default config",
			pluginName: "timestamp",
			config:     map[string]any{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new registry
			reg := NewRegistry(cfg)

			// Try to load the plugin
			err := reg.LoadPlugin(tt.pluginName, "", tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)

			// Verify plugin is loaded
			plugin, err := reg.GetPlugin(tt.pluginName)
			require.NoError(t, err)
			assert.NotNil(t, plugin)

			// Test plugin functionality
			ctx := context.Background()
			err = plugin.Initialize(ctx, tt.config, cfg)
			require.NoError(t, err)

			// Test GetMetadata
			metadata, err := plugin.GetMetadata(ctx, model.DomainEntry{Domain: "example.com"})
			require.NoError(t, err)
			assert.NotNil(t, metadata)
			assert.Contains(t, metadata, "timestamp")
			assert.Contains(t, metadata, "domain")

			// Test Close
			err = plugin.Close(ctx)
			require.NoError(t, err)
		})
	}
}

func TestLoadPluginTwice(t *testing.T) {
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		CA:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg := NewRegistry(cfg)

	// Load plugin first time
	err := reg.LoadPlugin("timestamp", "", map[string]any{})
	require.NoError(t, err)

	// Try to load same plugin again
	err = reg.LoadPlugin("timestamp", "", map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin timestamp is already loaded")
}

func TestGetNonExistentPlugin(t *testing.T) {
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		CA:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg := NewRegistry(cfg)

	// Try to get non-existent plugin
	plugin, err := reg.GetPlugin("non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin non-existent not found")
	assert.Nil(t, plugin)
}

func TestCloseRegistry(t *testing.T) {
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		CA:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg := NewRegistry(cfg)

	// Load a plugin
	err := reg.LoadPlugin("timestamp", "", map[string]any{})
	require.NoError(t, err)

	// Close registry
	ctx := context.Background()
	err = reg.Close(ctx)
	require.NoError(t, err)

	// Verify plugin is removed
	plugin, err := reg.GetPlugin("timestamp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin timestamp not found")
	assert.Nil(t, plugin)
}
