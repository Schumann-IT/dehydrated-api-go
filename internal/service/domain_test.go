package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDomainService(t *testing.T) {
	tests := []struct {
		name        string
		withWatcher bool
	}{
		{
			name:        "WithWatcher",
			withWatcher: true,
		},
		{
			name:        "WithoutWatcher",
			withWatcher: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create plugin config with a built-in timestamp plugin
			pluginConfig := map[string]internal.PluginConfig{}

			// Create domain service config
			config := DomainServiceConfig{
				DehydratedBaseDir: tmpDir,
				EnableWatcher:     tt.withWatcher,
				PluginConfig:      pluginConfig,
			}

			// Create a domain service
			service, err := NewDomainService(config)
			assert.NoError(t, err)
			defer service.Close()

			// Test CreateDomain
			t.Run("CreateDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "example.com",
				}
				entry, err := service.CreateDomain(req)
				assert.NoError(t, err)
				assert.Equal(t, "example.com", entry.Domain)
			})

			// Test CreateInvalidDomain
			t.Run("CreateInvalidDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "invalid..domain",
				}
				_, err := service.CreateDomain(req)
				assert.Error(t, err)
			})

			// Test CreateDuplicateDomain
			t.Run("CreateDuplicateDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "example.com",
				}
				_, err := service.CreateDomain(req)
				assert.Error(t, err)
			})

			// Test GetDomain
			t.Run("GetDomain", func(t *testing.T) {
				entry, err := service.GetDomain("example.com")
				assert.NoError(t, err)
				assert.Equal(t, "example.com", entry.Domain)
			})

			// Test GetNonExistentDomain
			t.Run("GetNonExistentDomain", func(t *testing.T) {
				_, err := service.GetDomain("nonexistent.com")
				assert.Error(t, err)
			})

			// Test UpdateDomain
			t.Run("UpdateDomain", func(t *testing.T) {
				req := model.UpdateDomainRequest{
					Enabled: true,
				}
				entry, err := service.UpdateDomain("example.com", req)
				assert.NoError(t, err)
				assert.True(t, entry.Enabled)
			})

			// Test ListDomains
			t.Run("ListDomains", func(t *testing.T) {
				entries, err := service.ListDomains()
				assert.NoError(t, err)
				assert.Len(t, entries, 1)
				assert.Equal(t, "example.com", entries[0].Domain)
			})

			// Test DeleteDomain
			t.Run("DeleteDomain", func(t *testing.T) {
				err := service.DeleteDomain("example.com")
				assert.NoError(t, err)

				_, err = service.GetDomain("example.com")
				assert.Error(t, err)
			})
		})
	}
}

func TestNewDomainService(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Test with valid config
	t.Run("ValidConfig", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     true,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		if service.domainsFile != domainsFile {
			t.Errorf("Expected domains file %s, got %s", domainsFile, service.domainsFile)
		}
		if service.watcher == nil {
			t.Error("Expected watcher to be initialized")
		}
	})

	// Test with invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "nonexistent", "domains.txt")
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: filepath.Join(tmpDir, "nonexistent"),
			EnableWatcher:     true,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		// Verify that the directory and file were created
		if _, err := os.Stat(invalidPath); os.IsNotExist(err) {
			t.Error("Expected domains file to be created")
		}
	})

	// Test without watcher
	t.Run("WithoutWatcher", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		if service.watcher != nil {
			t.Error("Expected watcher to be nil")
		}
	})
}

// buildTestPlugin builds the test plugin
func buildTestPlugin(t *testing.T) string {
	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Build the test plugin
	pluginPath := filepath.Join(dir, "..", "..", "internal", "plugin", "grpc", "testdata", "test-plugin", "test-plugin")
	cmd := exec.Command("go", "build", "-o", pluginPath, "main.go")
	cmd.Dir = filepath.Join(dir, "..", "..", "internal", "plugin", "grpc", "testdata", "test-plugin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test plugin: %v", err)
	}

	return pluginPath
}
