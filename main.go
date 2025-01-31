package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os/exec"
	"strings"
)

const dir = "/nas-data/data/downloads"
const bashScript = "~/unarchiver/bash/unarchive"

func main() {
	log.Println("Starting Watch instance for " + dir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(watcher)

	go listenForEvents(watcher)

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func listenForEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleEvent(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func handleEvent(event fsnotify.Event) {
	log.Println("Event: " + event.Name)
	if !event.Has(fsnotify.Create) {
		return
	}

	log.Println("Created event captured: " + event.Name)

	if !strings.Contains(event.Name, "7z") {
		return
	}

	err := exec.Command(bashScript, event.Name).Run()

	if err != nil {
		log.Println("error:", err)
		return
	}
}
