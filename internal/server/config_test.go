package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin"

	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	// Test default values
	t.Run("DefaultValues", func(t *testing.T) {
		cfg := NewConfig()
		if cfg.Port != 3000 {
			t.Errorf("Expected default port 3000, got %d", cfg.Port)
		}
		if cfg.DehydratedBaseDir != "." {
			t.Errorf("Expected default base dir ., got %s", cfg.DehydratedBaseDir)
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
		cfg.Plugins = map[string]plugin.PluginConfig{
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

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid configuration",
			setupConfig: func() *Config {
				tmpDir := t.TempDir()
				pluginPath := filepath.Join(tmpDir, "test-plugin")
				err := os.WriteFile(pluginPath, []byte("test"), 0755)
				if err != nil {
					t.Fatalf("Failed to create test plugin: %v", err)
				}
				return &Config{
					Port:              3000,
					DehydratedBaseDir: tmpDir,
					Plugins: map[string]plugin.PluginConfig{
						"test": {
							Enabled: true,
							Path:    pluginPath,
						},
					},
				}
			},
			wantErr: false,
		},
		{
			name: "invalid port - below range",
			setupConfig: func() *Config {
				return &Config{
					Port:              0,
					DehydratedBaseDir: ".",
				}
			},
			wantErr:     true,
			errContains: "invalid port number",
		},
		{
			name: "invalid port - above range",
			setupConfig: func() *Config {
				return &Config{
					Port:              65536,
					DehydratedBaseDir: ".",
				}
			},
			wantErr:     true,
			errContains: "invalid port number",
		},
		{
			name: "non-existent dehydrated base dir",
			setupConfig: func() *Config {
				return &Config{
					Port:              3000,
					DehydratedBaseDir: "/non/existent/path",
				}
			},
			wantErr:     true,
			errContains: "dehydrated base dir does not exist",
		},
		{
			name: "enabled plugin without path",
			setupConfig: func() *Config {
				return &Config{
					Port:              3000,
					DehydratedBaseDir: ".",
					Plugins: map[string]plugin.PluginConfig{
						"test": {
							Enabled: true,
							Path:    "",
						},
					},
				}
			},
			wantErr:     true,
			errContains: "plugin path is required for enabled plugin",
		},
		{
			name: "non-existent plugin path",
			setupConfig: func() *Config {
				return &Config{
					Port:              3000,
					DehydratedBaseDir: ".",
					Plugins: map[string]plugin.PluginConfig{
						"test": {
							Enabled: true,
							Path:    "/non/existent/plugin",
						},
					},
				}
			},
			wantErr:     true,
			errContains: "plugin path does not exist",
		},
		{
			name: "relative plugin path",
			setupConfig: func() *Config {
				tmpDir := t.TempDir()
				pluginPath := filepath.Join(tmpDir, "test-plugin")
				err := os.WriteFile(pluginPath, []byte("test"), 0755)
				if err != nil {
					t.Fatalf("Failed to create test plugin: %v", err)
				}
				return &Config{
					Port:              3000,
					DehydratedBaseDir: tmpDir,
					Plugins: map[string]plugin.PluginConfig{
						"test": {
							Enabled: true,
							Path:    "test-plugin", // Relative path
						},
					},
				}
			},
			wantErr:     true,
			errContains: "plugin path does not exist",
		},
		{
			name: "disabled plugin with invalid path",
			setupConfig: func() *Config {
				return &Config{
					Port:              3000,
					DehydratedBaseDir: ".",
					Plugins: map[string]plugin.PluginConfig{
						"test": {
							Enabled: false,
							Path:    "/non/existent/plugin",
						},
					},
				}
			},
			wantErr: false, // Should not validate disabled plugins
		},
		{
			name: "multiple plugins - mixed validity",
			setupConfig: func() *Config {
				tmpDir := t.TempDir()
				validPluginPath := filepath.Join(tmpDir, "valid-plugin")
				err := os.WriteFile(validPluginPath, []byte("test"), 0755)
				if err != nil {
					t.Fatalf("Failed to create test plugin: %v", err)
				}
				return &Config{
					Port:              3000,
					DehydratedBaseDir: tmpDir,
					Plugins: map[string]plugin.PluginConfig{
						"valid": {
							Enabled: true,
							Path:    validPluginPath,
						},
						"disabled": {
							Enabled: false,
							Path:    "/non/existent/plugin",
						},
						"invalid": {
							Enabled: true,
							Path:    "/non/existent/plugin",
						},
					},
				}
			},
			wantErr:     true,
			errContains: "plugin path does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigLoad(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedConfig *Config
		expectError    bool
		setupFiles     func(dir string) error
	}{
		{
			name: "load complete configuration",
			configContent: `
port: 8080
dehydratedBaseDir: /test/dir
enableWatcher: true
plugins:
  test:
    enabled: true
    path: /test/plugin
    config:
      key: value
logging:
  level: debug
  encoding: json
  outputPath: /test/log
`,
			expectError: false,
			expectedConfig: &Config{
				Port:              8080,
				DehydratedBaseDir: "/test/dir",
				EnableWatcher:     true,
				Plugins: map[string]plugin.PluginConfig{
					"test": {
						Enabled: true,
						Path:    "/test/plugin",
						Config: map[string]any{
							"key": "value",
						},
					},
				},
				Logging: &logger.Config{
					Level:      "debug",
					Encoding:   "json",
					OutputPath: "/test/log",
				},
			},
		},
		{
			name: "load partial configuration",
			configContent: `
port: 8080
plugins:
  test:
    enabled: true
`,
			expectError: false,
			expectedConfig: &Config{
				Port:              8080,
				DehydratedBaseDir: ".",
				EnableWatcher:     false,
				Plugins: map[string]plugin.PluginConfig{
					"test": {
						Enabled: true,
						Config:  make(map[string]any),
					},
				},
			},
		},
		{
			name:          "load non-existent file",
			configContent: "",
			expectError:   false,
			expectedConfig: &Config{
				Port:              3000,
				DehydratedBaseDir: ".",
				EnableWatcher:     false,
				Plugins:           make(map[string]plugin.PluginConfig),
			},
		},
		{
			name: "load invalid yaml",
			configContent: `
port: not-a-number
`,
			expectError: true,
			expectedConfig: &Config{
				Port:              3000,
				DehydratedBaseDir: ".",
				EnableWatcher:     false,
				Plugins:           make(map[string]plugin.PluginConfig),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if tt.configContent != "" {
				err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
				assert.NoError(t, err)
			}

			if tt.setupFiles != nil {
				err := tt.setupFiles(tmpDir)
				assert.NoError(t, err)
			}

			cfg := NewConfig()
			cfg.Load(configPath)
			if cfg.err != nil {
				if tt.expectError {
					assert.Error(t, cfg.err)
				} else {
					t.Fatalf("Unexpected error: %v", cfg.err)
				}
				return
			}
			if !filepath.IsAbs(cfg.parsedConfig.DehydratedBaseDir) {
				tt.expectedConfig.DehydratedBaseDir = filepath.Join(tmpDir, tt.expectedConfig.DehydratedBaseDir)
			}

			// Compare configurations
			if tt.expectedConfig != nil {
				assert.Equal(t, tt.expectedConfig.Port, cfg.Port)
				assert.Equal(t, tt.expectedConfig.DehydratedBaseDir, cfg.DehydratedBaseDir)
				assert.Equal(t, tt.expectedConfig.EnableWatcher, cfg.EnableWatcher)
				assert.Equal(t, len(tt.expectedConfig.Plugins), len(cfg.Plugins))

				for name, expectedPlugin := range tt.expectedConfig.Plugins {
					actualPlugin, exists := cfg.Plugins[name]
					assert.True(t, exists)
					assert.Equal(t, expectedPlugin.Enabled, actualPlugin.Enabled)
					assert.Equal(t, expectedPlugin.Path, actualPlugin.Path)
					assert.Equal(t, expectedPlugin.Config, actualPlugin.Config)
				}

				if tt.expectedConfig.Logging != nil {
					assert.NotNil(t, cfg.Logging)
					assert.Equal(t, tt.expectedConfig.Logging.Level, cfg.Logging.Level)
					assert.Equal(t, tt.expectedConfig.Logging.Encoding, cfg.Logging.Encoding)
					assert.Equal(t, tt.expectedConfig.Logging.OutputPath, cfg.Logging.OutputPath)
				}
			}
		})
	}
}
