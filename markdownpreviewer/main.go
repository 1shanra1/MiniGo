package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// 1. Create the Broker and run its engine in the background.
	broker := NewBroker()
	go broker.Run()

	// 2. Create the Server, giving it a reference to the Broker.
	server := &Server{b: broker}
	// Start the server in a new goroutine so it doesn't block main.
	go func() {
		if err := server.Start(); err != nil {
			// Using Fatalf will exit the application if the server can't start.
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// 3. Set up the file watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 4. Start the watcher's event-listening goroutine.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// We only care about .md files that have been written to.
				if event.Has(fsnotify.Write) && strings.HasSuffix(event.Name, ".md") {
					log.Println("modified markdown file:", event.Name)
					// Send the path of the modified file to the Broker's channel.
					broker.fp <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()

	// 5. Set up the directory to watch.
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	mdFilesPath := filepath.Join(currentDir, "mdfiles")

	// Add the path to the watcher.
	err = watcher.Add(mdFilesPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Watching for .md file changes in: %s", mdFilesPath)

	// 6. Block the main goroutine indefinitely to keep the application running.
	<-make(chan struct{})
}
