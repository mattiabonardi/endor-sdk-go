package sdk

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	baseDir     = os.ExpandEnv("$HOME/etc/endor/dsl")
	watcher     *fsnotify.Watcher
	watchMu     sync.Mutex
	watchedDirs = make(map[string]bool)
)

type Watcher struct{}

func (h *Watcher) Start() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create watcher:", err)
	}
	defer watcher.Close()

	// Initial scan and watch
	h.scanAndWatch()

	// Start watching for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			h.handleEvent(event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func (h *Watcher) scanAndWatch() {
	// Always watch root dsl folder
	h.addWatch(baseDir)

	// Walk through all subdirs and add watchers to:
	// 1. app folders (dsl/*)
	// 2. resources folders (dsl/*/resources)
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if base == "resources" || filepath.Dir(path) == baseDir {
				h.addWatch(path)
			}
		}
		return nil
	})
}

func (h *Watcher) handleEvent(event fsnotify.Event) {
	// Log resource YAML changes
	if strings.Contains(event.Name, "/resources/") && strings.HasSuffix(event.Name, ".yaml") {
		log.Printf("YAML changed: %s (%s)", event.Name, event.Op)
	}

	if event.Op&fsnotify.Create != 0 {
		fi, err := os.Stat(event.Name)
		if err == nil && fi.IsDir() {
			log.Println("New directory created:", event.Name)

			// Always watch the newly created directory
			h.addWatch(event.Name)

			// If a "resources" folder was just created, scan for YAML
			if filepath.Base(event.Name) == "resources" {
				h.scanAndWatch() // Ensure .yaml files inside are picked up
			}
		}
	}

	// Stop watching deleted folders
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		watchMu.Lock()
		if watchedDirs[event.Name] {
			log.Println("Removing watcher:", event.Name)
			watcher.Remove(event.Name)
			delete(watchedDirs, event.Name)
		}
		watchMu.Unlock()
	}
}

// Thread-safe watch registration
func (h *Watcher) addWatch(dir string) {
	watchMu.Lock()
	defer watchMu.Unlock()

	if watchedDirs[dir] {
		return
	}

	err := watcher.Add(dir)
	if err != nil {
		log.Printf("Failed to watch %s: %v", dir, err)
		return
	}

	watchedDirs[dir] = true
	log.Println("Watching:", dir)
}
