package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
	"github.com/khhini/open-knowledge-engine.git/pkg/store"
)

type Watcher struct {
	mu       sync.Mutex
	baseDir  string
	memStore *store.MemoryStore
	wathcher *fsnotify.Watcher
	timer    *time.Timer
}

func StartWatcher(baseDir string, memStore *store.MemoryStore) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		baseDir:  baseDir,
		memStore: memStore,
		wathcher: fsWatcher,
	}

	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fsWatcher.Add(path)
		}
		return nil
	})

	go w.eventLoop()
	log.Printf("[FS WATCHER] Monitoring '%s' for real-time Markdown edits...", baseDir)

	return w, nil
}

func (w *Watcher) eventLoop() {
	for {
		select {
		case event, ok := <-w.wathcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = w.wathcher.Add(event.Name)
					log.Printf("[FS WATCHER] Watching new directory: %s", event.Name)
				}
			}

			filename := filepath.Base(event.Name)
			if strings.HasSuffix(filename, ".md") && filename != "index.md" && filename != "log.md" {
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					w.debounceReload(event)
				}
			}

		case err, ok := <-w.wathcher.Errors:
			if !ok {
				return
			}
			log.Printf("[FS WATCHER ERROR] %v", err)
		}
	}
}

func (w *Watcher) debounceReload(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timer != nil {
		w.timer.Stop()
	}

	w.timer = time.AfterFunc(200*time.Millisecond, func() {
		log.Printf("[FS WATCHER] File event (%s on %s) -> Reloading In-Memory Graph...", event.Op, filepath.Base(event.Name))

		if err := w.memStore.LoadAll(); err != nil {
			log.Printf("[FS WATCHER ERROR] Failed to reload store: %v", err)
		}

		if err := okf.GenerateBundleIndex(w.memStore.List(), w.baseDir); err != nil {
			log.Printf("[FS WATCHER ERROR] Failed to update index.md: %v", err)
		} else {
			log.Panicf("[FS WATCHER] In-Memory Graph and index.md updated successfully!")
		}
	})
}

func (w *Watcher) Close() error {
	return w.wathcher.Close()
}
