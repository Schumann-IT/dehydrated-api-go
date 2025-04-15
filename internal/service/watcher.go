// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"go.uber.org/zap"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches for changes to a file and triggers a callback when changes are detected.
// It implements debouncing to prevent multiple rapid callbacks for the same file change.
type FileWatcher struct {
	filePath    string               // Path to the file being watched
	watcher     *fsnotify.Watcher    // Underlying filesystem watcher
	onChange    func() error         // Callback function to execute on file changes
	mutex       sync.Mutex           // Mutex for thread-safe access to debounce map
	debounceMap map[string]time.Time // Map for tracking last event time per file
	done        chan struct{}        // Channel for signaling shutdown
	logger      *zap.Logger
}

// NewFileWatcher creates a new FileWatcher instance for the specified file.
// It sets up the filesystem watcher and starts a goroutine to monitor for changes.
// The onChange callback will be called when the file is modified, created, or removed.
func NewFileWatcher(filePath string, onChange func() error) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		filePath:    filePath,
		watcher:     watcher,
		onChange:    onChange,
		debounceMap: make(map[string]time.Time),
		done:        make(chan struct{}),
		logger:      zap.NewNop(),
	}

	// Watch both the file and its parent directory
	if err := watcher.Add(filePath); err != nil {
		watcher.Close()
		return nil, err
	}

	if err := watcher.Add(filepath.Dir(filePath)); err != nil {
		watcher.Close()
		return nil, err
	}

	go fw.watch()

	return fw, nil
}

func (fw *FileWatcher) WithLogger(l *zap.Logger) *FileWatcher {
	fw.logger = l
	return fw
}

// watch monitors the file for changes and triggers the callback when appropriate.
// It implements debouncing to prevent multiple rapid callbacks for the same file change.
// The method runs in a goroutine and continues until the watcher is closed.
func (fw *FileWatcher) watch() {
	const debounceInterval = 100 * time.Millisecond

	fw.logger.Info("Starting file watcher", zap.Any("debounce interval", debounceInterval))

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Check if the event is related to our file
			if event.Name != fw.filePath && filepath.Base(event.Name) != filepath.Base(fw.filePath) {
				continue
			}

			// Check if the event is a write, create, or remove event
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				fw.logger.Info("Write event detected", zap.String("event", event.Name))
				fw.mutex.Lock()
				lastEvent, exists := fw.debounceMap[event.Name]
				now := time.Now()
				fw.debounceMap[event.Name] = now
				shouldHandle := !exists || now.Sub(lastEvent) >= debounceInterval
				fw.mutex.Unlock()

				// Handle the change if not debounced
				if shouldHandle {
					fw.logger.Info("Calling onChange callback", zap.String("event", event.Name))
					if err := fw.onChange(); err != nil {
						fw.logger.Error("Callback onChange failed", zap.String("event", event.Name), zap.Error(err))
					}
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Error("Error watching file", zap.String("file", fw.filePath), zap.Error(err))

		case <-fw.done:
			return
		}
	}
}

// Close stops watching the file and cleans up resources.
// It signals the watch goroutine to exit and closes the underlying watcher.
func (fw *FileWatcher) Close() error {
	fw.logger.Info("Closing file watcher")
	close(fw.done)
	return fw.watcher.Close()
}
