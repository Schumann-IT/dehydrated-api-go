// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches for changes to a file and triggers a callback when changes are detected.
// It implements debouncing to prevent multiple rapid callbacks for the same file change.
type FileWatcher struct {
	filePath         string               // Path to the file being watched
	watcher          *fsnotify.Watcher    // Underlying filesystem watcher
	onChange         func() error         // Callback function to execute on file changes
	mutex            sync.Mutex           // Mutex for thread-safe access to debounce map
	debounceMap      map[string]time.Time // Map for tracking last event time per file
	done             chan struct{}        // Channel for signaling shutdown
	logger           *zap.Logger          // Logger for the file watcher
	suspended        bool                 // Flag to indicate if the watcher is suspended
	debounceInterval time.Duration        // Interval for debouncing file change events
}

// NewFileWatcher creates a new FileWatcher instance for the specified file.
// It sets up the filesystem watcher and starts a goroutine to monitor for changes.
// The onChange callback will be called when the file is modified, created, or removed.
func NewFileWatcher(filePath string, onChange func() error) (*FileWatcher, error) {
	// Validate inputs
	if onChange == nil {
		return nil, fmt.Errorf("onChange callback cannot be nil")
	}

	// Check if the directory exists (we'll watch the directory, not the file)
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dirPath)
	}

	fw := &FileWatcher{
		filePath:         filePath,
		onChange:         onChange,
		logger:           zap.NewNop(),
		suspended:        false,
		debounceInterval: 100 * time.Millisecond,
	}

	return fw, nil
}

func (fw *FileWatcher) reset() error {
	_ = fw.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fw.watcher = watcher
	if err = fw.watcher.Add(filepath.Dir(fw.filePath)); err != nil {
		_ = fw.watcher.Close()
		fw.watcher = nil
		return err
	}

	fw.debounceMap = make(map[string]time.Time)
	fw.done = make(chan struct{})

	fw.reload()

	return nil
}

func (fw *FileWatcher) WithLogger(l *zap.Logger) *FileWatcher {
	fw.logger = l
	return fw
}

func (fw *FileWatcher) Watch() {
	if err := fw.reset(); err != nil {
		fw.logger.Error("Failed to watch",
			zap.String("file", fw.filePath),
			zap.Error(err))
		return
	}

	go fw.watch()
}

func (fw *FileWatcher) Disable() {
	fw.suspended = true
	fw.logger.Debug("Disabled file watcher")
}

func (fw *FileWatcher) Enable() {
	fw.logger.Debug("Enable file watcher and reload entries.")
	fw.suspended = false
	err := fw.reset()
	if err != nil {
		fw.logger.Error("Failed to reload entries after enabling watcher")
	}
}

// watch monitors the file for changes and triggers the callback when appropriate.
// It implements debouncing to prevent multiple rapid callbacks for the same file change.
// The method runs in a goroutine and continues until the watcher is closed.
func (fw *FileWatcher) watch() {
	fw.logger.Info("Starting file watcher",
		zap.String("file", fw.filePath),
		zap.String("dir", filepath.Dir(fw.filePath)),
		zap.Duration("debounce", fw.debounceInterval))

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if fw.removed(event) {
				fw.logger.Info("File removed, resetting watcher", zap.String("file", event.Name))
				_ = fw.reset()
				continue
			}

			// If the watcher is suspended, skip processing events
			if fw.suspended {
				fw.logger.Debug("Watcher is suspended, ignoring event",
					zap.String("event", event.Op.String()),
					zap.String("file", event.Name))
				continue
			}

			if !fw.shouldHandle(event) {
				continue
			}

			fw.logger.Info("Handling event",
				zap.String("operation", event.Op.String()),
				zap.String("file", event.Name))

			if !fw.shouldDebounce(event) {
				fw.logger.Debug("Triggering onChange callback",
					zap.String("operation", event.Op.String()),
					zap.String("file", event.Name))

				if err := fw.onChange(); err != nil {
					fw.logger.Error("Callback onChange failed",
						zap.String("operation", event.Op.String()),
						zap.String("file", event.Name),
						zap.Error(err))
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Error("Watcher error", zap.Error(err))
		case <-fw.done:
			fw.done = nil
			return
		}
	}
}

func (fw *FileWatcher) shouldDebounce(event fsnotify.Event) bool {
	debounce := false

	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	// If the file was recreated, clear the debounce entry so new events are not ignored
	if event.Op&fsnotify.Create != 0 {
		delete(fw.debounceMap, event.Name)
		// Try to re-add the directory to the watcher in case the file was recreated
		dirPath := filepath.Dir(fw.filePath)
		if err := fw.watcher.Remove(dirPath); err != nil {
			fw.logger.Warn("Failed to remove directory from watcher", zap.String("dir", dirPath), zap.Error(err))
		}
		if err := fw.watcher.Add(dirPath); err != nil {
			fw.logger.Warn("Failed to re-add directory to watcher", zap.String("dir", dirPath), zap.Error(err))
		} else {
			fw.logger.Debug("Re-added directory to watcher after file recreation", zap.String("dir", dirPath))
		}
	}

	now := time.Now()
	if lastEventTime, exists := fw.debounceMap[event.Name]; exists && now.Sub(lastEventTime) <= fw.debounceInterval {
		fw.logger.Debug("Debouncing event",
			zap.String("operation", event.Op.String()),
			zap.String("file", event.Name))
		debounce = true
	}

	fw.debounceMap[event.Name] = now

	return debounce
}

func (fw *FileWatcher) shouldHandle(event fsnotify.Event) bool {
	// It's generally not recommended to take action on fsnotify.Chmod, as it may
	// get triggered very frequently by some software. For example, Spotlight
	// indexing on macOS, anti-virus software, backup software, etc.
	if event.Op == fsnotify.Chmod {
		fw.logger.Debug("Ignoring chmod event")
		return false
	}

	// Only handle Rename, Write, and Create events
	if event.Op&(fsnotify.Rename|fsnotify.Write|fsnotify.Create) == 0 {
		fw.logger.Debug("Ignoring unsupported operation",
			zap.String("operation", event.Op.String()),
			zap.String("event_path", event.Name),
			zap.String("watch_path", fw.filePath))

		return false
	}

	return fw.canHandle(event)
}

func (fw *FileWatcher) removed(event fsnotify.Event) bool {
	if event.Op&(fsnotify.Remove) != 0 && fw.canHandle(event) {
		return true
	}

	return false
}

func (fw *FileWatcher) canHandle(event fsnotify.Event) bool {
	// first try direct comparison after cleaning
	if filepath.Clean(event.Name) == filepath.Clean(fw.filePath) {
		return true
	}

	// Try to resolve both paths to their absolute paths
	abs1, err1 := filepath.Abs(event.Name)
	abs2, err2 := filepath.Abs(fw.filePath)

	if err1 == nil && err2 == nil && filepath.Clean(abs1) == filepath.Clean(abs2) {
		return true
	}

	fw.logger.Debug("Ignoring event due to path mismatch",
		zap.String("operation", event.Op.String()),
		zap.String("event_path", event.Name),
		zap.String("watch_path", fw.filePath))

	return false
}

func (fw *FileWatcher) reload() {
	if err := fw.onChange(); err != nil {
		fw.logger.Error("Callback onChange failed")
	}
}

// Close stops the file watcher and releases associated resources.
func (fw *FileWatcher) Close() error {
	if fw.done != nil {
		close(fw.done)
	}
	if fw.watcher != nil {
		return fw.watcher.Close()
	}

	return nil
}
