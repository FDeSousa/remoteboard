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
		// Debounce: editors often emit multiple events for a single save.
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
					debounce = time.After(200 * time.Millisecond)
				}
			case <-debounce:
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
