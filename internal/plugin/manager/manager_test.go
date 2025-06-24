package manager

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	logger := zap.NewNop()

	manager := NewManager(logger, tempDir)

	t.Run("LocalPluginPath", func(t *testing.T) {
		// Test with local plugin path
		pluginConfig := config.PluginConfig{
			Enabled: true,
			Path:    "/usr/local/bin/test-plugin",
			Config:  map[string]any{},
		}

		path, err := manager.GetPluginPath(pluginConfig)
		require.NoError(t, err)
		require.Equal(t, "/usr/local/bin/test-plugin", path)
	})

	t.Run("GitHubPluginConfig", func(t *testing.T) {
		// Test GitHub plugin configuration
		pluginConfig := config.PluginConfig{
			Enabled: true,
			GitHub: &config.GitHubConfig{
				Repository: "test-owner/test-repo",
				Version:    "v1.0.0",
				Platform:   "linux-amd64",
			},
			Config: map[string]any{},
		}

		// This should fail because we can't actually download from a non-existent repo
		_, err := manager.GetPluginPath(pluginConfig)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to fetch release info")
	})

	t.Run("PathOverride", func(t *testing.T) {
		// Test that path takes precedence over GitHub config
		pluginConfig := config.PluginConfig{
			Enabled: true,
			Path:    "/custom/path/plugin",
			GitHub: &config.GitHubConfig{
				Repository: "test-owner/test-repo",
				Version:    "v1.0.0",
			},
			Config: map[string]any{},
		}

		path, err := manager.GetPluginPath(pluginConfig)
		require.NoError(t, err)
		require.Equal(t, "/custom/path/plugin", path)
	})

	t.Run("NoConfiguration", func(t *testing.T) {
		// Test with no path or GitHub configuration
		pluginConfig := config.PluginConfig{
			Enabled: true,
			Config:  map[string]any{},
		}

		_, err := manager.GetPluginPath(pluginConfig)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no plugin path or GitHub configuration provided")
	})
}

func TestPlatformDetection(t *testing.T) {
	logger := zap.NewNop()
	manager := NewManager(logger, "")

	platform := manager.detectPlatform()
	require.NotEmpty(t, platform)

	// Should contain OS and architecture
	require.Contains(t, platform, "-")
}

func TestDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	manager := NewManager(logger, tempDir)

	t.Run("ParseRepository", func(t *testing.T) {
		org, pluginName, err := manager.parseRepository("Schumann-IT/dehydrated-api-metadata-plugin-netscaler")
		require.NoError(t, err)
		require.Equal(t, "Schumann-IT", org)
		require.Equal(t, "dehydrated-api-metadata-plugin-netscaler", pluginName)
	})

	t.Run("InvalidRepository", func(t *testing.T) {
		_, _, err := manager.parseRepository("invalid-format")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid repository format")
	})

	t.Run("GenerateCachePath", func(t *testing.T) {
		cachePath := manager.generateCachePath("github", "Schumann-IT", "netscaler-plugin", "v1.0.0", "linux-amd64")
		expectedPath := filepath.Join(tempDir, "github", "Schumann-IT", "netscaler-plugin", "v1.0.0", "linux-amd64", "netscaler-plugin")
		require.Equal(t, expectedPath, cachePath)
	})
}

func TestAssetFinding(t *testing.T) {
	logger := zap.NewNop()
	manager := NewManager(logger, "")

	assets := []GitHubAsset{
		{
			Name:               "plugin-linux-amd64",
			BrowserDownloadURL: "https://example.com/plugin-linux-amd64",
		},
		{
			Name:               "plugin-darwin-amd64",
			BrowserDownloadURL: "https://example.com/plugin-darwin-amd64",
		},
	}

	t.Run("ExactMatch", func(t *testing.T) {
		asset, err := manager.findAsset(assets, "linux-amd64")
		require.NoError(t, err)
		require.Equal(t, "plugin-linux-amd64", asset.Name)
	})

	t.Run("UnderscoreVariation", func(t *testing.T) {
		// Test with underscore variation
		assetsWithUnderscore := []GitHubAsset{
			{
				Name:               "plugin-linux_amd64",
				BrowserDownloadURL: "https://example.com/plugin-linux_amd64",
			},
		}

		asset, err := manager.findAsset(assetsWithUnderscore, "linux-amd64")
		require.NoError(t, err)
		require.Equal(t, "plugin-linux_amd64", asset.Name)
	})

	t.Run("NoMatch", func(t *testing.T) {
		_, err := manager.findAsset(assets, "windows-amd64")
		require.Error(t, err)
		require.Contains(t, err.Error(), "no asset found for platform")
	})
}

func TestPluginCaching(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	manager := NewManager(logger, tempDir)

	// Create a test plugin file
	testPluginPath := filepath.Join(tempDir, "test-plugin")
	err := os.WriteFile(testPluginPath, []byte("#!/bin/sh\necho 'test'"), 0755)
	require.NoError(t, err)

	// Test that the plugin is detected as cached
	require.True(t, manager.isPluginCached(testPluginPath))

	// Test with non-executable file
	nonExecPath := filepath.Join(tempDir, "non-exec")
	err = os.WriteFile(nonExecPath, []byte("not executable"), 0644)
	require.NoError(t, err)

	require.False(t, manager.isPluginCached(nonExecPath))

	// Test with non-existent file
	require.False(t, manager.isPluginCached(filepath.Join(tempDir, "nonexistent")))
}

func TestCleanup(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()
	manager := NewManager(logger, tempDir)

	// Create directory structure with old and new files
	pluginDir := filepath.Join(tempDir, "github", "owner", "plugin", "v1.0.0", "linux-amd64")
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	// Create test files
	oldFile := filepath.Join(pluginDir, "old-plugin")
	newFile := filepath.Join(pluginDir, "new-plugin")

	err = os.WriteFile(oldFile, []byte("old plugin"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(newFile, []byte("new plugin"), 0644)
	require.NoError(t, err)

	// Make the old file appear old
	err = os.Chtimes(oldFile, time.Now().Add(-24*time.Hour), time.Now().Add(-24*time.Hour))
	require.NoError(t, err)

	// Run cleanup with 1 hour max age
	err = manager.Cleanup(1 * time.Hour)
	require.NoError(t, err)

	// Check that old file was removed
	_, err = os.Stat(oldFile)
	require.True(t, os.IsNotExist(err))

	// Check that new file still exists
	_, err = os.Stat(newFile)
	require.NoError(t, err)
}

func TestManagerWithWorkingDirectory(t *testing.T) {
	// Test that manager uses working directory when no cache dir is provided
	logger := zap.NewNop()
	manager := NewManager(logger, "")

	// Should contain .dehydrated-api-go/plugins in the path
	require.Contains(t, manager.cacheDir, ".dehydrated-api-go/plugins")
}
