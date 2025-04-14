package service

import (
	"github.com/schumann-it/dehydrated-api-go/internal/plugin"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDomainService(t *testing.T) {
	tests := []struct {
		name        string
		withWatcher bool
	}{
		//{
		//  name:        "WithWatcher",
		//	withWatcher: true,
		// },
		{
			name:        "WithoutWatcher",
			withWatcher: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create plugin config with a built-in timestamp plugin
			pluginConfig := map[string]plugin.PluginConfig{}

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

func TestDomainServiceErrors(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("RegistryCreationFailure", func(t *testing.T) {
		// Create invalid plugin config to force registry creation failure
		pluginConfig := map[string]plugin.PluginConfig{
			"invalid": {
				Enabled: true,
				Path:    "/nonexistent/path",
			},
		}

		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		}

		_, err := NewDomainService(config)
		assert.Error(t, err)
	})

	t.Run("CacheReloadFailure", func(t *testing.T) {
		// Create a read-only directory to force cache reload failure
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		assert.NoError(t, err)

		config := DomainServiceConfig{
			DehydratedBaseDir: readOnlyDir,
			EnableWatcher:     false,
		}

		_, err = NewDomainService(config)
		assert.Error(t, err)
	})

	t.Run("WatcherSetupFailure", func(t *testing.T) {
		// Create a non-existent path to force watcher setup failure
		config := DomainServiceConfig{
			DehydratedBaseDir: "/nonexistent/path",
			EnableWatcher:     true,
		}

		_, err := NewDomainService(config)
		assert.Error(t, err)
	})
}

func TestMetadataEnrichment(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("SuccessfulEnrichment", func(t *testing.T) {
		pluginPath := buildTestPlugin(t)
		pluginConfig := map[string]plugin.PluginConfig{
			"test": {
				Enabled: true,
				Path:    pluginPath,
			},
		}

		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		defer service.Close()

		// Create a domain
		req := model.CreateDomainRequest{
			Domain: "example.com",
		}
		_, err = service.CreateDomain(req)
		assert.NoError(t, err)

		// Get the domain with metadata
		enriched, err := service.GetDomain("example.com")
		assert.NoError(t, err)
		assert.NotNil(t, enriched.Metadata)
		assert.Contains(t, enriched.Metadata, "test")
	})

	t.Run("EnrichmentFailure", func(t *testing.T) {
		// Create a plugin config that will fail
		pluginConfig := map[string]plugin.PluginConfig{
			"failing": {
				Enabled: true,
				Path:    "/nonexistent/path",
			},
		}

		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		}

		// Service creation should fail due to invalid plugin path
		service, err := NewDomainService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

func TestConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	config := DomainServiceConfig{
		DehydratedBaseDir: tmpDir,
		EnableWatcher:     false,
	}

	service, err := NewDomainService(config)
	assert.NoError(t, err)
	defer service.Close()

	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		var wg sync.WaitGroup
		domains := []string{"domain1.com", "domain2.com", "domain3.com"}

		// Start multiple goroutines that read and write
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				domain := domains[idx%len(domains)]

				// Create domain
				req := model.CreateDomainRequest{
					Domain: domain,
				}
				_, err := service.CreateDomain(req)
				if err != nil && err.Error() != "domain already exists: "+domain {
					t.Errorf("Unexpected error creating domain: %v", err)
				}

				// Read domain
				_, err = service.GetDomain(domain)
				if err != nil {
					t.Errorf("Unexpected error getting domain: %v", err)
				}

				// List domains
				_, err = service.ListDomains()
				if err != nil {
					t.Errorf("Unexpected error listing domains: %v", err)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("EmptyDomainList", func(t *testing.T) {
		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		defer service.Close()

		entries, err := service.ListDomains()
		assert.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("InvalidPluginConfig", func(t *testing.T) {
		pluginConfig := map[string]plugin.PluginConfig{
			"invalid": {
				Enabled: true,
				Path:    "/nonexistent/path",
			},
		}

		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		}

		_, err := NewDomainService(config)
		assert.Error(t, err)
	})

	t.Run("FileSystemErrors", func(t *testing.T) {
		// Create a read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		assert.NoError(t, err)

		config := DomainServiceConfig{
			DehydratedBaseDir: readOnlyDir,
			EnableWatcher:     false,
		}

		// Service creation should fail due to read-only directory
		service, err := NewDomainService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

func TestFileWatcherEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	t.Run("InvalidPath", func(t *testing.T) {
		_, err := NewFileWatcher("/nonexistent/path", nil)
		assert.Error(t, err)
	})

	t.Run("NilCallback", func(t *testing.T) {
		_, err := NewFileWatcher(domainsFile, nil)
		assert.Error(t, err)
	})

	t.Run("FileDeletedAndRecreated", func(t *testing.T) {
		callbackCh := make(chan struct{}, 1)
		callback := func() error {
			callbackCh <- struct{}{}
			return nil
		}

		// Create initial file
		err := os.WriteFile(domainsFile, []byte("example.com"), 0644)
		assert.NoError(t, err)

		watcher, err := NewFileWatcher(domainsFile, callback)
		assert.NoError(t, err)
		defer watcher.Close()

		// Delete the file
		err = os.Remove(domainsFile)
		assert.NoError(t, err)

		// Recreate the file with different content
		err = os.WriteFile(domainsFile, []byte("example2.com"), 0644)
		assert.NoError(t, err)

		// Wait for the callback to be called with a timeout
		select {
		case <-callbackCh:
			// Success: callback was called
		case <-time.After(time.Second):
			t.Error("Callback was not called within the timeout period")
		}
	})
}

func TestDomainServiceCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("CleanupWithWatcher", func(t *testing.T) {
		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     true,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		assert.NotNil(t, service.watcher)

		// Wait a bit for the watcher to initialize
		time.Sleep(100 * time.Millisecond)

		err = service.Close()
		assert.NoError(t, err)
		// Note: We can't assert service.watcher is nil because Close() only stops the watcher
		// but doesn't set it to nil. This is an implementation detail.
	})

	t.Run("CleanupWithoutWatcher", func(t *testing.T) {
		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		assert.Nil(t, service.watcher)

		err = service.Close()
		assert.NoError(t, err)
	})
}

func TestDomainServiceOperations(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("UpdateNonExistentDomain", func(t *testing.T) {
		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		defer service.Close()

		req := model.UpdateDomainRequest{
			Enabled: true,
		}
		_, err = service.UpdateDomain("nonexistent.com", req)
		assert.Error(t, err)
	})

	t.Run("DeleteNonExistentDomain", func(t *testing.T) {
		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		defer service.Close()

		err = service.DeleteDomain("nonexistent.com")
		assert.Error(t, err)
	})

	t.Run("CreateDomainWithInvalidMetadata", func(t *testing.T) {
		pluginPath := buildTestPlugin(t)
		pluginConfig := map[string]plugin.PluginConfig{
			"test": {
				Enabled: true,
				Path:    pluginPath,
			},
		}

		config := DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		}

		service, err := NewDomainService(config)
		assert.NoError(t, err)
		defer service.Close()

		// Create a domain with metadata
		req := model.CreateDomainRequest{
			Domain: "example.com",
			Metadata: map[string]string{
				"test": "value",
			},
		}

		// Create the domain
		_, err = service.CreateDomain(req)
		assert.NoError(t, err)

		// Get the domain to verify metadata
		enriched, err := service.GetDomain("example.com")
		assert.NoError(t, err)
		assert.NotNil(t, enriched.Metadata)
		assert.Contains(t, enriched.Metadata, "test")
	})
}
