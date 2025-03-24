package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 8080
dehydratedBaseDir: %s
enableWatcher: true
plugins:
  openssl:
    enabled: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test config loading
	cfg := internal.NewConfig().Load(configPath)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, tmpDir, cfg.DehydratedBaseDir)
	assert.True(t, cfg.EnableWatcher)
	assert.NotNil(t, cfg.Plugins["openssl"])
}

func TestMainIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 0
dehydratedBaseDir: %s
enableWatcher: false
plugins:
  openssl:
    enabled: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set up test environment
	os.Args = []string{"cmd", "-config", configPath}

	// Run main in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main panicked: %v", r)
			}
		}()
		main()
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
}

func TestMainWithInvalidPort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary config file with invalid port
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: -1
dehydratedBaseDir: %s
enableWatcher: false
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set up test environment
	os.Args = []string{"cmd", "-config", configPath}

	// Run main and expect it to fail
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with invalid port, but got none")
		}
	}()
	main()
}
