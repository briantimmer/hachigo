package server

import (
	"fmt"
	"hachigo/pkg/config"
	"hachigo/pkg/site"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Start spins up the HTTP preview server and optionally starts the filesystem watcher
func Start(cfg *config.Config, port string, watch bool) error {
	destDir := cfg.Destination

	// Run initial build on start
	fmt.Println("Running initial build...")
	if err := site.Build(cfg); err != nil {
		fmt.Printf("Initial build failed: %v\n", err)
	} else {
		fmt.Println("Initial build complete!")
	}

	if watch {
		go watchChanges(cfg)
	}

	addr := ":" + port
	fmt.Printf("Starting HTTP preview server at http://localhost:%s/\n", port)
	fmt.Printf("Serving static files from: %s\n", destDir)

	handler := http.FileServer(http.Dir(destDir))
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return server.ListenAndServe()
}

func watchChanges(cfg *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Walk and watch all subdirectories under the source directory
	err = filepath.Walk(cfg.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			// Skip hidden directories (like .git, .sass-cache)
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error configuring watcher directories: %v\n", err)
		return
	}

	fmt.Println("Watching source files for changes...")

	var rebuildTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Ignore temp/hidden files (e.g. from editors)
			if strings.HasPrefix(filepath.Base(event.Name), ".") {
				continue
			}

			// Rebuild on file writes, creations, deletions, or renames
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				// If a new directory is created, watch it as well
				if event.Has(fsnotify.Create) {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						watcher.Add(event.Name)
					}
				}

				if rebuildTimer != nil {
					rebuildTimer.Stop()
				}

				rebuildTimer = time.AfterFunc(debounceDuration, func() {
					fmt.Printf("\nChange detected (%s). Rebuilding site...\n", event.Name)
					if err := site.Build(cfg); err != nil {
						fmt.Printf("Rebuild failed: %v\n", err)
					} else {
						fmt.Println("Rebuild completed successfully!")
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}
