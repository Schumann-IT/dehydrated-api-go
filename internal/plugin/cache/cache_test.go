package cache

import (
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/stretchr/testify/require"
)

func TestPluginConfig_NewRegistry(t *testing.T) {
	tests := []struct {
		name        string
		config      config.PluginConfig
		expectError bool
	}{
		//{
		//	name: "valid_local_registry",
		//	config: config.PluginConfig{
		//		Enabled: true,
		//		Registry: &config.RegistryConfig{
		//			Type: "local",
		//			Config: map[string]any{
		//				"path": "../../../examples/plugins/simple/simple",
		//			},
		//		},
		//	},
		//	expectError: false,
		//},
		{
			name: "missing_registry_type",
			config: config.PluginConfig{
				Enabled: true,
				Registry: &config.RegistryConfig{
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expectError: true,
		},
		{
			name: "unsupported_registry_type",
			config: config.PluginConfig{
				Enabled: true,
				Registry: &config.RegistryConfig{
					Type: "unsupported",
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expectError: true,
		},
		{
			name: "local_registry_missing_path",
			config: config.PluginConfig{
				Enabled: true,
				Registry: &config.RegistryConfig{
					Type:   "local",
					Config: map[string]any{},
				},
			},
			expectError: true,
		},
		{
			name: "github_registry_missing_repository",
			config: config.PluginConfig{
				Enabled: true,
				Registry: &config.RegistryConfig{
					Type: "github",
					Config: map[string]any{
						"version": "v1.0.0",
					},
				},
			},
			expectError: true,
		},
		{
			name: "github_registry_invalid_repository_format",
			config: config.PluginConfig{
				Enabled: true,
				Registry: &config.RegistryConfig{
					Type: "github",
					Config: map[string]any{
						"repository": "invalid-repo-format",
						"version":    "v1.0.0",
					},
				},
			},
			expectError: true,
		},
	}

	tmp := t.TempDir()
	err := Prepare(tmp)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi, err := Add(tt.name, tt.config.Registry)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				path, _ := pi.Path(tt.name)
				require.Contains(t, path, tt.name)
			}
		})
	}

	Clean()
}
