// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

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
	logger      *zap.Logger          // Logger for the file watcher
	suspended   bool                 // Flag to indicate if the watcher is suspended
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
		suspended:   false,
	}

	// Watch both the file and its parent directory
	if err = watcher.Add(filePath); err != nil {
		watcher.Close()
		return nil, err
	}

	return fw, nil
}

func (fw *FileWatcher) WithLogger(l *zap.Logger) *FileWatcher {
	fw.logger = l
	return fw
}

func (fw *FileWatcher) Watch() {
	go fw.watch()
}

func (fw *FileWatcher) Disable() {
	fw.suspended = true
	fw.logger.Debug("Disabled file watcher")
}

func (fw *FileWatcher) Enable() {
	fw.suspended = false
	fw.logger.Debug("Enabled file watcher")
}

// watch monitors the file for changes and triggers the callback when appropriate.
// It implements debouncing to prevent multiple rapid callbacks for the same file change.
// The method runs in a goroutine and continues until the watcher is closed.
func (fw *FileWatcher) watch() {
	const debounceInterval = 100 * time.Millisecond

	fw.logger.Info("Starting file watcher",
		zap.String("file", fw.filePath),
		zap.String("dir", filepath.Dir(fw.filePath)),
		zap.Duration("debounce", debounceInterval))

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// If the watcher is suspended, skip processing events
			if fw.suspended {
				fw.logger.Debug("Watcher is suspended, ignoring event",
					zap.String("event", event.Op.String()),
					zap.String("file", event.Name))
				continue
			}

			// Normalize paths for comparison
			eventPath := filepath.Clean(event.Name)
			watchPath := filepath.Clean(fw.filePath)

			// Check if the event is related to our file
			if eventPath != watchPath {
				fw.logger.Debug("Ignoring event for different file",
					zap.String("event_path", eventPath),
					zap.String("watch_path", watchPath))
				continue
			}

			// Log all events for debugging
			fw.logger.Info("File event detected",
				zap.String("operation", event.Op.String()),
				zap.String("file", event.Name))

			// Check if the event is a write, create, or remove event
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				fw.mutex.Lock()
				lastEvent, exists := fw.debounceMap[event.Name]
				now := time.Now()

				// Always trigger callback for Remove and Create events
				shouldHandle := event.Op&(fsnotify.Create|fsnotify.Remove) != 0 ||
					!exists || now.Sub(lastEvent) >= debounceInterval

				fw.debounceMap[event.Name] = now
				fw.mutex.Unlock()

				// Handle the change if not debounced or if it's a critical event
				if shouldHandle {
					fw.logger.Info("Triggering onChange callback",
						zap.String("operation", event.Op.String()),
						zap.String("file", event.Name))

					if err := fw.onChange(); err != nil {
						fw.logger.Error("Callback onChange failed",
							zap.String("operation", event.Op.String()),
							zap.String("file", event.Name),
							zap.Error(err))
					}
				} else {
					fw.logger.Debug("Debouncing event",
						zap.String("operation", event.Op.String()),
						zap.String("file", event.Name))
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Error("Watcher error", zap.Error(err))
		case <-fw.done:
			return
		}
	}
}

// Close stops the file watcher and releases associated resources.
func (fw *FileWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}
