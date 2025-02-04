package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const rootDir = "/nas-data/data/downloads"

func main() {
	// lets first make sure we dont have existing archives that need extracting
	handleAll()

	log.Println("Starting Watch instance for " + rootDir)
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

	err = watcher.Add(rootDir)
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
		handleDir(stat.Name())
		return
	}

	if !checkIfArchive(file.Name()) {
		return
	}

	handleUnarchiveCommand(event.Name, rootDir)
}

func handleAll() {
	dirs, err := os.ReadDir(rootDir)

	if checkError(err) {
		return
	}

	for _, fdir := range dirs {
		if fdir.IsDir() {

			if checkError(err) {
				continue
			}

			handleDir(rootDir + "/" + fdir.Name())
		}
	}
}

func handleDir(dir string) {
	files, err := os.ReadDir(dir)

	if checkError(err) {
		return
	}

	for _, file := range files {
		if !checkIfArchive(file.Name()) {
			continue
		}

		handleUnarchiveCommand(filepath.Join(dir, file.Name()), dir)
	}
}

func handleUnarchiveCommand(name string, outDir string) {
	cmd := exec.Command("/usr/bin/7z", "x", name, "-o"+outDir, "-aos")
	result, err := cmd.CombinedOutput()

	log.Println(string(result))

	if checkError(err) {
		return
	}
}

func checkIfArchive(name string) bool {
	if strings.Contains(name, ".rar") {
		return true
	}

	return false
}

func checkError(err error) bool {
	if err != nil {
		log.Println("error:", err)
		return true
	}
	return false
}
