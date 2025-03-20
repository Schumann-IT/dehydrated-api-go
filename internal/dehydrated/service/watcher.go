package service

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches for changes to a file
type FileWatcher struct {
	filePath    string
	watcher     *fsnotify.Watcher
	onChange    func() error
	mutex       sync.Mutex
	debounceMap map[string]time.Time
	done        chan struct{}
}

// NewFileWatcher creates a new FileWatcher instance
func NewFileWatcher(filePath string, onChange func() error) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	fw := &FileWatcher{
		filePath:    filePath,
		watcher:     watcher,
		onChange:    onChange,
		debounceMap: make(map[string]time.Time),
		done:        make(chan struct{}),
	}

	// Watch both the file and its parent directory
	if err := watcher.Add(filePath); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch file %s: %w", filePath, err)
	}

	if err := watcher.Add(filepath.Dir(filePath)); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch directory %s: %w", filepath.Dir(filePath), err)
	}

	go fw.watch()

	return fw, nil
}

// watch monitors the file for changes
func (fw *FileWatcher) watch() {
	const debounceInterval = 100 * time.Millisecond

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
				fw.mutex.Lock()
				lastEvent, exists := fw.debounceMap[event.Name]
				now := time.Now()
				fw.debounceMap[event.Name] = now
				fw.mutex.Unlock()

				// Debounce events that occur too quickly
				if exists && now.Sub(lastEvent) < debounceInterval {
					continue
				}

				// Handle the change
				if err := fw.onChange(); err != nil {
					log.Printf("Error handling file change: %v", err)
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching file: %v", err)

		case <-fw.done:
			return
		}
	}
}

// Close stops watching the file and cleans up resources
func (fw *FileWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}
