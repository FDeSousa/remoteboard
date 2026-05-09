// Package watcher watches the config file for changes using fsnotify and
// calls the provided callback whenever the file is written.
package server

import (
	"log"
	"time"

	"github.com/FDeSousa/remoteboard/internal/config"
	"github.com/fsnotify/fsnotify"
)

// WatchConfig starts a goroutine that watches cfgPath for writes. Whenever
// the file changes it is reloaded and onReload is called with the new config.
// The goroutine exits when the returned stop function is called.
func WatchConfig(cfgPath string, onReload func(*config.Config)) (stop func(), err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.Add(cfgPath); err != nil {
		w.Close()
		return nil, err
	}

	done := make(chan struct{})
	go func() {
		// debounce is nil (disabled) until a file-write event arms it.
		// A nil channel in a Go select case is skipped, so the reload only
		// fires 200 ms after the last write event — even if the editor emits
		// multiple rapid events for a single save. After firing we reset to
		// nil so subsequent loops do not re-trigger the reload.
		var debounce <-chan time.Time
		for {
			select {
			case <-done:
				w.Close()
				return
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// (Re-)arm the debounce timer.
					debounce = time.After(200 * time.Millisecond)
				}
			case <-debounce:
				// Disarm so this case is skipped until the next write event.
				debounce = nil
				cfg, err := config.Load(cfgPath)
				if err != nil {
					log.Printf("watcher: reload error: %v", err)
					continue
				}
				log.Printf("watcher: config reloaded from %s", cfgPath)
				onReload(cfg)
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Printf("watcher: %v", err)
			}
		}
	}()

	return func() { close(done) }, nil
}
