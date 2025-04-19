package service

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/assert"
)

// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.

// TestDomainService tests the core functionality of the DomainService.
// It verifies domain creation, retrieval, updating, listing, and deletion operations
// with both watcher enabled and disabled configurations.
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

			dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
			service := NewDomainService(dc)
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

// TestNewDomainService tests the initialization of the DomainService.
// It verifies proper setup with valid and invalid configurations,
// including watcher initialization and file path handling.
func TestNewDomainService(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Test with valid config
	t.Run("ValidConfig", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc).WithFileWatcher()
		defer service.Close()

		if service.DehydratedConfig.DomainsFile != domainsFile {
			t.Errorf("Expected domains file %s, got %s", domainsFile, service.DehydratedConfig.DomainsFile)
		}
		if service.watcher == nil {
			t.Error("Expected watcher to be initialized")
		}
	})

	// Test with invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "nonexistent", "domains.txt")

		dc := &dehydrated.Config{
			DomainsFile: invalidPath,
		}

		service := NewDomainService(dc)
		defer service.Close()

		// Verify that the directory and file were created
		if _, err := os.Stat(invalidPath); os.IsNotExist(err) {
			t.Error("Expected domains file to be created")
		}
	})

	// Test without watcher
	t.Run("WithoutWatcher", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc)
		defer service.Close()

		if service.watcher != nil {
			t.Error("Expected watcher to be nil")
		}
	})
}

// TestDomainServiceErrors tests error handling in the DomainService.
// It verifies proper error responses for invalid operations and edge cases.
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

		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		service := NewDomainService(dc).WithPlugins(pluginConfig)
		service.Close()
	})

	t.Run("CacheReloadFailure", func(t *testing.T) {
		// Create a read-only directory to force cache reload failure
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		assert.NoError(t, err)

		dc := &dehydrated.Config{
			DomainsFile: filepath.Join(readOnlyDir, "domains.txt"),
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		_ = NewDomainService(dc)
	})
}

// TestMetadataEnrichment tests the metadata enrichment functionality.
// It verifies that domain entries are properly enriched with metadata from plugins.
func TestMetadataEnrichment(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("SuccessfulEnrichment", func(t *testing.T) {
		pluginConfig := map[string]plugin.PluginConfig{
			"test": {
				Enabled: true,
				Path:    mustGetPluginPath(t),
			},
		}

		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc).WithPlugins(pluginConfig)
		defer service.Close()

		// Create a domain
		req := model.CreateDomainRequest{
			Domain: "example.com",
		}
		_, err := service.CreateDomain(req)
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

		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		service := NewDomainService(dc).WithPlugins(pluginConfig)
		service.Close()
	})
}

// TestConcurrentOperations tests the thread-safety of the DomainService.
// It verifies that concurrent operations on the service work correctly
// without race conditions or data corruption.
func TestConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// load dehydrated config
	dc := dehydrated.NewConfig().WithBaseDir(tmpDir)

	l, _ := logger.NewLogger(nil)
	service := NewDomainService(dc).WithLogger(l)
	defer service.Close()

	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		// Start multiple goroutines that read and write
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// Use unique domain name for each goroutine
				domain := fmt.Sprintf("domain%d.com", idx)

				// Create domain
				req := model.CreateDomainRequest{
					Domain: domain,
				}
				_, err := service.CreateDomain(req)
				if err != nil {
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

// TestEdgeCases tests various edge cases in the DomainService.
// It verifies proper handling of special domain names, empty values,
// and other boundary conditions.
func TestEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("EmptyDomainList", func(t *testing.T) {
		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)

		service := NewDomainService(dc)
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

		// load dehydrated config
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		service := NewDomainService(dc).WithPlugins(pluginConfig)
		service.Close()
	})

	t.Run("FileSystemErrors", func(t *testing.T) {
		// Create a read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		assert.NoError(t, err)

		dc := &dehydrated.Config{
			DomainsFile: filepath.Join(readOnlyDir, "domains.txt"),
		}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		// Service creation should fail due to read-only directory
		_ = NewDomainService(dc)
	})
}

// TestFileWatcherEdgeCases tests edge cases in the file watcher functionality.
// It verifies proper handling of file system events, including rapid changes
// and file system errors.
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

		// Start the watcher
		watcher.Watch()

		// Give the watcher a moment to initialize
		time.Sleep(100 * time.Millisecond)

		// Delete the file
		err = os.Remove(domainsFile)
		assert.NoError(t, err)

		// Give the watcher time to process the deletion
		time.Sleep(200 * time.Millisecond)

		// Recreate the file with different content
		err = os.WriteFile(domainsFile, []byte("example2.com"), 0644)
		assert.NoError(t, err)

		// Wait for the callback to be called with a timeout
		select {
		case <-callbackCh:
			// Success: callback was called
		case <-time.After(2 * time.Second):
			t.Error("Callback was not called within the timeout period")
		}
	})
}

// TestDomainServiceCleanup tests the cleanup functionality of the DomainService.
// It verifies that resources are properly released when the service is closed.
func TestDomainServiceCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("CleanupWithWatcher", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)
		service := NewDomainService(dc).WithFileWatcher()
		assert.NotNil(t, service.watcher)

		// Wait a bit for the watcher to initialize
		time.Sleep(100 * time.Millisecond)

		err := service.Close()
		assert.NoError(t, err)
		// Note: We can't assert service.watcher is nil because Close() only stops the watcher
		// but doesn't set it to nil. This is an implementation detail.
	})

	t.Run("CleanupWithoutWatcher", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)
		service := NewDomainService(dc)
		assert.Nil(t, service.watcher)

		err := service.Close()
		assert.NoError(t, err)
	})
}

// TestDomainServiceOperations tests various domain service operations.
// It verifies the complete lifecycle of domain entries, including creation,
// modification, and deletion with various configurations.
func TestDomainServiceOperations(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("UpdateNonExistentDomain", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)
		service := NewDomainService(dc)
		defer service.Close()

		req := model.UpdateDomainRequest{
			Enabled: true,
		}
		_, err := service.UpdateDomain("nonexistent.com", req)
		assert.Error(t, err)
	})

	t.Run("DeleteNonExistentDomain", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)
		service := NewDomainService(dc)
		defer service.Close()

		err := service.DeleteDomain("nonexistent.com")
		assert.Error(t, err)
	})

	t.Run("CreateDomainWithInvalidMetadata", func(t *testing.T) {
		pluginConfig := map[string]plugin.PluginConfig{
			"test": {
				Enabled: true,
				Path:    mustGetPluginPath(t),
			},
		}

		dc := dehydrated.NewConfig().WithBaseDir(tmpDir)
		service := NewDomainService(dc).WithPlugins(pluginConfig)
		defer service.Close()

		// Create a domain with metadata
		req := model.CreateDomainRequest{
			Domain: "example.com",
			Metadata: map[string]string{
				"test": "value",
			},
		}

		// Create the domain
		_, err := service.CreateDomain(req)
		assert.NoError(t, err)

		// Get the domain to verify metadata
		enriched, err := service.GetDomain("example.com")
		assert.NoError(t, err)
		assert.NotNil(t, enriched.Metadata)
		assert.Contains(t, enriched.Metadata, "test")
	})
}

func mustGetPluginPath(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	p := filepath.Join(dir, "..", "plugin", "grpc", "testdata", "test-plugin", "test-plugin")

	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("Failed to get abs path for %s: %v", p, err)
	}

	return abs
}
