package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	mdFilesPath := filepath.Join(currentDir, "mdfiles")

	err = watcher.Add(mdFilesPath)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{}) // blocks main goroutine indefinitely
}
