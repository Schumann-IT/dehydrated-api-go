package registry

import (
	"context"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/pkg/dehydrated"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
