package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginConfig(t *testing.T) {
	t.Run("ValidLocalPlugin", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			Path:    "/usr/local/bin/plugin",
			Config:  map[string]any{},
		}

		err := config.Validate()
		require.NoError(t, err)
		require.False(t, config.IsGitHubPlugin())
		require.Equal(t, "/usr/local/bin/plugin", config.GetPluginPath())
	})

	t.Run("ValidGitHubPlugin", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			GitHub: &GitHubConfig{
				Repository: "owner/repo",
				Version:    "v1.0.0",
				Platform:   "linux-amd64",
			},
			Config: map[string]any{},
		}

		err := config.Validate()
		require.NoError(t, err)
		require.True(t, config.IsGitHubPlugin())
		require.Empty(t, config.GetPluginPath())
		require.NotNil(t, config.GetGitHubInfo())
	})

	t.Run("PathOverride", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			Path:    "/custom/path/plugin",
			GitHub: &GitHubConfig{
				Repository: "owner/repo",
				Version:    "v1.0.0",
			},
			Config: map[string]any{},
		}

		err := config.Validate()
		require.NoError(t, err)
		require.True(t, config.IsGitHubPlugin())
		require.Equal(t, "/custom/path/plugin", config.GetPluginPath())
	})

	t.Run("DisabledPlugin", func(t *testing.T) {
		config := PluginConfig{
			Enabled: false,
			Config:  map[string]any{},
		}

		err := config.Validate()
		require.NoError(t, err) // Disabled plugins don't need validation
	})

	t.Run("InvalidGitHubRepository", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			GitHub: &GitHubConfig{
				Repository: "invalid-repo-format",
				Version:    "v1.0.0",
			},
			Config: map[string]any{},
		}

		err := config.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid GitHub repository format")
	})

	t.Run("EmptyGitHubRepository", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			GitHub: &GitHubConfig{
				Repository: "",
				Version:    "v1.0.0",
			},
			Config: map[string]any{},
		}

		err := config.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "GitHub repository is required")
	})

	t.Run("NoConfiguration", func(t *testing.T) {
		config := PluginConfig{
			Enabled: true,
			Config:  map[string]any{},
		}

		err := config.Validate()
		require.NoError(t, err) // This is valid - will be handled by manager
	})
}

func TestGitHubConfig(t *testing.T) {
	t.Run("ValidRepository", func(t *testing.T) {
		config := GitHubConfig{
			Repository: "owner/repo",
			Version:    "v1.0.0",
			Platform:   "linux-amd64",
		}

		require.Equal(t, "owner/repo", config.Repository)
		require.Equal(t, "v1.0.0", config.Version)
		require.Equal(t, "linux-amd64", config.Platform)
	})

	t.Run("DefaultVersion", func(t *testing.T) {
		config := GitHubConfig{
			Repository: "owner/repo",
			Platform:   "linux-amd64",
		}

		require.Empty(t, config.Version) // Will default to "latest" in manager
	})
}

func TestToProto(t *testing.T) {
	config := PluginConfig{
		Enabled: true,
		Path:    "/usr/local/bin/plugin",
		Config: map[string]any{
			"string_value": "test",
			"int_value":    42,
			"bool_value":   true,
			"float_value":  3.14,
		},
	}

	proto, err := config.ToProto()
	require.NoError(t, err)
	require.NotNil(t, proto)
	require.Len(t, proto, 4)

	// Check that values are properly converted
	require.Equal(t, "test", proto["string_value"].GetStringValue())
	require.Equal(t, float64(42), proto["int_value"].GetNumberValue())
	require.True(t, proto["bool_value"].GetBoolValue())
	require.Equal(t, 3.14, proto["float_value"].GetNumberValue())
}

func TestPluginConfig_NewRegistry(t *testing.T) {
	tests := []struct {
		name        string
		config      PluginConfig
		expectError bool
	}{
		{
			name: "valid local registry",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Type: "local",
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid github registry",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Type: "github",
					Config: map[string]any{
						"repository": "test/repo",
						"version":    "v1.0.0",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing registry type",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expectError: true,
		},
		{
			name: "unsupported registry type",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Type: "unsupported",
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expectError: true,
		},
		{
			name: "local registry missing path",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Type:   "local",
					Config: map[string]any{},
				},
			},
			expectError: true,
		},
		{
			name: "github registry missing repository",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
					Type: "github",
					Config: map[string]any{
						"version": "v1.0.0",
					},
				},
			},
			expectError: true,
		},
		{
			name: "github registry invalid repository format",
			config: PluginConfig{
				Enabled: true,
				Registry: &RegistryConfig{
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPluginConfig_IsGitHubPlugin_NewRegistry(t *testing.T) {
	tests := []struct {
		name     string
		config   PluginConfig
		expected bool
	}{
		{
			name: "github registry type",
			config: PluginConfig{
				Registry: &RegistryConfig{
					Type: "github",
					Config: map[string]any{
						"repository": "test/repo",
					},
				},
			},
			expected: true,
		},
		{
			name: "local registry type",
			config: PluginConfig{
				Registry: &RegistryConfig{
					Type: "local",
					Config: map[string]any{
						"path": "/tmp/test-plugin",
					},
				},
			},
			expected: false,
		},
		{
			name: "no registry config",
			config: PluginConfig{
				GitHub: &GitHubConfig{
					Repository: "test/repo",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsGitHubPlugin()
			require.Equal(t, tt.expected, result)
		})
	}
}
