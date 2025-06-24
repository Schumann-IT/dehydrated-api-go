package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatcher(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create initial file
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a channel to track changes
	changes := make(chan struct{}, 1)
	onChange := func() error {
		changes <- struct{}{}
		return nil
	}

	// Create watcher
	watcher, err := NewFileWatcher(testFile, onChange)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	watcher.Watch()
	defer watcher.Close()

	// Test file modification
	t.Run("FileModification", func(t *testing.T) {
		// Modify the file
		if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for change notification
		select {
		case <-changes:
			// Change detected successfully
		case <-time.After(time.Second):
			t.Error("Timeout waiting for file change notification")
		}
	})

	// Test debouncing
	t.Run("Debouncing", func(t *testing.T) {
		// Make multiple rapid modifications
		for range 5 {
			if err := os.WriteFile(testFile, []byte("rapid change"), 0644); err != nil {
				t.Fatalf("Failed to modify test file: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Should receive fewer notifications than modifications due to debouncing
		notifications := 0
		timeout := time.After(500 * time.Millisecond)

		for {
			select {
			case <-changes:
				notifications++
			case <-timeout:
				if notifications >= 5 {
					t.Error("Debouncing failed: received too many notifications")
				}
				return
			}
		}
	})

	// Test file recreation
	t.Run("FileRecreation", func(t *testing.T) {
		// Remove and recreate the file
		if err := os.Remove(testFile); err != nil {
			t.Fatalf("Failed to remove test file: %v", err)
		}
		if err := os.WriteFile(testFile, []byte("recreated content"), 0644); err != nil {
			t.Fatalf("Failed to recreate test file: %v", err)
		}

		// Wait for change notification
		select {
		case <-changes:
			// Change detected successfully
		case <-time.After(time.Second):
			t.Error("Timeout waiting for file recreation notification")
		}
	})
}
