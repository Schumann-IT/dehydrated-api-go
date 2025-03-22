package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	// Test default values
	t.Run("DefaultValues", func(t *testing.T) {
		cfg := NewConfig()
		if cfg.Port != 8080 {
			t.Errorf("Expected default port 8080, got %d", cfg.Port)
		}
		if cfg.DehydratedBaseDir != "/etc/dehydrated" {
			t.Errorf("Expected default base dir /etc/dehydrated, got %s", cfg.DehydratedBaseDir)
		}
		if len(cfg.Plugins) != 0 {
			t.Errorf("Expected no plugins by default, got %d", len(cfg.Plugins))
		}
	})

	// Test loading from file
	t.Run("LoadFromFile", func(t *testing.T) {
		// Create a temporary directory for test files
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		// Create test config file
		configData := []byte(fmt.Sprintf(`
port: 9090
dehydratedBaseDir: %s
plugins:
  test:
    enabled: true
    path: /usr/local/bin/test-plugin
    config:
      key: value
`, tmpDir))
		if err := os.WriteFile(configFile, configData, 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Load config
		cfg := NewConfig()
		data, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			t.Fatalf("Failed to parse config file: %v", err)
		}

		// Verify loaded values
		if cfg.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", cfg.Port)
		}
		if cfg.DehydratedBaseDir != tmpDir {
			t.Errorf("Expected base dir %s, got %s", tmpDir, cfg.DehydratedBaseDir)
		}
		if len(cfg.Plugins) != 1 {
			t.Errorf("Expected 1 plugin, got %d", len(cfg.Plugins))
		}
		if plugin, ok := cfg.Plugins["test"]; !ok {
			t.Error("Expected test plugin to be present")
		} else {
			if !plugin.Enabled {
				t.Error("Expected test plugin to be enabled")
			}
			if plugin.Path != "/usr/local/bin/test-plugin" {
				t.Errorf("Expected plugin path /usr/local/bin/test-plugin, got %s", plugin.Path)
			}
			if val, ok := plugin.Config["key"].(string); !ok || val != "value" {
				t.Errorf("Expected plugin config key=value, got %v", plugin.Config)
			}
		}
	})

	// Test validation
	t.Run("Validation", func(t *testing.T) {
		// Test invalid port
		cfg := NewConfig()
		cfg.Port = 0
		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid port")
		}

		// Test invalid base dir
		cfg = NewConfig()
		cfg.DehydratedBaseDir = "/nonexistent"
		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid base dir")
		}

		// Test invalid plugin config
		cfg = NewConfig()
		cfg.Plugins = map[string]PluginConfig{
			"test": {
				Enabled: true,
				Path:    "",
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid plugin config")
		}
	})
}
