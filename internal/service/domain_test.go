package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/registry"
	"github.com/schumann-it/dehydrated-api-go/internal/util"
	"github.com/stretchr/testify/require"
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
			service := NewDomainService(dc, nil)
			defer service.Close()

			// Test CreateDomain
			t.Run("CreateDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "example.com",
				}
				entry, err := service.CreateDomain(req)
				require.NoError(t, err)
				require.Equal(t, "example.com", entry.Domain)
			})

			// Test CreateInvalidDomain
			t.Run("CreateInvalidDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "invalid..domain",
				}
				_, err := service.CreateDomain(req)
				require.Error(t, err)
			})

			// Test CreateDuplicateDomain
			t.Run("CreateDuplicateDomain", func(t *testing.T) {
				req := model.CreateDomainRequest{
					Domain: "example.com",
				}
				_, err := service.CreateDomain(req)
				require.Error(t, err)
			})

			// Test GetDomain
			t.Run("GetDomain", func(t *testing.T) {
				entry, err := service.GetDomain("example.com")
				require.NoError(t, err)
				require.Equal(t, "example.com", entry.Domain)
			})

			// Test GetNonExistentDomain
			t.Run("GetNonExistentDomain", func(t *testing.T) {
				_, err := service.GetDomain("nonexistent.com")
				require.Error(t, err)
			})

			// Test UpdateDomain
			t.Run("UpdateDomain", func(t *testing.T) {
				req := model.UpdateDomainRequest{
					Enabled: util.BoolPtr(true),
				}
				entry, err := service.UpdateDomain("example.com", req)
				require.NoError(t, err)
				require.True(t, entry.Enabled)
			})

			// Test ListDomains
			t.Run("ListDomains", func(t *testing.T) {
				entries, err := service.ListDomains()
				require.NoError(t, err)
				require.Len(t, entries, 1)
				require.Equal(t, "example.com", entries[0].Domain)
			})

			// Test DeleteDomain
			t.Run("DeleteDomain", func(t *testing.T) {
				err := service.DeleteDomain("example.com")
				require.NoError(t, err)

				_, err = service.GetDomain("example.com")
				require.Error(t, err)
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
		service := NewDomainService(dc, nil).WithFileWatcher()
		defer service.Close()

		if service.DehydratedConfig.DomainsFile != domainsFile {
			t.Errorf("Expected domains file %s, got %s", domainsFile, service.DehydratedConfig.DomainsFile)
		}
		if service.watcher == nil {
			t.Error("Expected watcher to be initialized")
		}
	})

	// Test without watcher
	t.Run("WithoutWatcher", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc, nil)
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

	t.Run("CacheReloadFailure", func(t *testing.T) {
		// Create a read-only directory to force cache reload failure
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		require.NoError(t, err)

		dc := &dehydrated.Config{
			DehydratedConfig: pb.DehydratedConfig{
				DomainsFile: filepath.Join(readOnlyDir, "domains.txt"),
			},
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		_ = NewDomainService(dc, nil)
	})
}

// TestConcurrentOperations tests the thread-safety of the DomainService.
// It verifies that concurrent operations on the service work correctly
// without race conditions or data corruption.
func TestConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// load dehydrated config
	dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

	l, _ := logger.NewLogger(nil)
	service := NewDomainService(dc, nil).WithLogger(l)
	defer service.Close()

	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		// Start multiple goroutines that read and write
		for i := range numGoroutines {
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
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()

		service := NewDomainService(dc, nil)
		defer service.Close()

		entries, err := service.ListDomains()
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("FileSystemErrors", func(t *testing.T) {
		// Create a read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0444)
		require.NoError(t, err)

		dc := &dehydrated.Config{
			DehydratedConfig: pb.DehydratedConfig{
				DomainsFile: filepath.Join(readOnlyDir, "domains.txt"),
			},
		}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The function did not panic as expected")
			}
		}()

		// Service creation should fail due to read-only directory
		_ = NewDomainService(dc, nil)
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
		require.Error(t, err)
	})

	t.Run("NilCallback", func(t *testing.T) {
		_, err := NewFileWatcher(domainsFile, nil)
		require.Error(t, err)
	})

	t.Run("FileDeletedAndRecreated", func(t *testing.T) {
		callbackCh := make(chan struct{}, 1)
		callback := func() error {
			callbackCh <- struct{}{}
			return nil
		}

		// Create initial file
		err := os.WriteFile(domainsFile, []byte("example.com"), 0644)
		require.NoError(t, err)

		watcher, err := NewFileWatcher(domainsFile, callback)
		require.NoError(t, err)
		defer watcher.Close()

		// Start the watcher
		watcher.Watch()

		// Give the watcher a moment to initialize
		time.Sleep(100 * time.Millisecond)

		// Delete the file
		err = os.Remove(domainsFile)
		require.NoError(t, err)

		// Give the watcher time to process the deletion
		time.Sleep(200 * time.Millisecond)

		// Recreate the file with different content
		err = os.WriteFile(domainsFile, []byte("example2.com"), 0644)
		require.NoError(t, err)

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
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc, nil).WithFileWatcher()
		require.NotNil(t, service.watcher)

		// Wait a bit for the watcher to initialize
		time.Sleep(100 * time.Millisecond)

		err := service.Close()
		require.NoError(t, err)
		// Note: We can't require service.watcher is nil because Close() only stops the watcher
		// but doesn't set it to nil. This is an implementation detail.
	})

	t.Run("CleanupWithoutWatcher", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc, nil)
		require.Nil(t, service.watcher)

		err := service.Close()
		require.NoError(t, err)
	})
}

// TestDomainServiceOperations tests various domain service operations.
// It verifies the complete lifecycle of domain entries, including creation,
// modification, and deletion with various configurations.
func TestDomainServiceOperations(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("UpdateNonExistentDomain", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc, nil)
		defer service.Close()

		req := model.UpdateDomainRequest{
			Enabled: util.BoolPtr(true),
		}
		_, err := service.UpdateDomain("nonexistent.com", req)
		require.Error(t, err)
	})

	t.Run("DeleteNonExistentDomain", func(t *testing.T) {
		dc := dehydrated.NewConfig().WithBaseDir(tmpDir).Load()
		service := NewDomainService(dc, nil)
		defer service.Close()

		err := service.DeleteDomain("nonexistent.com")
		require.Error(t, err)
	})
}

func TestDomainService_UpdateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		req     model.UpdateDomainRequest
		wantErr bool
	}{
		{
			name:   "valid update",
			domain: "example.com",
			req: model.UpdateDomainRequest{
				AlternativeNames: util.StringSlicePtr([]string{"www.example.com"}),
				Enabled:          util.BoolPtr(true),
			},
			wantErr: false,
		},
		{
			name:   "invalid domain",
			domain: "invalid",
			req: model.UpdateDomainRequest{
				AlternativeNames: util.StringSlicePtr([]string{"www.example.com"}),
				Enabled:          util.BoolPtr(true),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new domain service with default config and empty registry
			cfg := dehydrated.NewConfig()
			reg := registry.NewRegistry("", make(map[string]config.PluginConfig), zap.NewNop())
			service := NewDomainService(cfg, reg)

			// Create a test domain
			if tt.domain == "example.com" {
				_, err := service.CreateDomain(model.CreateDomainRequest{
					Domain:           tt.domain,
					AlternativeNames: []string{"www.example.com"},
					Enabled:          true,
				})
				require.NoError(t, err)
			}

			// Update the domain
			updated, err := service.UpdateDomain(tt.domain, tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, updated)

				// Verify the domain was updated
				domain, err := service.GetDomain(tt.domain)
				require.NoError(t, err)
				require.Equal(t, tt.domain, domain.Domain)
				require.Equal(t, util.StringSlice(tt.req.AlternativeNames), domain.AlternativeNames)
				require.Equal(t, util.Bool(tt.req.Enabled), domain.Enabled)
			}
		})
	}
}
