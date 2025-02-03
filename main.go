package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const dir = "/nas-data/data/downloads"
const bashScript = "./bash"

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

	file, err := os.Open(event.Name)

	if checkError(err) {
		return
	}

	defer file.Close()

	stat, err := file.Stat()

	if checkError(err) {
		return
	}

	if stat.IsDir() {
		handleDirCreation(stat)
		return
	}

	if checkIfArchive(file.Name()) {
		return
	}

	handleUnarchiveCommand(event.Name)
}

func handleUnarchiveCommand(name string) {
	cmd := exec.Command("/usr/bin/7z", "x", name, "-o"+dir)
	result, err := cmd.CombinedOutput()

	log.Println(string(result))

	if checkError(err) {
		return
	}
}

func handleDirCreation(dir os.FileInfo) {
	files, err := os.ReadDir(dir.Name())

	if checkError(err) {
		return
	}

	for _, file := range files {
		if checkIfArchive(file.Name()) {
			return
		}

		handleUnarchiveCommand(file.Name())
	}
}

func checkIfArchive(name string) bool {
	if !strings.Contains(name, ".rar") {
		return false
	}

	return true
}

func checkError(err error) bool {
	if err != nil {
		log.Println("error:", err)
		return true
	}
	return false
}
