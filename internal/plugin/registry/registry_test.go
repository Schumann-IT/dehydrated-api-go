package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"

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

	// Build the test plugin
	pluginPath := buildTestPlugin(t)
	defer os.Remove(pluginPath)

	pc := internal.PluginConfig{
		Enabled: true,
		Path:    pluginPath,
		Config:  map[string]any{"key": "value"},
	}
	registry, err := NewRegistry(map[string]internal.PluginConfig{
		"test": pc,
	}, &dehydrated.Config{})
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

	// Build the test plugin
	pluginPath := buildTestPlugin(t)
	defer os.Remove(pluginPath)

	pc := internal.PluginConfig{
		Enabled: true,
		Path:    pluginPath,
		Config:  map[string]any{"key": "value"},
	}
	registry, err := NewRegistry(map[string]internal.PluginConfig{}, &dehydrated.Config{})
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
	// Create test config
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		Ca:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	tests := []struct {
		name         string
		pluginName   string
		pluginConfig internal.PluginConfig
		wantErr      bool
		errContains  string
	}{
		{
			name:       "load openssl plugin successfully",
			pluginName: "openssl",
			pluginConfig: internal.PluginConfig{
				Enabled: true,
				Config:  map[string]any{},
			},
			wantErr: false,
		},
		{
			name:       "load non-existent built-in plugin",
			pluginName: "non-existent",
			pluginConfig: internal.PluginConfig{
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
			reg, err := NewRegistry(map[string]internal.PluginConfig{
				tt.pluginName: tt.pluginConfig,
			}, cfg)
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
			err = plugin.Initialize(ctx, tt.pluginConfig.Config, cfg)
			require.NoError(t, err)

			// Test GetMetadata
			metadata, err := plugin.GetMetadata(ctx, model.DomainEntry{Domain: "example.com"})
			require.NoError(t, err)
			assert.NotNil(t, metadata)

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
		Ca:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg, err := NewRegistry(map[string]internal.PluginConfig{
		"openssl": {
			Enabled: true,
			Config:  map[string]any{},
		},
	}, cfg)
	require.NoError(t, err)

	// Try to load same plugin again
	err = reg.LoadPlugin("openssl", internal.PluginConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin openssl is already loaded")
}

func TestGetNonExistentPlugin(t *testing.T) {
	cfg := &dehydrated.Config{
		BaseDir:       "/test/base",
		CertDir:       "/test/certs",
		DomainsDir:    "/test/domains",
		ChallengeType: "dns-01",
		Ca:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg, err := NewRegistry(map[string]internal.PluginConfig{}, cfg)
	require.NoError(t, err)

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
		Ca:            "https://acme-v02.api.letsencrypt.org/directory",
	}

	reg, err := NewRegistry(map[string]internal.PluginConfig{
		"openssl": {
			Enabled: true,
			Config:  map[string]any{},
		},
	}, cfg)
	require.NoError(t, err)

	// Close registry
	ctx := context.Background()
	err = reg.Close(ctx)
	require.NoError(t, err)

	// Verify plugin is removed
	plugin, err := reg.GetPlugin("openssl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plugin openssl not found")
	assert.Nil(t, plugin)
}
