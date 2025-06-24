package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/manager"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPluginRegistry(t *testing.T) {
	logger := zap.NewNop()
	manager := manager.NewManager(logger, "")

	tests := []struct {
		name        string
		config      config.RegistryConfig
		expectError bool
	}{
		{
			name: "local registry",
			config: config.RegistryConfig{
				Type: "local",
				Config: map[string]any{
					"path": "/tmp/test-plugin",
				},
			},
			expectError: false,
		},
		{
			name: "github registry",
			config: config.RegistryConfig{
				Type: "github",
				Config: map[string]any{
					"repository": "test/repo",
					"version":    "v1.0.0",
				},
			},
			expectError: false,
		},
		{
			name: "unsupported registry type",
			config: config.RegistryConfig{
				Type:   "unsupported",
				Config: map[string]any{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := NewPluginRegistry(tt.config, logger, manager)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, registry)
			} else {
				require.NoError(t, err)
				require.NotNil(t, registry)
			}
		})
	}
}

func TestLocalRegistry_GetPluginPath(t *testing.T) {
	logger := zap.NewNop()

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-plugin-*")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	config := map[string]any{
		"path": tempFile.Name(),
	}

	registry := NewLocalRegistry(config, logger)

	path, err := registry.GetPluginPath()
	require.NoError(t, err)
	require.NotEmpty(t, path)

	// Should return absolute path
	absPath, err := filepath.Abs(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, absPath, path)
}

func TestLocalRegistry_Validate(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"path": "/tmp/test-plugin",
			},
			expectError: false,
		},
		{
			name:        "missing path",
			config:      map[string]any{},
			expectError: true,
		},
		{
			name: "empty path",
			config: map[string]any{
				"path": "",
			},
			expectError: true,
		},
		{
			name: "wrong path type",
			config: map[string]any{
				"path": 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewLocalRegistry(tt.config, logger)
			err := registry.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGitHubRegistry_Validate(t *testing.T) {
	logger := zap.NewNop()
	manager := manager.NewManager(logger, "")

	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"repository": "test/repo",
				"version":    "v1.0.0",
			},
			expectError: false,
		},
		{
			name: "missing repository",
			config: map[string]any{
				"version": "v1.0.0",
			},
			expectError: true,
		},
		{
			name: "empty repository",
			config: map[string]any{
				"repository": "",
				"version":    "v1.0.0",
			},
			expectError: true,
		},
		{
			name: "invalid repository format",
			config: map[string]any{
				"repository": "invalid-repo-format",
				"version":    "v1.0.0",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewGitHubRegistry(tt.config, logger, manager)
			err := registry.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
