package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test default values
	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}

	if len(cfg.Plugins) != 0 {
		t.Errorf("Expected empty plugins map, got %d entries", len(cfg.Plugins))
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := fmt.Sprintf(`
port: 8081
dehydratedBaseDir: %s
plugins:
  test:
    enabled: true
    path: %s/bin/test-plugin
    config:
      key: value
`, tmpDir, tmpDir)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a dummy plugin file
	pluginPath := filepath.Join(tmpDir, "bin", "test-plugin")
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}
	if err := os.WriteFile(pluginPath, []byte("#!/bin/bash\nexit 0"), 0755); err != nil {
		t.Fatalf("Failed to create plugin file: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test loaded values
	if cfg.Port != 8081 {
		t.Errorf("Expected port 8081, got %d", cfg.Port)
	}

	if cfg.DehydratedBaseDir != tmpDir {
		t.Errorf("Expected config dir %s, got %s", tmpDir, cfg.DehydratedBaseDir)
	}

	plugin, ok := cfg.Plugins["test"]
	if !ok {
		t.Error("Expected test plugin to be configured")
	}

	if !plugin.Enabled {
		t.Error("Expected test plugin to be enabled")
	}

	if plugin.Path != pluginPath {
		t.Errorf("Expected plugin path %s, got %s", pluginPath, plugin.Path)
	}

	if plugin.Config["key"] != "value" {
		t.Errorf("Expected plugin config key=value, got %v", plugin.Config["key"])
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "Invalid port",
			content: `
port: 0
dehydratedBaseDir: /etc/dehydrated
`,
			expected: "invalid port number: 0",
		},
		{
			name: "Missing plugin path",
			content: `
port: 8080
dehydratedBaseDir: /
plugins:
  test:
    enabled: true
`,
			expected: "plugin path is required for enabled plugin: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Try to load the config
			_, err := LoadConfig(configPath)
			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != nil && err.Error() != tt.expected {
				t.Errorf("Expected error %q, got %q", tt.expected, err.Error())
			}
		})
	}
}
